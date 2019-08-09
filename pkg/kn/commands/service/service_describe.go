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

	"github.com/knative/serving/pkg/apis/autoscaling"
	"github.com/knative/serving/pkg/apis/serving"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/knative/client/pkg/printers"
	serving_kn_v1alpha1 "github.com/knative/client/pkg/serving/v1alpha1"

	"github.com/knative/pkg/apis"
	"github.com/knative/pkg/apis/duck/v1beta1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/knative/client/pkg/kn/commands"
)

// Command for printing out a description of a service, meant to be consumed by humans
// It will show information about the service itself, but also a summary
// about the associated revisions.

// Whether to print extended information
var printDetails bool

// Max length When to truncate long strings (when not "all" mode switched on)
const truncateAt = 100

// Matching image digest
var imageDigestRegexp = regexp.MustCompile(`(?i)sha256:([0-9a-f]{64})`)

// View object for collecting revision related information in the context
// of a Service. These are plain data types which can be directly used
// for printing out
type revisionDesc struct {
	name                    string
	configuration           string
	configurationGeneration int
	creationTimestamp       time.Time

	percent int
	latest  *bool

	logURL         string
	timeoutSeconds *int64

	image       string
	imageDigest string
	env         []string
	port        *int32

	// concurrency options
	maxScale          *int
	minScale          *int
	concurrencyTarget *int
	concurrencyLimit  *int64

	// resource options
	requestsMemory string
	requestsCPU    string
	limitsMemory   string
	limitsCPU      string
}

// [REMOVE COMMENT WHEN MOVING TO 0.7.0]
// For transition to v1beta1 this command uses the migration approach as described
// in https://docs.google.com/presentation/d/1mOhnhy8kA4-K9Necct-NeIwysxze_FUule-8u5ZHmwA/edit#slide=id.p
// With serving 0.6.0 we are at step #1
// I.e we first look at new fields of the v1alpha1 API before falling back to the original ones.
// As this command does not do any writes/updates, it's just a matter of fallbacks.
// [/REMOVE COMMENT WHEN MOVING TO 0.7.0]

// NewServiceDescribeCommand returns a new command for describing a service.
func NewServiceDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

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

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}

			service, err := client.GetService(serviceName)
			if err != nil {
				return err
			}

			// Print out machine readable output if requested
			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(service, cmd.OutOrStdout())
			}

			printDetails, err = cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			revisionDescs, err := getRevisionDescriptions(client, service, printDetails)

			return describe(cmd.OutOrStdout(), service, revisionDescs)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")
	machineReadablePrintFlags.AddFlags(command)
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
	dw.WriteColsLn(printers.Level0, l("Name"), service.Name)
	dw.WriteColsLn(printers.Level0, l("Namespace"), service.Namespace)
	dw.WriteColsLn(printers.Level0, l("URL"), extractURL(service))
	if service.Status.Address != nil {
		url := service.Status.Address.GetURL()
		dw.WriteColsLn(printers.Level0, l("Address"), url.String())
	}
	writeMapDesc(dw, printers.Level0, service.Labels, l("Labels"), "")
	writeMapDesc(dw, printers.Level0, service.Annotations, l("Annotations"), "")
	dw.WriteColsLn(printers.Level0, l("Age"), age(service.CreationTimestamp.Time))
}

// Write out revisions associated with this service. By default only active
// target revisions are printed, but with --all also inactive revisions
// created by this services are shown
func writeRevisions(dw printers.PrefixWriter, revisions []*revisionDesc) {
	dw.WriteColsLn(printers.Level0, l("Revisions"))
	for _, revisionDesc := range revisions {
		dw.WriteColsLn(printers.Level1, formatPercentage(revisionDesc.percent), l("Name"), getRevisionNameWithGenerationAndAge(revisionDesc))
		dw.WriteColsLn(printers.Level1, "", l("Image"), getImageDesc(revisionDesc))
		if revisionDesc.port != nil {
			dw.WriteColsLn(printers.Level1, "", l("Port"), strconv.FormatInt(int64(*revisionDesc.port), 10))
		}
		writeSliceDesc(dw, printers.Level1, revisionDesc.env, l("Env"), "\t")

		// Scale spec if given
		if revisionDesc.maxScale != nil || revisionDesc.minScale != nil {
			dw.WriteColsLn(printers.Level1, "", l("Scale"), formatScale(revisionDesc.minScale, revisionDesc.maxScale))
		}

		// Concurrency specs if given
		if revisionDesc.concurrencyLimit != nil || revisionDesc.concurrencyTarget != nil {
			writeConcurrencyOptions(dw, revisionDesc)
		}

		// Resources if given
		writeResources(dw, "Memory", revisionDesc.requestsMemory, revisionDesc.limitsMemory)
		writeResources(dw, "CPU", revisionDesc.requestsCPU, revisionDesc.limitsCPU)
	}
}

