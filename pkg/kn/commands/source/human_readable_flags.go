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

package source

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	clientduck "knative.dev/client/pkg/kn/commands/source/duck"
	"knative.dev/client/pkg/printers"
)

var sourceTypeDescription = map[string]string{
	"ApiServerSource": "Watch and send Kubernetes API events to addressable",
	"SinkBinding":     "Binding for connecting a PodSpecable to addressable",
	"PingSource":      "Send periodically ping events to addressable",
	"ContainerSource": "Generate events by Container image and send to addressable",
	// TODO: source plugin could bring the description that kn could look for based on the availability
	// of the plugin and fetch the description from there, for now we dont have that capability in kn
	// so we're shipping hardcoded short description of the KafkaSource as below
	"KafkaSource": "Route events from Apache Kafka Server to addressable",
}

// ListTypesHandlers handles printing human readable table for `kn source list-types`
func ListTypesHandlers(h printers.PrintHandler) {
	sourceTypesColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Type", Type: "string", Description: "Kind / Type of the source type", Priority: 1},
		{Name: "Name", Type: "string", Description: "Name of the source type", Priority: 1},
		{Name: "Description", Type: "string", Description: "Description of the source type", Priority: 1},
	}
	h.TableHandler(sourceTypesColumnDefinitions, printSourceTypes)
	h.TableHandler(sourceTypesColumnDefinitions, printSourceTypesList)
}

// ListHandlers handles printing human readable table for `kn source list`
func ListHandlers(h printers.PrintHandler) {
	sourceListColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the source", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the created source", Priority: 1},
		{Name: "Type", Type: "string", Description: "Type of the source", Priority: 1},
		{Name: "Resource", Type: "string", Description: "Source type name", Priority: 1},
		{Name: "Sink", Type: "string", Description: "Sink of the source", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready condition status", Priority: 1},
	}
	h.TableHandler(sourceListColumnDefinitions, printSource)
	h.TableHandler(sourceListColumnDefinitions, printSourceList)
}

// printSourceTypes populates a single row of source types list table
func printSourceTypes(sourceType unstructured.Unstructured, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := sourceType.GetName()
	content := sourceType.UnstructuredContent()
	kind, found, err := unstructured.NestedString(content, "spec", "names", "kind")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("can't find specs.names.kind for %s", name)
	}

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: &sourceType},
	}
	row.Cells = append(row.Cells, kind, name, sourceTypeDescription[kind])
	return []metav1beta1.TableRow{row}, nil
}

// printSourceTypesList populates the source types list table rows
func printSourceTypesList(sourceTypesList *unstructured.UnstructuredList, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(sourceTypesList.Items))

	sort.SliceStable(sourceTypesList.Items, func(i, j int) bool {
		return sourceTypesList.Items[i].GetName() < sourceTypesList.Items[j].GetName()
	})
	for _, item := range sourceTypesList.Items {
		row, err := printSourceTypes(item, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}

// printSource populates a single row of source list table
func printSource(source *clientduck.Source, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: source},
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, source.GetNamespace())
	}

	row.Cells = append(row.Cells,
		source.Name,
		source.SourceKind,
		source.Resource,
		source.Sink,
		source.Ready,
	)
	return []metav1beta1.TableRow{row}, nil
}

// printSourceList populates the source list table rows
func printSourceList(sourceList *clientduck.SourceList, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(sourceList.Items))

	sort.SliceStable(sourceList.Items, func(i, j int) bool {
		return sourceList.Items[i].Name < sourceList.Items[j].Name
	})
	for _, source := range sourceList.Items {
		row, err := printSource(&source, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}
