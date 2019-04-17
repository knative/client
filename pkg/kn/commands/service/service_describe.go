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

package service

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/knative/client/pkg/printers"
	"github.com/knative/client/pkg/serving"
	"github.com/knative/pkg/apis"
	"github.com/knative/pkg/apis/duck/v1beta1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// Command for printing out a description of a service, meant to be consumed by humans
// It will show information about the serivce itself, but also a summary
// about the associated revisions.
// "kn service describe" knows three modes, which can be combined:
//
// `--all`    : Print more information. By default only are shorter summary is shown
// `--color`  : Use a colorful output (but only when on a tty)
// `--hipster`: Experimental output mode for a very compact representation using emojis

// Whether to print extended information
var printAll bool

// Max length When to truncate long strings (when not "all" mode switched on)
var truncateAt = 100

// View object for collecting revision related information in the context
// of a Service
type revisionDesc struct {
	name                    string
	configuration           string
	configurationGeneration int
	creationTimestamp       time.Time

	percent int
	latest  *bool

	logUrl         string
	timeoutSeconds *int64

	image       string
	imageDigest string
	env         []string
	port        *int32
}

// [REMOVE COMMENT WHEN MOVING TO 0.7.0]
// For transition to v1beta1 this command uses the migration approach as described
// in https://docs.google.com/presentation/d/1mOhnhy8kA4-K9Necct-NeIwysxze_FUule-8u5ZHmwA/edit#slide=id.p
// With serving 0.6.0 we are at step #1
// I.e we first look at new fields of the v1alpha1 API before falling back to the original ones.
// As this command does not any updates, it's just a matter of fallbacks.
// [/REMOVE COMMENT WHEN MOVING TO 0.7.0]

// Return a new command for describing a service.
func NewServiceDescribeCommand(p *commands.KnParams) *cobra.Command {
	command := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details for a given service",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("no service name provided")
			}
			if len(args) > 1 {
				return errors.New("more than one service name provided")
			}
			serviceName := args[0]

			namespace, err := GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}

			printAll, err = cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}

			useColor, err := cmd.Flags().GetBool("color")
			if err != nil {
				return err
			}
			// Set color option globally
			color.NoColor = !useColor

			service, err := client.Service(serviceName)
			if err != nil {
				return err
			}
			// Additional revision related information
			revisionDescs, err := getRevisionDescriptions(sClient, service, printAll)

			hipsterMode, err := cmd.Flags().GetBool("hipster")
			if err != nil {
				return err
			}
			if hipsterMode {
				// Compact mode is completely separated as it used different way
				// for formatting
				return describeCompact(cmd.OutOrStdout(), service, revisionDescs)
			}
			return describe(cmd.OutOrStdout(), service, revisionDescs)
		},
	}
	flags := command.Flags()
	AddNamespaceFlags(flags, false)
	flags.BoolP("all", "a", false, "don't truncate long information")
	flags.BoolP("color", "c", false, "use colorful output")
	flags.Bool("hipster", false, "🤓")
	return command
}

// Main action describing the service
func describe(w io.Writer, service *v1alpha1.Service, revisions []*revisionDesc) error {
	dw := printers.NewPrefixWriter(w)

	// Service info
	writeService(dw, service)
	dw.WriteLine()
	if err := dw.Flush(); err != nil {
		return err
	}

	// Revisions summary info
	writeRevisions(dw, revisions)
	dw.WriteLine()
	if err := dw.Flush(); err != nil {
		return err
	}

	// Condition info
	writeConditions(dw, service)
	if err := dw.Flush(); err != nil {
		return err
	}

	return nil
}