// Print out a table with conditions. Use green for 'ok', and red for 'nok' if color is enabled
func writeConditions(dw printers.PrefixWriter, service *v1alpha1.Service) {
	dw.WriteColsLn(printers.Level0, l("Conditions"))
	maxLen := getMaxTypeLen(service.Status.Conditions)
	formatHeader := "%-2s %-" + strconv.Itoa(maxLen) + "s %6s %-s\n"
	formatRow := "%-2s %-" + strconv.Itoa(maxLen) + "s %6s %-s\n"
	dw.Write(printers.Level1, formatHeader, "OK", "TYPE", "AGE", "REASON")
	for _, condition := range service.Status.Conditions {
		ok := formatStatus(condition.Status)
		reason := condition.Reason
		if printDetails && reason != "" {
			reason = fmt.Sprintf("%s (%s)", reason, condition.Message)
		}
		dw.Write(printers.Level1, formatRow, ok, formatConditionType(condition), age(condition.LastTransitionTime.Inner.Time), reason)
	}
}

func writeConcurrencyOptions(dw printers.PrefixWriter, desc *revisionDesc) {
	dw.WriteColsLn(printers.Level1, "", l("Concurrency"))
	if desc.concurrencyLimit != nil {
		dw.WriteColsLn(printers.Level2, "", "", l("Limit"), strconv.FormatInt(*desc.concurrencyLimit, 10))
	}
	if desc.concurrencyTarget != nil {
		dw.WriteColsLn(printers.Level2, "", "", l("Target"), strconv.Itoa(*desc.concurrencyTarget))
	}
}

// ======================================================================================
// Helper functions

