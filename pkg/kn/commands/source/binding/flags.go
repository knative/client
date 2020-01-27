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
	"strings"

	"github.com/spf13/cobra"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/tracker"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/source"
	hprinters "knative.dev/client/pkg/printers"

	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
)

type bindingUpdateFlags struct {
	subject              string
	subjectNamespace     string
	subjectSelectorParts []string
}

func (b *bindingUpdateFlags) addBindingFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&b.subject, "subject", "s", "Subject which emits cloud events")
	cmd.Flags().StringVarP(&b.subjectNamespace, "subject-namespace", "", "", "Namespace where the referenced binding subject can be found")
}

func BindingListHandlers(h hprinters.PrintHandler) {
	sourceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the sink binding", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of sink binding", Priority: 1},
		{Name: "Subject", Type: "string", Description: "Subject part of binding"},
		{Name: "Sink", Type: "string", Description: "Sink part of binding", Priority: 1},
	}
	h.TableHandler(sourceColumnDefinitions, printSinkBinding)
	h.TableHandler(sourceColumnDefinitions, printSinkBindingList)
}

// printSource populates a single row of source cronjob list table
func printSinkBinding(binding *v1alpha1.SinkBinding, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: binding},
	}

	name := binding.Name
	subject := subjectToString(binding.Spec.Subject)
	sink := source.SinkToString(binding.Spec.Sink)
	conditions := commands.ConditionsValue(binding.Status.Conditions)
	ready := commands.ReadyCondition(binding.Status.Conditions)
	reason := commands.NonReadyConditionReason(binding.Status.Conditions)

	if options.AllNamespaces {
		row.Cells = append(row.Cells, binding.Namespace)
	}

	row.Cells = append(row.Cells, name, subject, sink, conditions, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

func printSinkBindingList(sinkBindingList *v1alpha1.SinkBindingList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {

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

// subjectToString converts a reference to a string representation
func subjectToString(ref tracker.Reference) string {

	ret := ref.APIVersion + ":" + ref.Kind
	if ref.Name != "" {
		return ret + ":" + ref.Name
	}
	var keyValues []string
	selector := ref.Selector
	if selector != nil {
		for k, v := range selector.MatchLabels {
			keyValues = append(keyValues, k+"="+v)
		}
		return ret + ":" + strings.Join(keyValues, ",")
	}
	return ret
}
