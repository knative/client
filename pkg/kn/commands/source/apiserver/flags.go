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
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	hprinters "knative.dev/client/pkg/printers"
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

type resourceSpec struct {
	kind         string
	apiVersion   string
	isController bool
}

// GetAPIServerResourceArray is to return an array of ApiServerResource from a string. A sample is Event:v1:true,Pod:v2:false
func (f *APIServerSourceUpdateFlags) GetAPIServerResourceArray() (*[]v1alpha1.ApiServerResource, error) {
	var resourceList []v1alpha1.ApiServerResource
	for _, r := range f.Resources {
		resourceSpec, err := getValidResource(r)
		if err != nil {
			return nil, err
		}
		resourceRef := v1alpha1.ApiServerResource{
			APIVersion: resourceSpec.apiVersion,
			Kind:       resourceSpec.kind,
			Controller: resourceSpec.isController,
		}
		resourceList = append(resourceList, resourceRef)
	}
	return &resourceList, nil
}

func getValidResource(resource string) (*resourceSpec, error) {
	var isController = false //false as default
	var err error

	parts := strings.Split(resource, apiVersionSplitChar)
	kind := parts[0]
	if len(parts) < 2 {
		return nil, fmt.Errorf("no APIVersion given for resource %s", resource)
	}
	version := parts[1]
	if len(parts) >= 3 {
		isController, err = strconv.ParseBool(parts[2])
		if err != nil {
			return nil, fmt.Errorf("cannot parse controller flage in resource specification %s", resource)
		}
	}
	return &resourceSpec{apiVersion: version, kind: kind, isController: isController}, nil
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
	cmd.Flags().StringSliceVar(&f.Resources,
		"resource",
		nil,
		`Specification for which events to listen, in the format Kind:APIVersion:isController, e.g. Deployment:apps/v1:true.
"isController" can be omitted and is "false" by default.`)
}

// APIServerSourceListHandlers handles printing human readable table for `kn source apiserver list` command's output
func APIServerSourceListHandlers(h hprinters.PrintHandler) {
	sourceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the ApiServer source", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the ApiServer source", Priority: 1},
		{Name: "Resources", Type: "string", Description: "Event resources configured for the ApiServer source", Priority: 1},
		{Name: "Sink", Type: "string", Description: "Sink of the ApiServer source", Priority: 1},
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
	conditions := commands.ConditionsValue(source.Status.Conditions)
	ready := commands.ReadyCondition(source.Status.Conditions)
	reason := strings.TrimSpace(commands.NonReadyConditionReason(source.Status.Conditions))
	var resources []string
	for _, resource := range source.Spec.Resources {
		resources = append(resources, fmt.Sprintf("%s:%s:%s", resource.Kind, resource.APIVersion, strconv.FormatBool(resource.Controller)))
	}

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

	row.Cells = append(row.Cells, name, strings.Join(resources[:], ","), sink, conditions, ready, reason)
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
