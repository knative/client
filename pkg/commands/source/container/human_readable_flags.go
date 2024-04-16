/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

import (
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	sinkfl "knative.dev/client-pkg/pkg/commands/flags/sink"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"knative.dev/client/pkg/commands"

	hprinters "knative.dev/client-pkg/pkg/printers"
)

// ContainerSourceListHandlers handles printing human readable table for `kn source apiserver list` command's output
func ContainerSourceListHandlers(h hprinters.PrintHandler) {
	sourceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Container source", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Container source", Priority: 1},
		{Name: "Image", Type: "string", Description: "Image URI configured for the Container source", Priority: 1},
		{Name: "Sink", Type: "string", Description: "Sink of the Container source", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the Container source", Priority: 1},
		{Name: "Conditions", Type: "string", Description: "Ready state conditions", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the Container source", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason if state is not Ready", Priority: 1},
	}
	h.TableHandler(sourceColumnDefinitions, printSource)
	h.TableHandler(sourceColumnDefinitions, printSourceList)
}

// printSource populates a single row of source apiserver list table
func printSource(source *v1.ContainerSource, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: source},
	}

	name := source.Name
	age := commands.TranslateTimestampSince(source.CreationTimestamp)
	conditions := commands.ConditionsValue(source.Status.Conditions)
	ready := commands.ReadyCondition(source.Status.Conditions)
	reason := strings.TrimSpace(commands.NonReadyConditionReason(source.Status.Conditions))
	image := source.Spec.Template.Spec.Containers[0].Image

	// Not moving to SinkToString() as it references v1beta1.Destination
	// This source is going to be moved/removed soon to v1, so no need to move
	// it now
	sink := sinkfl.SinkToString(source.Spec.Sink)

	if options.AllNamespaces {
		row.Cells = append(row.Cells, source.Namespace)
	}

	row.Cells = append(row.Cells, name, image, sink, age, conditions, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

// printSourceList populates the source apiserver list table rows
func printSourceList(sourceList *v1.ContainerSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
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
func printSourceListWithNamespace(sourceList *v1.ContainerSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
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
