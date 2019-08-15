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

package route

import (
	"fmt"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/client/pkg/kn/commands"
	hprinters "knative.dev/client/pkg/printers"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// RouteListHandlers adds print handlers for route list command
func RouteListHandlers(h hprinters.PrintHandler) {
	kRouteColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: "Name of the Knative route."},
		{Name: "URL", Type: "string", Description: "URL of the Knative route."},
		{Name: "Age", Type: "string", Description: "Age of the Knative route."},
		{Name: "Conditions", Type: "string", Description: "Conditions describing statuses of route components."},
		{Name: "Traffic", Type: "integer", Description: "Traffic configured for route."},
	}
	h.TableHandler(kRouteColumnDefinitions, printRoute)
	h.TableHandler(kRouteColumnDefinitions, printKRouteList)
}

// printKRouteList populates the Knative route list table rows
func printKRouteList(kRouteList *servingv1alpha1.RouteList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(kRouteList.Items))
	for _, ksvc := range kRouteList.Items {
		r, err := printRoute(&ksvc, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printRoute populates the Knative route table rows
func printRoute(route *servingv1alpha1.Route, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := route.Name
	url := route.Status.URL
	age := commands.TranslateTimestampSince(route.CreationTimestamp)
	conditions := commands.ConditionsValue(route.Status.Conditions)
	traffic := calculateTraffic(route.Status.Traffic)
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: route},
	}
	row.Cells = append(row.Cells,
		name,
		url,
		age,
		conditions,
		traffic)
	return []metav1beta1.TableRow{row}, nil
}

func calculateTraffic(targets []servingv1alpha1.TrafficTarget) string {
	var traffic string
	for _, target := range targets {
		if len(traffic) > 0 {
			traffic = fmt.Sprintf("%s, %d%% -> %s", traffic, target.Percent, target.RevisionName)
		} else {
			traffic = fmt.Sprintf("%d%% -> %s", target.Percent, target.RevisionName)
		}
	}
	return traffic
}
