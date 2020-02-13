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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	eventinglegacy "knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	"knative.dev/client/pkg/printers"
)

var sourceTypeDescription = map[string]string{
	"ApiServerSource": "Watch and send Kubernetes API events to a sink",
	"ContainerSource": "Connect a custom container image to a sink",
	"CronJobSource":   "Send periodically constant data to a sink",
	"SinkBinding":     "Binding for connecting a PodSpecable to a sink",
	"PingSource":      "Send periodically ping events to a sink",
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
func printSource(source unstructured.Unstructured, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := source.GetName()
	sourceType := source.GetKind()
	sourceTypeName := getSourceTypeName(source)
	sink := findSink(source)
	ready := isReady(source)

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: &source},
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, source.GetNamespace())
	}

	row.Cells = append(row.Cells, name, sourceType, sourceTypeName, sink, ready)
	return []metav1beta1.TableRow{row}, nil
}

// printSourceList populates the source list table rows
func printSourceList(sourceList *unstructured.UnstructuredList, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(sourceList.Items))

	sort.SliceStable(sourceList.Items, func(i, j int) bool {
		return sourceList.Items[i].GetName() < sourceList.Items[j].GetName()
	})
	for _, item := range sourceList.Items {
		row, err := printSource(item, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}

func findSink(source unstructured.Unstructured) string {
	sourceType := source.GetKind()
	sourceJSON, err := source.MarshalJSON()
	if err != nil {
		return ""
	}

	switch sourceType {
	case "ApiServerSource":
		var apiSource eventinglegacy.ApiServerSource
		err := json.Unmarshal(sourceJSON, &apiSource)
		if err != nil {
			return ""
		}
		return sinkToString(apiSource.Spec.Sink)
	case "CronJobSource":
		var cronSource eventinglegacy.CronJobSource
		err := json.Unmarshal(sourceJSON, &cronSource)
		if err != nil {
			return ""
		}
		return sinkToString(cronSource.Spec.Sink)
	case "SinkBinding":
		var binding sourcesv1alpha1.SinkBinding
		err := json.Unmarshal(sourceJSON, &binding)
		if err != nil {
			return ""
		}
		return flags.SinkToString(binding.Spec.Sink)
	case "PingSource":
		var pingSource sourcesv1alpha1.PingSource
		err := json.Unmarshal(sourceJSON, &pingSource)
		if err != nil {
			return ""
		}
		return flags.SinkToString(*pingSource.Spec.Sink)
	// TODO: Find out how to find sink in untyped sources
	default:
		return "<unknown>"
	}
}

func isReady(source unstructured.Unstructured) string {
	var err error
	sourceType := source.GetKind()
	sourceJSON, err := source.MarshalJSON()
	if err != nil {
		return "<unknown>"
	}

	switch sourceType {
	case "ApiServerSource":
		var tSource eventinglegacy.ApiServerSource
		err = json.Unmarshal(sourceJSON, &tSource)
		if err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	case "CronJobSource":
		var tSource eventinglegacy.CronJobSource
		err = json.Unmarshal(sourceJSON, &tSource)
		if err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}

	case "SinkBinding":
		var tSource eventinglegacy.SinkBinding
		err = json.Unmarshal(sourceJSON, &tSource)
		if err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}

	case "PingSource":
		var tSource sourcesv1alpha1.PingSource
		err = json.Unmarshal(sourceJSON, &tSource)
		if err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	}

	return "<unknown>"
}

// temporary sinkToString for deprecated sources
func sinkToString(sink *duckv1beta1.Destination) string {
	if sink != nil {
		if sink.Ref != nil {
			if sink.Ref.Kind == "Service" {
				return fmt.Sprintf("svc:%s", sink.Ref.Name)
			}
			return fmt.Sprintf("%s:%s", strings.ToLower(sink.Ref.Kind), sink.Ref.Name)
		}

		if sink.URI != nil {
			return sink.URI.String()
		}
	}
	return "<unknown>"
}

func getSourceTypeName(source unstructured.Unstructured) string {
	return fmt.Sprintf("%s%s.%s",
		strings.ToLower(source.GetKind()),
		"s",
		strings.Split(source.GetAPIVersion(), "/")[0],
	)
}
