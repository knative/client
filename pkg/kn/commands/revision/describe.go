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
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
	clientserving "knative.dev/client/pkg/serving"
)

// Matching image digest
var imageDigestRegexp = regexp.MustCompile(`(?i)sha256:([0-9a-f]{64})`)

func NewRevisionDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a revision",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn revision describe' requires name of the revision as single argument")
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewServingClient(namespace)
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
			if err != nil {
				return err
			}
			var service *servingv1.Service
			serviceName, ok := revision.Labels[serving.ServiceLabelKey]
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

func describe(w io.Writer, revision *servingv1.Revision, service *servingv1.Service, printDetails bool) error {
	dw := printers.NewPrefixWriter(w)
	commands.WriteMetadata(dw, &revision.ObjectMeta, printDetails)
	WriteImage(dw, revision)
	WritePort(dw, revision)
	WriteEnv(dw, revision, printDetails)
	WriteEnvFrom(dw, revision, printDetails)
	WriteScale(dw, revision)
	WriteConcurrencyOptions(dw, revision)
	WriteResources(dw, revision)
	serviceName, ok := revision.Labels[serving.ServiceLabelKey]
	if ok {
		serviceSection := dw.WriteAttribute("Service", serviceName)
		if printDetails {
			serviceSection.WriteAttribute("Configuration Generation", revision.Labels[serving.ConfigurationGenerationLabelKey])
			serviceSection.WriteAttribute("Latest Created", strconv.FormatBool(revision.Name == service.Status.LatestCreatedRevisionName))
			serviceSection.WriteAttribute("Latest Ready", strconv.FormatBool(revision.Name == service.Status.LatestReadyRevisionName))
			percent, tags := trafficAndTagsForRevision(revision.Name, service)
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

func WriteConcurrencyOptions(dw printers.PrefixWriter, revision *servingv1.Revision) {
	target := clientserving.ConcurrencyTarget(&revision.ObjectMeta)
	limit := revision.Spec.ContainerConcurrency
	autoscaleWindow := clientserving.AutoscaleWindow(&revision.ObjectMeta)
	if target != nil || limit != nil && *limit != 0 || autoscaleWindow != "" {
		section := dw.WriteAttribute("Concurrency", "")
		if limit != nil && *limit != 0 {
			section.WriteAttribute("Limit", strconv.FormatInt(int64(*limit), 10))
		}
		if target != nil {
			section.WriteAttribute("Target", strconv.Itoa(*target))
		}
		if autoscaleWindow != "" {
			section.WriteAttribute("Window", autoscaleWindow)
		}
	}

}

// Write the image attribute (with
func WriteImage(dw printers.PrefixWriter, revision *servingv1.Revision) {
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

func WritePort(dw printers.PrefixWriter, revision *servingv1.Revision) {
	port := clientserving.Port(&revision.Spec)
	if port != nil {
		dw.WriteAttribute("Port", strconv.FormatInt(int64(*port), 10))
	}
}

func WriteEnv(dw printers.PrefixWriter, revision *servingv1.Revision, printDetails bool) {
	env := stringifyEnv(revision)
	if env != nil {
		commands.WriteSliceDesc(dw, env, "Env", printDetails)
	}
}

func WriteEnvFrom(dw printers.PrefixWriter, revision *servingv1.Revision, printDetails bool) {
	envFrom := stringifyEnvFrom(revision)
	if envFrom != nil {
		commands.WriteSliceDesc(dw, envFrom, "EnvFrom", printDetails)
	}
}

func WriteScale(dw printers.PrefixWriter, revision *servingv1.Revision) {
	// Scale spec if given
	scale, err := clientserving.ScalingInfo(&revision.ObjectMeta)
	if err != nil {
		dw.WriteAttribute("Scale", fmt.Sprintf("Misformatted: %v", err))
	}
	if scale != nil && (scale.Max != nil || scale.Min != nil) {
		dw.WriteAttribute("Scale", formatScale(scale.Min, scale.Max))
	}
}

func WriteResources(dw printers.PrefixWriter, r *servingv1.Revision) {
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

func stringifyEnv(revision *servingv1.Revision) []string {
	container, err := clientserving.ContainerOfRevisionSpec(&revision.Spec)
	if err != nil {
		return nil
	}

	envVars := make([]string, 0, len(container.Env))
	for _, env := range container.Env {
		value := env.Value
		if env.ValueFrom != nil {
			value = "[ref]"
		}
		envVars = append(envVars, fmt.Sprintf("%s=%s", env.Name, value))
	}
	return envVars
}

func stringifyEnvFrom(revision *servingv1.Revision) []string {
	container, err := clientserving.ContainerOfRevisionSpec(&revision.Spec)
	if err != nil {
		return nil
	}

	var result []string
	for _, envFromSource := range container.EnvFrom {
		if envFromSource.ConfigMapRef != nil {
			result = append(result, "cm:"+envFromSource.ConfigMapRef.Name)
		}
		if envFromSource.SecretRef != nil {
			result = append(result, "secret:"+envFromSource.SecretRef.Name)
		}
	}
	return result
}