// Write out main service information. Use colors for major items.
func writeService(dw printers.PrefixWriter, service *v1alpha1.Service) {
	dw.WriteColsLn(printers.LEVEL_0, l("Name"), wc(service.Name, color.FgYellow))
	dw.WriteColsLn(printers.LEVEL_0, l("Namespace"), service.Namespace)
	dw.WriteColsLn(printers.LEVEL_0, l("URL"), wc(extractURL(service), color.FgGreen))
	if service.Status.Address != nil {
		dw.WriteColsLn(printers.LEVEL_0, l("Address"), service.Status.Address.Hostname)
	}
	writeMapDesc(dw, printers.LEVEL_0, service.Labels, l("Labels"), "")
	writeMapDesc(dw, printers.LEVEL_0, service.Annotations, l("Annotations"), "")
	dw.WriteColsLn(printers.LEVEL_0, l("Age"), age(service.CreationTimestamp.Time))
}

// Write out revisions associated with this service. By default only active
// target revisions are printed, but with --all also inactive revisions
// created by this services are shown
func writeRevisions(dw printers.PrefixWriter, revisions []*revisionDesc) {
	dw.WriteColsLn(printers.LEVEL_0, l("Revisions"))
	for _, revisionDesc := range revisions {
		dw.WriteColsLn(printers.LEVEL_1, formatPercentage(revisionDesc.percent), l("Name"), getRevisionNameWithGenerationAndAge(revisionDesc))
		dw.WriteColsLn(printers.LEVEL_1, "", l("Image"), getImageDesc(revisionDesc))
		if revisionDesc.port != nil {
			dw.WriteColsLn(printers.LEVEL_1, "", l("Port"), strconv.FormatInt(int64(*revisionDesc.port), 10))
		}
		writeSliceDesc(dw, printers.LEVEL_1, revisionDesc.env, l("Env"), "\t")
	}
}

// Print out a table with conditions. Use green for 'ok', and red for 'nok' if color is enabled
func writeConditions(dw printers.PrefixWriter, service *v1alpha1.Service) {
	dw.WriteColsLn(printers.LEVEL_0, l("Conditions"))
	maxLen := getMaxTypeLen(service.Status.Conditions)
	formatHeader := "%-2s %-" + strconv.Itoa(maxLen) + "s %6s %-s\n"
	formatRow := "%-" + strconv.Itoa(2+colorOffset()) + "s %-" + strconv.Itoa(maxLen+colorOffset()) + "s %6s %-s\n"
	dw.Write(printers.LEVEL_1, formatHeader, "OK", "TYPE", "AGE", "REASON")
	for _, condition := range service.Status.Conditions {
		ok := formatStatus(condition.Status)
		reason := condition.Reason
		if printAll && reason != "" {
			reason = fmt.Sprintf("%s (%s)", reason, condition.Message)
		}
		dw.Write(printers.LEVEL_1, formatRow, ok, formatConditionType(condition), age(condition.LastTransitionTime.Inner.Time), reason)
	}
}

// ======================================================================================
// Helper functions

// Format label depending whether color mode is on or not
func l(label string) string {
	return label + ":"
}

// Get a colored string if color is enabled
func wc(value string, attributes ...color.Attribute) string {
	return color.New(attributes...).Sprintf("%s", value)
}

// How many invisible characters are used when coloring is switched on
func colorOffset() int {
	return len(color.RedString(""))
}

// ======================================================================================

// Format the revision name along with its generation. Use colors if enabled.
func getRevisionNameWithGenerationAndAge(desc *revisionDesc) string {
	return wc(desc.name, color.FgYellow) + " " +
		wc("["+strconv.Itoa(desc.configurationGeneration)+"]", color.FgHiBlack) +
		" " +
		wc("("+age(desc.creationTimestamp)+")", color.FgHiBlack)
}

// Used for conditions table to do own formatting for the table,
// as the tabbed writer doesn't work nicely with colors
func getMaxTypeLen(conditions v1beta1.Conditions) int {
	max := 0
	for _, condition := range conditions {
		if len(condition.Type) > max {
			max = len(condition.Type)
		}
	}
	return max
}

