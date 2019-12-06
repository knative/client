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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	hprinters "knative.dev/client/pkg/printers"
)

var sourceTypeDescription = map[string]string{
	"ApiServerSource": "Kubernetes API Server events source",
	"ContainerSource": "Container events source",
	"CronJobSource":   "CronJob events source",
}

func getSourceTypeDescription(kind string) string {
	return sourceTypeDescription[kind]
}

// ListTypesHandlers handles printing human readable table for `kn source list-types` command's output
func ListTypesHandlers(h hprinters.PrintHandler) {
	sourceTypesColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Type", Type: "string", Description: "Kind / Type of the source type", Priority: 1},
		{Name: "Name", Type: "string", Description: "Name of the source type", Priority: 1},
		{Name: "Description", Type: "string", Description: "Description of the source type", Priority: 1},
	}
	h.TableHandler(sourceTypesColumnDefinitions, printSourceTypes)
	h.TableHandler(sourceTypesColumnDefinitions, printSourceTypesList)
}

// printSourceTypes populates a single row of source types list table
func printSourceTypes(sourceType unstructured.Unstructured, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := sourceType.GetName()
	content := sourceType.UnstructuredContent()
	kind, found, err := unstructured.NestedString(content, "spec", "names", "kind")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("can't find kind of CRD for %s", name)
	}

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: &sourceType},
	}

	row.Cells = append(row.Cells, kind, name, getSourceTypeDescription(kind))
	return []metav1beta1.TableRow{row}, nil
}

// printSourceTypesList populates the source types list table rows
func printSourceTypesList(sourceTypesList *unstructured.UnstructuredList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(sourceTypesList.Items))

	for _, item := range sourceTypesList.Items {
		row, err := printSourceTypes(item, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}
