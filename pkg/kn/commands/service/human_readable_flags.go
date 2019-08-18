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
	"sort"

	"github.com/knative/client/pkg/kn/commands"
	hprinters "github.com/knative/client/pkg/printers"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/client/pkg/kn/commands"
	hprinters "knative.dev/client/pkg/printers"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// ServiceListHandlers adds print handlers for service list command
func ServiceListHandlers(h hprinters.PrintHandler) {
	kServiceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Knative service", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Knative service.", Priority: 1},
		{Name: "Url", Type: "string", Description: "URL of the Knative service.", Priority: 1},
		//{Name: "LastCreatedRevision", Type: "string", Description: "Name of last revision created.", Priority: 1},
		//{Name: "LastReadyRevision", Type: "string", Description: "Name of last ready revision.", Priority: 1},
		{Name: "Generation", Type: "integer", Description: "Sequence number of 'Generation' of the service that was last processed by the controller.", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the service.", Priority: 1},
		{Name: "Conditions", Type: "string", Description: "Conditions describing statuses of service components.", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready condition status of the service.", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason for non-ready condition of the service.", Priority: 1},
	}

	h.TableHandler(kServiceColumnDefinitions, printKService)
	h.TableHandler(kServiceColumnDefinitions, printKServiceList)
}

// Private functions

// printKServiceList populates the knative service list table rows
func printKServiceList(kServiceList *servingv1alpha1.ServiceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(kServiceList.Items))

	if options.AllNamespaces {
		return printKServiceWithNaemspace(kServiceList, options)
	}

	for _, ksvc := range kServiceList.Items {
		r, err := printKService(&ksvc, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printKServiceWithNaemspace populates the knative service table rows with namespace column
func printKServiceWithNaemspace(kServiceList *servingv1alpha1.ServiceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(kServiceList.Items))

	// temporary slice for sorting services in non-default namespace
	others := []metav1beta1.TableRow{}

	for _, ksvc := range kServiceList.Items {
		// Fill in with services in `default` namespace at first
		if ksvc.Namespace == "default" {
			r, err := printKService(&ksvc, options)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
			continue
		}
		// put other services in temporary slice
		r, err := printKService(&ksvc, options)
		if err != nil {
			return nil, err
		}
		others = append(others, r...)
	}

	// sort other services list alphabetically by namespace name
	sort.SliceStable(others, func(i, j int) bool {
		return others[i].Cells[0].(string) < others[j].Cells[0].(string)
	})

	return append(rows, others...), nil
}

// printKService populates the knative service table rows
func printKService(kService *servingv1alpha1.Service, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := kService.Name
	url := kService.Status.URL
	//lastCreatedRevision := kService.Status.LatestCreatedRevisionName
	//lastReadyRevision := kService.Status.LatestReadyRevisionName
	generation := kService.Status.ObservedGeneration
	age := commands.TranslateTimestampSince(kService.CreationTimestamp)
	conditions := commands.ConditionsValue(kService.Status.Conditions)
	ready := commands.ReadyCondition(kService.Status.Conditions)
	reason := commands.NonReadyConditionReason(kService.Status.Conditions)

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: kService},
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, kService.Namespace)
	}

	row.Cells = append(row.Cells,
		name,
		url,
		//lastCreatedRevision,
		//lastReadyRevision,
		generation,
		age,
		conditions,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}
