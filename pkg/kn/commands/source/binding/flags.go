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

package binding

import (
	"github.com/spf13/cobra"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	hprinters "knative.dev/client/pkg/printers"

	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
)

type bindingUpdateFlags struct {
	subject     string
	ceOverrides []string
}

func (b *bindingUpdateFlags) addBindingFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&b.subject, "subject", "", "Subject which emits cloud events. This argument takes format kind:apiVersion:name for named resources or kind:apiVersion:labelKey1=value1,labelKey2=value2 for matching via a label selector")
	cmd.Flags().StringArrayVar(&b.ceOverrides, "ce-override", nil, "Cloud Event overrides to apply before sending event to sink in the format '--ce-override key=value'. --ce-override can be provide multiple times")
}

func BindingListHandlers(h hprinters.PrintHandler) {
	sourceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the sink binding", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of sink binding", Priority: 1},
		{Name: "Subject", Type: "string", Description: "Subject part of binding", Priority: 1},
		{Name: "Sink", Type: "string", Description: "Sink part of binding", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of binding", Priority: 1},
		{Name: "Conditions", Type: "string", Description: "Ready state conditions", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the sink binding", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason if state is not Ready", Priority: 1},
	}
	h.TableHandler(sourceColumnDefinitions, printSinkBinding)
	h.TableHandler(sourceColumnDefinitions, printSinkBindingList)
}

// printSinkBinding populates a single row of source sink binding list table
func printSinkBinding(binding *v1alpha2.SinkBinding, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: binding},
	}

	name := binding.Name
	subject := subjectToString(binding.Spec.Subject)
	sink := flags.SinkToString(binding.Spec.Sink)
	age := commands.TranslateTimestampSince(binding.CreationTimestamp)
	conditions := commands.ConditionsValue(binding.Status.Conditions)
	ready := commands.ReadyCondition(binding.Status.Conditions)
	reason := commands.NonReadyConditionReason(binding.Status.Conditions)

	if options.AllNamespaces {
		row.Cells = append(row.Cells, binding.Namespace)
	}

	row.Cells = append(row.Cells, name, subject, sink, age, conditions, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

func printSinkBindingList(sinkBindingList *v1alpha2.SinkBindingList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {

	rows := make([]metav1beta1.TableRow, 0, len(sinkBindingList.Items))
	for _, binding := range sinkBindingList.Items {
		r, err := printSinkBinding(&binding, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}
