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

package apiserver

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	hprinters "knative.dev/client/pkg/printers"
	"knative.dev/client/pkg/util"
)

const (
	resourceSplitChar = ":"
	labelSplitChar    = ","
	keyValueSplitChar = "="
)

// APIServerSourceUpdateFlags are flags for create and update a ApiServerSource
type APIServerSourceUpdateFlags struct {
	ServiceAccountName string
	Mode               string
	Resources          []string
	ceOverrides        []string
}

// getAPIServerVersionKindSelector is to construct an array of resources.
func (f *APIServerSourceUpdateFlags) getAPIServerVersionKindSelector() ([]sourcesv1.APIVersionKindSelector, error) {
	resourceList := make([]sourcesv1.APIVersionKindSelector, 0, len(f.Resources))
	for _, r := range f.Resources {
		resourceSpec, err := getValidAPIVersionKindSelector(r)
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, *resourceSpec)
	}
	return resourceList, nil
}

// updateExistingAPIVersionKindSelectorArray is to update an array of resources.
func (f *APIServerSourceUpdateFlags) updateExistingAPIVersionKindSelectorArray(existing []sourcesv1.APIVersionKindSelector) ([]sourcesv1.APIVersionKindSelector, error) {
	added, removed, err := f.getUpdateAPIVersionKindSelectorArray()
	if err != nil {
		return nil, err
	}

	if existing == nil {
		existing = []sourcesv1.APIVersionKindSelector{}
	}

	existing = append(existing, added...)

	var updated []sourcesv1.APIVersionKindSelector
OUTER:
	for _, ref := range existing {
		for i, item := range removed {
			if reflect.DeepEqual(item, ref) {
				removed = append(removed[:i], removed[i+1:]...)
				continue OUTER
			}
		}
		updated = append(updated, ref)
	}

	if len(removed) > 0 {
		var errTxt []string
		for _, item := range removed {
			errTxt = append(errTxt, apiVersionKindSelectorToString(item))
		}
		return nil, fmt.Errorf("cannot find resources to remove: %s", strings.Join(errTxt, ", "))
	}

	return updated, nil
}

// getUpdateAPIVersionKindSelectorArray is to construct an array of resources for update action.
func (f *APIServerSourceUpdateFlags) getUpdateAPIVersionKindSelectorArray() ([]sourcesv1.APIVersionKindSelector, []sourcesv1.APIVersionKindSelector, error) {
	addedArray, removedArray := util.AddedAndRemovalListsFromArray(f.Resources)
	added, err := constructAPIVersionKindSelector(addedArray)
	if err != nil {
		return nil, nil, err
	}
	removed, err := constructAPIVersionKindSelector(removedArray)
	if err != nil {
		return nil, nil, err
	}
	return added, removed, nil
}

func constructAPIVersionKindSelector(s []string) ([]sourcesv1.APIVersionKindSelector, error) {
	array := make([]sourcesv1.APIVersionKindSelector, 0)
	for _, r := range s {
		resourceSpec, err := getValidAPIVersionKindSelector(r)
		if err != nil {
			return array, err
		}
		array = append(array, *resourceSpec)
	}
	return array, nil
}

// getValidAPIVersionKindSelector is to parse resource spec from a string
func getValidAPIVersionKindSelector(resource string) (*sourcesv1.APIVersionKindSelector, error) {
	var err error

	parts := strings.SplitN(resource, resourceSplitChar, 3)
	if len(parts[0]) == 0 {
		return nil, fmt.Errorf("cannot find 'kind' part in resource specification %s (expected: <kind:apiVersion[:label1=val1,label2=val2,..]>", resource)
	}
	kind := parts[0]

	if len(parts) < 2 || len(parts[1]) == 0 {
		return nil, fmt.Errorf("cannot find 'APIVersion' part in resource specification %s (expected: <kind:apiVersion[:label1=val1,label2=val2,..]>", resource)
	}
	version := parts[1]

	labelSelector, err := extractLabelSelector(parts)
	if err != nil {
		return nil, err
	}
	return &sourcesv1.APIVersionKindSelector{Kind: kind, APIVersion: version, LabelSelector: labelSelector}, nil
}

// Add is to set parameters
func (f *APIServerSourceUpdateFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ServiceAccountName,
		"service-account",
		"",
		"Name of the service account to use to run this source")
	cmd.Flags().StringVar(&f.Mode,
		"mode",
		"Reference",
		`The mode the receive adapter controller runs under:,
"Reference" sends only the reference to the resource,
"Resource" send the full resource.`)
	cmd.Flags().StringArrayVar(&f.Resources,
		"resource",
		[]string{},
		`Specification for which events to listen, in the format Kind:APIVersion:LabelSelector, e.g. "Event:sourcesv1:key=value".
"LabelSelector" is a list of comma separated key value pairs. "LabelSelector" can be omitted, e.g. "Event:sourcesv1".`)
	cmd.Flags().StringArrayVar(&f.ceOverrides,
		"ce-override",
		[]string{},
		"Cloud Event overrides to apply before sending event to sink. "+
			"Example: '--ce-override key=value' "+
			"You may be provide this flag multiple times. "+
			"To unset, append \"-\" to the key (e.g. --ce-override key-).")
}