// Color the type of the conditions
func formatConditionType(condition apis.Condition) string {
	switch condition.Status {
	case v1.ConditionTrue:
		return color.HiGreenString(string(condition.Type))
	case v1.ConditionFalse:
		return color.HiRedString(string(condition.Type))
	default:
		return string(condition.Type)
	}
}

// Status in ASCII format
func formatStatus(status corev1.ConditionStatus) string {
	switch status {
	case v1.ConditionTrue:
		return color.HiGreenString("++")
	case v1.ConditionFalse:
		return color.HiRedString("--")
	default:
		return ""
	}
}

// Return either image name with tag or together with its resolved digest
func getImageDesc(desc *revisionDesc) string {
	image := color.CyanString(desc.image)
	if printAll && desc.imageDigest != "" {
		digest := color.HiBlackString("(" + shortenDigest(desc.imageDigest) + ")")
		return fmt.Sprintf("%s %s", image, digest)
	}
	return image
}

// Extract pure sha sum and shorten to 8 digits,
// as the digest should to be user consumable. Use the resource via `kn service get`
// to get to the full sha
func shortenDigest(digest string) string {
	digestRegexp := regexp.MustCompile(`(?i)sha256:([0-9a-f]+)`)
	match := digestRegexp.FindStringSubmatch(digest)
	if len(match) > 1 {
		return string(match[1][:12])
	}
	return digest
}

// Write a map either compact in a single line (possibly truncated) or, if printAll is set,
// over multiple line, one line per key-value pair
func writeMapDesc(dw printers.PrefixWriter, indent int, m map[string]string, label string, labelPrefix string) {
	if len(m) == 0 {
		return
	}

	if printAll {
		l := labelPrefix + label
		for key, value := range m {
			dw.WriteColsLn(indent, l, key+"="+value)
			l = labelPrefix
		}
		return
	}

	dw.WriteColsLn(indent, label, joinAndTruncate(m))
}

// Writer a slice compact (printAll == false) in one line, or over multiple line
// with key-value line-by-line (printAll == true)
func writeSliceDesc(dw printers.PrefixWriter, indent int, s []string, label string, labelPrefix string) {

	if len(s) == 0 {
		return
	}

	if printAll {
		l := labelPrefix + label
		for _, value := range s {
			dw.WriteColsLn(indent, l, value)
			l = labelPrefix
		}
		return
	}

	joined := strings.Join(s, ", ")
	if len(joined) > truncateAt {
		joined = joined[:truncateAt-4] + " ..."
	}
	dw.WriteColsLn(indent, labelPrefix+label, joined)
}

// Join to key=value pair, comma separated, and truncate if longer than a limit
func joinAndTruncate(m map[string]string) string {
	ret := ""
	for key, value := range m {
		ret += fmt.Sprintf("%s=%s, ", key, value)
		if len(ret) > truncateAt {
			break
		}
	}
	// cut of two latest chars
	ret = strings.TrimRight(ret, ", ")
	if len(ret) <= truncateAt {
		return ret
	}
	return string(ret[:truncateAt-4]) + " ..."
}

// Format target percentage that it fits in the revision table
func formatPercentage(percentage int) string {
	if percentage == 0 {
		return "   -"
	}
	return fmt.Sprintf("%-3d%%", percentage)
}

func age(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return duration.ShortHumanDuration(time.Now().Sub(t))
}

// Call the backend to query revisions for the given service and build up
// the view objects used for output
func getRevisionDescriptions(client *serving.NamespacedClient, service *v1alpha1.Service, all bool) ([]*revisionDesc, error) {
	revisionDescs := make(map[string]*revisionDesc)

	trafficTargets := service.Status.Traffic

	for _, target := range trafficTargets {
		revision, err := extractRevisionFromTarget(client, target)
		if err != nil {
			return nil, fmt.Errorf("cannot extract revision from service %s: %v", service.Name, err)
		}
		revisionDescs[revision.Name], err = newRevisionDesc(revision, &target)
		if err != nil {
			return nil, err
		}
	}
	if all {
		if err := completeWithUntargetedRevisions(client, service, revisionDescs); err != nil {
			return nil, err
		}
	}
	return orderByConfigurationGeneration(revisionDescs), nil
}

