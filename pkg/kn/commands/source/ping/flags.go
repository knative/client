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

package ping

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client/pkg/kn/commands"
	hprinters "knative.dev/client/pkg/printers"

	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
)

type pingUpdateFlags struct {
	schedule string
	data     string
}

func (c *pingUpdateFlags) addPingFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.schedule, "schedule", "", "Optional schedule specification in crontab format (e.g. '*/2 * * * *' for every two minutes. By default fire every minute.")
	cmd.Flags().StringVarP(&c.data, "data", "d", "", "Json data to send")
}

// PingListHandlers handles printing human readable table for `kn source ping list` command's output
func PingSourceListHandlers(h hprinters.PrintHandler) {
	sourceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Ping source", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Ping source", Priority: 1},
		{Name: "Schedule", Type: "string", Description: "Schedule of the Ping source", Priority: 1},
		{Name: "Sink", Type: "string", Description: "Sink of the Ping source", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the Ping source", Priority: 1},
		{Name: "Conditions", Type: "string", Description: "Ready state conditions", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the Ping source", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason if state is not Ready", Priority: 1},
	}
	h.TableHandler(sourceColumnDefinitions, printSource)
	h.TableHandler(sourceColumnDefinitions, printSourceList)
}

// printSource populates a single row of Ping source list
func printSource(source *v1alpha2.PingSource, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: source},
	}

	name := source.Name
	schedule := source.Spec.Schedule
	age := commands.TranslateTimestampSince(source.CreationTimestamp)
	conditions := commands.ConditionsValue(source.Status.Conditions)
	ready := commands.ReadyCondition(source.Status.Conditions)
	reason := commands.NonReadyConditionReason(source.Status.Conditions)

	// Not moving to SinkToString() as it references v1beta1.Destination
	// This source is going to be moved/removed soon to v1, so no need to move
	// it now
	var sink string
	if source.Spec.Sink.Ref != nil {
		if source.Spec.Sink.Ref.Kind == "Service" {
			sink = fmt.Sprintf("svc:%s", source.Spec.Sink.Ref.Name)
		} else {
			sink = fmt.Sprintf("%s:%s", source.Spec.Sink.Ref.Kind, source.Spec.Sink.Ref.Name)
		}
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, source.Namespace)
	}

	row.Cells = append(row.Cells, name, schedule, sink, age, conditions, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

// printSourceList populates the Ping source list table rows
func printSourceList(sourceList *v1alpha2.PingSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
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
func printSourceListWithNamespace(sourceList *v1alpha2.PingSourceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(sourceList.Items))

	// temporary slice for sorting services in non-default namespace
	var others []metav1beta1.TableRow

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
