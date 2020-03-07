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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"

	"knative.dev/client/pkg/kn/commands"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"

	hprinters "knative.dev/client/pkg/printers"
	"knative.dev/client/pkg/util"
)

const (
	apiVersionSplitChar = ":"
)

// APIServerSourceUpdateFlags are flags for create and update a ApiServerSource
type APIServerSourceUpdateFlags struct {
	ServiceAccountName string
	Mode               string
	Resources          []string
}

// getAPIServerResourceArray is to construct an array of resources.
func (f *APIServerSourceUpdateFlags) getAPIServerResourceArray() ([]v1alpha1.ApiServerResource, error) {
	var resourceList []v1alpha1.ApiServerResource
	for _, r := range f.Resources {
		resourceSpec, err := getValidResource(r)
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, *resourceSpec)
	}
	return resourceList, nil
}

// updateExistingAPIServerResourceArray is to update an array of resources.
func (f *APIServerSourceUpdateFlags) updateExistingAPIServerResourceArray(existing []v1alpha1.ApiServerResource) ([]v1alpha1.ApiServerResource, error) {
	var found bool

	added, removed, err := f.getUpdateAPIServerResourceArray()
	if err != nil {
		return nil, err
	}

	if existing == nil {
		existing = []v1alpha1.ApiServerResource{}
	}

	existing = append(existing, added...)

	for _, item := range removed {
		found = false
		for i, ref := range existing {
			if reflect.DeepEqual(item, ref) {
				existing = append(existing[:i], existing[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cannot find resource %s:%s:%t to remove", item.Kind, item.APIVersion, item.Controller)
		}
	}
	return existing, nil
}

// getUpdateAPIServerResourceArray is to construct an array of resources for update action.
func (f *APIServerSourceUpdateFlags) getUpdateAPIServerResourceArray() ([]v1alpha1.ApiServerResource, []v1alpha1.ApiServerResource, error) {
	addedArray, removedArray := util.AddedAndRemovalListsFromArray(f.Resources)
	added, err := constructApiServerResourceArray(addedArray)
	if err != nil {
		return nil, nil, err
	}
	removed, err := constructApiServerResourceArray(removedArray)
	if err != nil {
		return nil, nil, err
	}
	return added, removed, nil
}

func constructApiServerResourceArray(s []string) ([]v1alpha1.ApiServerResource, error) {
	array := make([]v1alpha1.ApiServerResource, 0)
	for _, r := range s {
		resourceSpec, err := getValidResource(r)
		if err != nil {
			return array, err
		}
		array = append(array, *resourceSpec)
	}
	return array, nil
}

//getValidResource is to parse resource spec from a string
func getValidResource(resource string) (*v1alpha1.ApiServerResource, error) {
	var isController bool
	var err error

	parts := strings.SplitN(resource, apiVersionSplitChar, 3)
	if len(parts[0]) == 0 {
		return nil, fmt.Errorf("cannot find 'Kind' part in resource specification %s (expected: <Kind:ApiVersion[:controllerFlag]>", resource)
	}
	kind := parts[0]

	if len(parts) < 2 || len(parts[1]) == 0 {
		return nil, fmt.Errorf("cannot find 'APIVersion' part in resource specification %s (expected: <Kind:ApiVersion[:controllerFlag]>", resource)
	}
	version := parts[1]

	if len(parts) >= 3 && len(parts[2]) > 0 {
		isController, err = strconv.ParseBool(parts[2])
		if err != nil {
			return nil, fmt.Errorf("controller flag is not a boolean in resource specification %s (expected: <Kind:ApiVersion[:controllerFlag]>)", resource)
		}
	} else {
		isController = false
	}
	return &v1alpha1.ApiServerResource{Kind: kind, APIVersion: version, Controller: isController}, nil
}

//Add is to set parameters
func (f *APIServerSourceUpdateFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.ServiceAccountName,
		"service-account",
		"",
		"Name of the service account to use to run this source")
	cmd.Flags().StringVar(&f.Mode,
		"mode",
		"Ref",
		`The mode the receive adapter controller runs under:,
"Ref" sends only the reference to the resource,
"Resource" send the full resource.`)
	cmd.Flags().StringArrayVar(&f.Resources,
		"resource",
		[]string{},
		`Specification for which events to listen, in the format Kind:APIVersion:isController, e.g. "Event:v1:true".
"isController" can be omitted and is "false" by default, e.g. "Event:v1".`)
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
func printSource(source *v1alpha1.ApiServerSource, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: source},
	}

	name := source.Name
	age := commands.TranslateTimestampSince(source.CreationTimestamp)
	conditions := commands.ConditionsValue(source.Status.Conditions)
	ready := commands.ReadyCondition(source.Status.Conditions)
	reason := strings.TrimSpace(commands.NonReadyConditionReason(source.Status.Conditions))
	var resources []string
	for _, resource := range source.Spec.Resources {
		resources = append(resources, fmt.Sprintf("%s:%s:%s", resource.Kind, resource.APIVersion, strconv.FormatBool(resource.Controller)))
	}

	// Not moving to SinkToString() as it references v1beta1.Destination
	// This source is going to be moved/removed soon to v1, so no need to move
	// it now
	var sink string
	if source.Spec.Sink != nil {
		if source.Spec.Sink.Ref != nil {
			if source.Spec.Sink.Ref.Kind == "Service" {
				sink = fmt.Sprintf("svc:%s", source.Spec.Sink.Ref.Name)
			} else {
				sink = fmt.Sprintf("%s:%s", source.Spec.Sink.Ref.Kind, source.Spec.Sink.Ref.Name)
			}
		}
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, source.Namespace)
	}

	row.Cells = append(row.Cells, name, strings.Join(resources[:], ","), sink, age, conditions, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

// printSourceList populates the source apiserver list table rows
func printSourceList(sourceList *v1alpha1.ApiServerSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	if options.AllNamespaces {
		return printSourceListWithNamespace(sourceList, options)
	}

	rows := make([]metav1beta1.TableRow, 0, len(sourceList.Items))

	sort.SliceStable(sourceList.Items, func(i, j int) bool {
		return sourceList.Items[i].GetName() < sourceList.Items[j].GetName()
	})

	for _, item := range sourceList.Items {
		row, err := printSource(&item, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}

// printSourceListWithNamespace populates the knative service table rows with namespace column
func printSourceListWithNamespace(sourceList *v1alpha1.ApiServerSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(sourceList.Items))

	// temporary slice for sorting services in non-default namespace
	others := []metav1beta1.TableRow{}

	for _, source := range sourceList.Items {
		// Fill in with services in `default` namespace at first
		if source.Namespace == "default" {
			r, err := printSource(&source, options)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
			continue
		}
		// put other services in temporary slice
		r, err := printSource(&source, options)
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