// Order the list of revisions so that the newest revisions are at the top
func orderByConfigurationGeneration(descs map[string]*revisionDesc) []*revisionDesc {
	descsList := make([]*revisionDesc, len(descs))
	idx := 0
	for _, desc := range descs {
		descsList[idx] = desc
		idx++
	}
	sort.SliceStable(descsList, func(i, j int) bool {
		return descsList[i].configurationGeneration > descsList[j].configurationGeneration
	})
	return descsList
}

func completeWithUntargetedRevisions(client *serving.NamespacedClient, service *v1alpha1.Service, descs map[string]*revisionDesc) error {
	revisions, err := client.RevisionsForService(service)
	if err != nil {
		return err
	}
	for _, revision := range revisions {
		if _, ok := descs[revision.Name]; !ok {
			descs[revision.Name], err = newRevisionDesc(revision, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newRevisionDesc(revision *v1alpha1.Revision, target *v1alpha1.TrafficTarget) (*revisionDesc, error) {
	container := extractContainer(revision)
	generation, err := strconv.ParseInt(revision.Labels["serving.knative.dev/configurationGeneration"], 0, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot extract configuration generation for revision %s: %v", revision.Name, err)
	}
	revisionDesc := revisionDesc{
		name:              revision.Name,
		logUrl:            revision.Status.LogURL,
		timeoutSeconds:    revision.Spec.TimeoutSeconds,
		imageDigest:       revision.Status.ImageDigest,
		creationTimestamp: revision.CreationTimestamp.Time,

		configurationGeneration: int(generation),
		configuration:           revision.Labels["serving.knative.dev/configuration"],
	}

	addContainerInfo(&revisionDesc, container)
	addTargetInfo(&revisionDesc, target)
	return &revisionDesc, nil
}

func addTargetInfo(desc *revisionDesc, target *v1alpha1.TrafficTarget) {
	if target != nil {
		desc.percent = target.Percent
		desc.latest = target.LatestRevision
	}
}

func addContainerInfo(desc *revisionDesc, container *v1.Container) {
	addImage(desc, container)
	addEnv(desc, container)
	addPort(desc, container)
}

func addEnv(desc *revisionDesc, container *v1.Container) {
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
	desc.env = envVars
}

func addPort(desc *revisionDesc, container *v1.Container) {
	if len(container.Ports) > 0 {
		port := container.Ports[0].ContainerPort
		desc.port = &port
	}
}

func addImage(desc *revisionDesc, container *v1.Container) {
	desc.image = container.Image
}

func extractContainer(revision *v1alpha1.Revision) *v1.Container {
	if revision.Spec.Containers != nil && len(revision.Spec.Containers) > 0 {
		return &revision.Spec.Containers[0]
	}
	return revision.Spec.DeprecatedContainer
}

func extractRevisionFromTarget(client *serving.NamespacedClient, target v1alpha1.TrafficTarget) (*v1alpha1.Revision, error) {
	var revisionName = target.RevisionName
	if revisionName == "" {
		configurationName := target.ConfigurationName
		if configurationName == "" {
			return nil, fmt.Errorf("neither RevisionName nor ConfigurationName set")
		}
		configuration, err := client.Configuration(configurationName)
		if err != nil {
			return nil, err
		}
		revisionName = configuration.Status.LatestCreatedRevisionName
	}
	return client.Revision(revisionName)
}

func extractURL(service *v1alpha1.Service) string {
	status := service.Status
	if status.URL != nil {
		return status.URL.String()
	}
	return status.DeprecatedDomain
}
