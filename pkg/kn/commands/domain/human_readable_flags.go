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

package domain

import (
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/serving/pkg/apis/serving/v1beta1"

	"knative.dev/client/pkg/kn/commands"
	hprinters "knative.dev/client/pkg/printers"
)

// DomainMappingListHandlers adds print handlers for route list command
func DomainMappingListHandlers(h hprinters.PrintHandler) {
	dmColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Knative service", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Knative domain mapping.", Priority: 1},
		{Name: "URL", Type: "string", Description: "URL of the Knative domain mapping.", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready condition status of the Knative domain mapping.", Priority: 1},
		{Name: "Ksvc", Type: "string", Description: "Name of the referenced Knative service", Priority: 1},
	}
	h.TableHandler(dmColumnDefinitions, printDomainMapping)
	h.TableHandler(dmColumnDefinitions, printDomainMappingList)
}

// printDomainMappingList populates the Knative domain mapping list table rows
func printDomainMappingList(domainMappingList *v1beta1.DomainMappingList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(domainMappingList.Items))
	for i := range domainMappingList.Items {
		dm := &domainMappingList.Items[i]
		r, err := printDomainMapping(dm, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printDomainMapping populates the Knative domain mapping table rows
func printDomainMapping(domainMapping *v1beta1.DomainMapping, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := domainMapping.Name
	url := domainMapping.Status.URL
	ready := commands.ReadyCondition(domainMapping.Status.Conditions)
	ksvc := domainMapping.Spec.Ref.Name
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: domainMapping},
	}
	if options.AllNamespaces {
		row.Cells = append(row.Cells, domainMapping.Namespace)
	}

	row.Cells = append(row.Cells,
		name,
		url,
		ready,
		ksvc)
	return []metav1beta1.TableRow{row}, nil
}
