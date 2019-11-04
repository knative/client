// Copyright © 2019 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package revision

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
	clientserving "knative.dev/client/pkg/serving"
	servingserving "knative.dev/serving/pkg/apis/serving"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// Matching image digest
var imageDigestRegexp = regexp.MustCompile(`(?i)sha256:([0-9a-f]{64})`)

func NewRevisionDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe revisions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires the revision name.")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}

			revision, err := client.GetRevision(args[0])
			if err != nil {
				return err
			}

			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(revision, cmd.OutOrStdout())
			}
			printDetails, err := cmd.Flags().GetBool("verbose")
			var service *v1alpha1.Service
			serviceName, ok := revision.Labels[servingserving.ServiceLabelKey]
			if printDetails && ok {
				service, err = client.GetService(serviceName)
				if err != nil {
					return err
				}
			}
			// Do the human-readable printing thing.
			return describe(cmd.OutOrStdout(), revision, service, printDetails)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	machineReadablePrintFlags.AddFlags(command)
	flags.BoolP("verbose", "v", false, "More output.")
	return command
}

func describe(w io.Writer, revision *v1alpha1.Revision, service *v1alpha1.Service, printDetails bool) error {
	dw := printers.NewPrefixWriter(w)
	commands.WriteMetadata(dw, &revision.ObjectMeta, printDetails)
	WriteImage(dw, revision)
	WritePort(dw, revision)
	WriteEnv(dw, revision, printDetails)
	WriteScale(dw, revision)
	WriteConcurrencyOptions(dw, revision)
	WriteResources(dw, revision)
	serviceName, ok := revision.Labels[servingserving.ServiceLabelKey]
	if ok {
		serviceSection := dw.WriteAttribute("Service", serviceName)
		if printDetails {
			serviceSection.WriteAttribute("Config Gen", revision.Labels[servingserving.ConfigurationGenerationLabelKey])
			serviceSection.WriteAttribute("Latest Created", strconv.FormatBool(revision.Name == service.Status.LatestCreatedRevisionName))
			serviceSection.WriteAttribute("Latest Ready", strconv.FormatBool(revision.Name == service.Status.LatestReadyRevisionName))
			percent, tags := trafficForRevision(revision.Name, service)
			if percent != 0 {
				serviceSection.WriteAttribute("Traffic", strconv.FormatInt(int64(percent), 10)+"%")
			}
			if len(tags) > 0 {
				commands.WriteSliceDesc(serviceSection, tags, "Tags", printDetails)
			}
		}

	}
	dw.WriteLine()
	commands.WriteConditions(dw, revision.Status.Conditions, printDetails)
	if err := dw.Flush(); err != nil {
		return err
	}
	return nil
}

func WriteConcurrencyOptions(dw printers.PrefixWriter, revision *v1alpha1.Revision) {
	target := clientserving.ConcurrencyTarget(&revision.ObjectMeta)
	limit := revision.Spec.ContainerConcurrency
	if target != nil || limit != 0 {
		section := dw.WriteAttribute("Concurrency", "")
		if limit != 0 {
			section.WriteAttribute("Limit", strconv.FormatInt(int64(limit), 10))
		}
		if target != nil {
			section.WriteAttribute("Target", strconv.Itoa(*target))
		}
	}
}

// Write the image attribute (with
func WriteImage(dw printers.PrefixWriter, revision *v1alpha1.Revision) {
	c, err := clientserving.ContainerOfRevisionSpec(&revision.Spec)
	if err != nil {
		dw.WriteAttribute("Image", "Unknown")
		return
	}
	image := c.Image
	// Check if the user image is likely a more user-friendly description
	pinnedDesc := "at"
	userImage := clientserving.UserImage(&revision.ObjectMeta)
	imageDigest := revision.Status.ImageDigest
	if userImage != "" && imageDigest != "" {
		var parts []string
		if strings.Contains(image, "@") {
			parts = strings.Split(image, "@")
		} else {
			parts = strings.Split(image, ":")
		}
		// Check if the user image refers to the same thing.
		if strings.HasPrefix(userImage, parts[0]) {
			pinnedDesc = "pinned to"
			image = userImage
		}
	}
	if imageDigest != "" {
		image = fmt.Sprintf("%s (%s %s)", image, pinnedDesc, shortenDigest(imageDigest))
	}
	dw.WriteAttribute("Image", image)
}

