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
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	hprinters "knative.dev/client/pkg/printers"
)

// ServiceListHandlers adds print handlers for service list command
func ServiceListHandlers(h hprinters.PrintHandler) {
	kServiceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Knative service", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Knative service.", Priority: 1},
		{Name: "Url", Type: "string", Description: "URL of the Knative service.", Priority: 1},
		{Name: "Latest", Type: "string", Description: "Name of the latest ready revision.", Priority: 1},
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
func printKServiceList(kServiceList *servingv1.ServiceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(kServiceList.Items))

	for _, ksvc := range kServiceList.Items {
		r, err := printKService(&ksvc, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printKService populates the knative service table rows
func printKService(kService *servingv1.Service, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := kService.Name
	url := kService.Status.URL
	latestRevision := kService.Status.ConfigurationStatusFields.LatestReadyRevisionName
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
		latestRevision,
		age,
		conditions,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}
