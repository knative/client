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

package trigger

import (
	"sort"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	hprinters "knative.dev/client/pkg/printers"
	eventing "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

// TriggerListHandlers handles printing human readable table for `kn source list-types` command's output
func TriggerListHandlers(h hprinters.PrintHandler) {
	sourceTypesColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the trigger.", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the trigger.", Priority: 1},
		{Name: "Broker", Type: "string", Description: "Name of the broker.", Priority: 1},
		{Name: "Sink", Type: "string", Description: "Sink for events, i.e. the subscriber.", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the trigger.", Priority: 1},
		{Name: "Conditions", Type: "string", Description: "Ready state conditions.", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready condition status of the trigger.", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason for non-ready condition of the trigger.", Priority: 1},
	}
	h.TableHandler(sourceTypesColumnDefinitions, printTrigger)
	h.TableHandler(sourceTypesColumnDefinitions, printTriggerList)
}

// printKService populates the knative service table rows
func printTrigger(trigger *eventing.Trigger, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := trigger.Name
	broker := trigger.Spec.Broker
	sink := flags.SinkToString(trigger.Spec.Subscriber)
	age := commands.TranslateTimestampSince(trigger.CreationTimestamp)
	conditions := commands.ConditionsValue(trigger.Status.Conditions)
	ready := commands.ReadyCondition(trigger.Status.Conditions)
	reason := commands.NonReadyConditionReason(trigger.Status.Conditions)

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: trigger},
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, trigger.Namespace)
	}

	row.Cells = append(row.Cells,
		name,
		broker,
		sink,
		age,
		conditions,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}

// printTriggerListWithNamespace populates the knative service table rows with namespace column
func printTriggerListWithNamespace(triggerList *eventing.TriggerList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(triggerList.Items))

	// temporary slice for sorting services in non-default namespace
	others := []metav1beta1.TableRow{}

	for _, trigger := range triggerList.Items {
		// Fill in with services in `default` namespace at first
		if trigger.Namespace == "default" {
			r, err := printTrigger(&trigger, options)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
			continue
		}
		// put other services in temporary slice
		r, err := printTrigger(&trigger, options)
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

// printSourceTypesList populates the source types list table rows
func printTriggerList(triggerList *eventing.TriggerList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(triggerList.Items))

	if options.AllNamespaces {
		return printTriggerListWithNamespace(triggerList, options)
	}

	for _, trigger := range triggerList.Items {
		r, err := printTrigger(&trigger, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}