func WritePort(dw printers.PrefixWriter, revision *v1alpha1.Revision) {
	port := clientserving.Port(&revision.Spec)
	if port != nil {
		dw.WriteAttribute("Port", strconv.FormatInt(int64(*port), 10))
	}
}

func WriteEnv(dw printers.PrefixWriter, revision *v1alpha1.Revision, printDetails bool) {
	env := stringifyEnv(revision)
	if env != nil {
		commands.WriteSliceDesc(dw, env, "Env", printDetails)
	}
}

func WriteScale(dw printers.PrefixWriter, revision *v1alpha1.Revision) {
	// Scale spec if given
	scale, _ := clientserving.ScalingInfo(&revision.ObjectMeta, &revision.Spec)
	if scale != nil && (scale.Max != nil || scale.Min != nil) {
		dw.WriteAttribute("Scale", formatScale(scale.Min, scale.Max))
	}
}

func WriteResources(dw printers.PrefixWriter, r *v1alpha1.Revision) {
	c, err := clientserving.ContainerOfRevisionSpec(&r.Spec)
	if err != nil {
		return
	}
	requests := c.Resources.Requests
	limits := c.Resources.Limits
	writeResourcesHelper(dw, "Memory", requests.Memory(), limits.Memory())
	writeResourcesHelper(dw, "CPU", requests.Cpu(), limits.Cpu())
}

// Write request ... limits or only one of them
func writeResourcesHelper(dw printers.PrefixWriter, label string, request *resource.Quantity, limit *resource.Quantity) {
	value := ""
	if !request.IsZero() && !limit.IsZero() {
		value = request.String() + " ... " + limit.String()
	} else if !request.IsZero() {
		value = request.String()
	} else if !limit.IsZero() {
		value = limit.String()
	}

	if value == "" {
		return
	}

	dw.WriteAttribute(label, value)
}

// Extract pure sha sum and shorten to 8 digits,
// as the digest should to be user consumable. Use the resource via `kn service get`
// to get to the full sha
func shortenDigest(digest string) string {
	match := imageDigestRegexp.FindStringSubmatch(digest)
	if len(match) > 1 {
		return string(match[1][:6])
	}
	return digest
}

// Format scale in the format "min ... max" with max = ∞ if not set
func formatScale(minScale *int, maxScale *int) string {
	ret := "0"
	if minScale != nil {
		ret = strconv.Itoa(*minScale)
	}

	ret += " ... "

	if maxScale != nil {
		ret += strconv.Itoa(*maxScale)
	} else {
		ret += "∞"
	}
	return ret
}

func stringifyEnv(revision *v1alpha1.Revision) []string {
	container, err := clientserving.ContainerOfRevisionSpec(&revision.Spec)
	if err != nil {
		return nil
	}

	envVars := make([]string, 0, len(container.Env))
	for _, env := range container.Env {
		var value string
		if env.ValueFrom != nil {
			value = "[ref]"
		} else {
			value = env.Value
		}
		envVars = append(envVars, fmt.Sprintf("%s=%s", env.Name, value))
	}
	return envVars
}

func trafficForRevision(name string, service *v1alpha1.Service) (int, []string) {
	if len(service.Status.Traffic) == 0 {
		return 0, nil
	}
	percent := 0
	tags := []string{}
	for _, target := range service.Status.Traffic {
		if target.RevisionName == name {
			percent += target.Percent
			if target.Tag != "" {
				tags = append(tags, target.Tag)
			}
		}
	}
	return percent, tags
}