// Format label (extracted so that color could be added more easily to all labels)
func l(label string) string {
	return label + ":"
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

// Format the revision name along with its generation. Use colors if enabled.
func getRevisionNameWithGenerationAndAge(desc *revisionDesc) string {
	return desc.name + " " +
		"[" + strconv.Itoa(desc.configurationGeneration) + "]" +
		" " +
		"(" + age(desc.creationTimestamp) + ")"
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
	return string(condition.Type)
}

// Status in ASCII format
func formatStatus(status corev1.ConditionStatus) string {
	switch status {
	case v1.ConditionTrue:
		return "++"
	case v1.ConditionFalse:
		return "--"
	default:
		return ""
	}
}

// Return either image name with tag or together with its resolved digest
func getImageDesc(desc *revisionDesc) string {
	image := desc.image
	if printDetails && desc.imageDigest != "" {
		return fmt.Sprintf("%s (%s)", image, shortenDigest(desc.imageDigest))
	}
	return image
}

// Extract pure sha sum and shorten to 8 digits,
// as the digest should to be user consumable. Use the resource via `kn service get`
// to get to the full sha
func shortenDigest(digest string) string {
	match := imageDigestRegexp.FindStringSubmatch(digest)
	if len(match) > 1 {
		return string(match[1][:12])
	}
	return digest
}

// Write a map either compact in a single line (possibly truncated) or, if printDetails is set,
// over multiple line, one line per key-value pair. The output is sorted by keys.
func writeMapDesc(dw printers.PrefixWriter, indent int, m map[string]string, label string, labelPrefix string) {
	if len(m) == 0 {
		return
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if printDetails {
		l := labelPrefix + label

		for _, key := range keys {
			dw.WriteColsLn(indent, l, key+"="+m[key])
			l = labelPrefix
		}
		return
	}

	dw.WriteColsLn(indent, label, joinAndTruncate(keys, m))
}

// Writer a slice compact (printDetails == false) in one line, or over multiple line
// with key-value line-by-line (printDetails == true)
func writeSliceDesc(dw printers.PrefixWriter, indent int, s []string, label string, labelPrefix string) {

	if len(s) == 0 {
		return
	}

	if printDetails {
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

// Write request ... limits or only one of them
func writeResources(dw printers.PrefixWriter, label string, request string, limit string) {
	value := ""
	if request != "" && limit != "" {
		value = request + " ... " + limit
	} else if request != "" {
		value = request
	} else if limit != "" {
		value = limit
	}

	if value == "" {
		return
	}

	dw.WriteColsLn(printers.Level1, "", l(label), value)
}

// Join to key=value pair, comma separated, and truncate if longer than a limit
func joinAndTruncate(sortedKeys []string, m map[string]string) string {
	ret := ""
	for _, key := range sortedKeys {
		ret += fmt.Sprintf("%s=%s, ", key, m[key])
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
	return fmt.Sprintf("%3d%%", percentage)
}

func age(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return duration.ShortHumanDuration(time.Now().Sub(t))
}

// Call the backend to query revisions for the given service and build up
// the view objects used for output
func getRevisionDescriptions(client serving_kn_v1alpha1.KnClient, service *v1alpha1.Service, withDetails bool) ([]*revisionDesc, error) {
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
	if withDetails {
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

func completeWithUntargetedRevisions(client serving_kn_v1alpha1.KnClient, service *v1alpha1.Service, descs map[string]*revisionDesc) error {
	revisions, err := client.ListRevisions(serving_kn_v1alpha1.WithService(service.Name))
	if err != nil {
		return err
	}
	for _, revision := range revisions.Items {
		if _, ok := descs[revision.Name]; !ok {
			descs[revision.Name], err = newRevisionDesc(&revision, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func newRevisionDesc(revision *v1alpha1.Revision, target *v1alpha1.TrafficTarget) (*revisionDesc, error) {
	container := extractContainer(revision)
	generation, err := strconv.ParseInt(revision.Labels[serving.ConfigurationGenerationLabelKey], 0, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot extract configuration generation for revision %s: %v", revision.Name, err)
	}
	revisionDesc := revisionDesc{
		name:              revision.Name,
		logURL:            revision.Status.LogURL,
		timeoutSeconds:    revision.Spec.TimeoutSeconds,
		imageDigest:       revision.Status.ImageDigest,
		creationTimestamp: revision.CreationTimestamp.Time,

		configurationGeneration: int(generation),
		configuration:           revision.Labels[serving.ConfigurationLabelKey],
	}

	addTargetInfo(&revisionDesc, target)
	addContainerInfo(&revisionDesc, container)
	addResourcesInfo(&revisionDesc, container)
	err = addConcurrencyAndScaleInfo(&revisionDesc, revision)
	if err != nil {
		return nil, err
	}
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

func addResourcesInfo(desc *revisionDesc, container *v1.Container) {
	requests := container.Resources.Requests
	if !requests.Memory().IsZero() {
		desc.requestsMemory = requests.Memory().String()
	}
	if !requests.Cpu().IsZero() {
		desc.requestsCPU = requests.Cpu().String()
	}

	limits := container.Resources.Limits
	if !limits.Memory().IsZero() {
		desc.limitsMemory = limits.Memory().String()
	}
	if !limits.Cpu().IsZero() {
		desc.limitsCPU = limits.Cpu().String()
	}
}

func addConcurrencyAndScaleInfo(desc *revisionDesc, revision *v1alpha1.Revision) error {
	min, err := annotationAsInt(revision, autoscaling.MinScaleAnnotationKey)
	if err != nil {
		return err
	}
	desc.minScale = min

	max, err := annotationAsInt(revision, autoscaling.MaxScaleAnnotationKey)
	if err != nil {
		return err
	}
	desc.maxScale = max

	target, err := annotationAsInt(revision, autoscaling.TargetAnnotationKey)
	if err != nil {
		return err
	}
	desc.concurrencyTarget = target

	if revision.Spec.ContainerConcurrency != 0 {
		limit := int64(revision.Spec.ContainerConcurrency)
		desc.concurrencyLimit = &limit
	}

	return nil
}

func annotationAsInt(revision *v1alpha1.Revision, annotationKey string) (*int, error) {
	annos := revision.Annotations
	if val, ok := annos[annotationKey]; ok {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		return &valInt, nil
	}
	return nil, nil
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

func extractRevisionFromTarget(client serving_kn_v1alpha1.KnClient, target v1alpha1.TrafficTarget) (*v1alpha1.Revision, error) {
	var revisionName = target.RevisionName
	if revisionName == "" {
		configurationName := target.ConfigurationName
		if configurationName == "" {
			return nil, fmt.Errorf("neither RevisionName nor ConfigurationName set")
		}
		configuration, err := client.GetConfiguration(configurationName)
		if err != nil {
			return nil, err
		}
		revisionName = configuration.Status.LatestCreatedRevisionName
	}
	return client.GetRevision(revisionName)
}

func extractURL(service *v1alpha1.Service) string {
	status := service.Status
	if status.URL != nil {
		return status.URL.String()
	}
	return status.DeprecatedDomain
}