// APIServerSourceListHandlers handles printing human readable table for `kn source apiserver list` command's output
func APIServerSourceListHandlers(h hprinters.PrintHandler) {
	sourceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the ApiServer source", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the ApiServer source", Priority: 1},
		{Name: "Resources", Type: "string", Description: "Event resources configured for the ApiServer source", Priority: 1},
		{Name: "Sink", Type: "string", Description: "Sink of the ApiServer source", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the ApiServer source", Priority: 1},
		{Name: "Conditions", Type: "string", Description: "Ready state conditions", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the ApiServer source", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason if state is not Ready", Priority: 1},
	}
	h.TableHandler(sourceColumnDefinitions, printSource)
	h.TableHandler(sourceColumnDefinitions, printSourceList)
}

// printSource populates a single row of source apiserver list table
func printSource(source *sourcesv1.ApiServerSource, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: source},
	}

	name := source.Name
	age := commands.TranslateTimestampSince(source.CreationTimestamp)
	conditions := commands.ConditionsValue(source.Status.Conditions)
	ready := commands.ReadyCondition(source.Status.Conditions)
	reason := strings.TrimSpace(commands.NonReadyConditionReason(source.Status.Conditions))
	resources := make([]string, 0, len(source.Spec.Resources))
	for _, resource := range source.Spec.Resources {
		resources = append(resources, apiVersionKindSelectorToString(resource))
	}

	// Not moving to SinkToString() as it references v1beta1.Destination
	// This source is going to be moved/removed soon to sourcesv1, so no need to move
	// it now
	sink := flags.SinkToString(source.Spec.Sink)

	if options.AllNamespaces {
		row.Cells = append(row.Cells, source.Namespace)
	}

	row.Cells = append(row.Cells, name, strings.Join(resources[:], ","), sink, age, conditions, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

func apiVersionKindSelectorToString(apiVersionKindSelector sourcesv1.APIVersionKindSelector) string {
	lTxt := labelSelectorToString(apiVersionKindSelector.LabelSelector)
	if lTxt != "" {
		lTxt = ":" + lTxt
	}
	return fmt.Sprintf("%s:%s%s", apiVersionKindSelector.Kind, apiVersionKindSelector.APIVersion, lTxt)
}

func labelSelectorToString(labelSelector *metav1.LabelSelector) string {
	if labelSelector == nil {
		return ""
	}
	labelsMap := labelSelector.MatchLabels
	if len(labelsMap) != 9 {
		keys := make([]string, 0, len(labelsMap))
		labels := make([]string, 0, len(labelsMap))
		for k := range labelsMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			labels[i] = k + "=" + labelsMap[k]
		}
		return strings.Join(labels, ",")
	}
	if len(labelSelector.MatchExpressions) > 0 {
		return "[match expressions]"
	}
	return ""
}

// printSourceList populates the source apiserver list table rows
func printSourceList(sourceList *sourcesv1.ApiServerSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	if options.AllNamespaces {
		return printSourceListWithNamespace(sourceList, options)
	}

	rows := make([]metav1beta1.TableRow, 0, len(sourceList.Items))

	sort.SliceStable(sourceList.Items, func(i, j int) bool {
		return sourceList.Items[i].GetName() < sourceList.Items[j].GetName()
	})

	for i := range sourceList.Items {
		item := &sourceList.Items[i]
		row, err := printSource(item, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}

// printSourceListWithNamespace populates the knative service table rows with namespace column
func printSourceListWithNamespace(sourceList *sourcesv1.ApiServerSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(sourceList.Items))

	// temporary slice for sorting services in non-default namespace
	others := []metav1beta1.TableRow{}

	for i := range sourceList.Items {
		source := &sourceList.Items[i]
		// Fill in with services in `default` namespace at first
		if source.Namespace == "default" {
			r, err := printSource(source, options)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
			continue
		}
		// put other services in temporary slice
		r, err := printSource(source, options)
		if err != nil {
			return nil, err
		}
		others = append(others, r...)
	}

	// sort other services list alphabetically by namespace
	sort.SliceStable(others, func(i, j int) bool {
		return others[i].Cells[0].(string) < others[j].Cells[0].(string)
	})

	return append(rows, others...), nil
}

func extractLabelSelector(parts []string) (*metav1.LabelSelector, error) {
	var labelSelector *metav1.LabelSelector
	if len(parts) >= 3 && len(parts[2]) > 0 {
		labelParts := strings.Split(parts[2], labelSplitChar)

		labels := make(map[string]string)
		for _, keyValue := range labelParts {
			keyValueParts := strings.SplitN(keyValue, keyValueSplitChar, 2)
			if len(keyValueParts) != 2 {
				return nil, fmt.Errorf("invalid label selector in resource specification %s (expected: <kind:apiVersion[:label1=val1,label2=val2,..]>", strings.Join(parts, resourceSplitChar))
			}
			labels[keyValueParts[0]] = keyValueParts[1]
		}
		labelSelector = &metav1.LabelSelector{MatchLabels: labels}
	}
	return labelSelector, nil
}
