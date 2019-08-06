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

package revision

import (
	"github.com/knative/client/pkg/kn/commands"
	hprinters "github.com/knative/client/pkg/printers"
	serving "github.com/knative/serving/pkg/apis/serving"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// RevisionListHandlers adds print handlers for revision list command
func RevisionListHandlers(h hprinters.PrintHandler) {
	RevisionColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: "Name of the revision."},
		{Name: "Service", Type: "string", Description: "Name of the Knative service."},
		{Name: "Generation", Type: "string", Description: "Generation of the revision"},
		{Name: "Age", Type: "string", Description: "Age of the revision."},
		{Name: "Conditions", Type: "string", Description: "Conditions describing statuses of the revision."},
		{Name: "Ready", Type: "string", Description: "Ready condition status of the revision."},
		{Name: "Reason", Type: "string", Description: "Reason for non-ready condition of the revision."},
	}
	h.TableHandler(RevisionColumnDefinitions, printRevision)
	h.TableHandler(RevisionColumnDefinitions, printRevisionList)
}

// Private functions

// printRevisionList populates the Knative revision list table rows
func printRevisionList(revisionList *servingv1alpha1.RevisionList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(revisionList.Items))
	for _, rev := range revisionList.Items {
		r, err := printRevision(&rev, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printRevision populates the Knative revision table rows
func printRevision(revision *servingv1alpha1.Revision, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := revision.Name
	service := revision.Labels[serving.ServiceLabelKey]
	generation := revision.Labels[serving.ConfigurationGenerationLabelKey]
	age := commands.TranslateTimestampSince(revision.CreationTimestamp)
	conditions := commands.ConditionsValue(revision.Status.Conditions)
	ready := commands.ReadyCondition(revision.Status.Conditions)
	reason := commands.NonReadyConditionReason(revision.Status.Conditions)
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: revision},
	}
	row.Cells = append(row.Cells,
		name,
		service,
		generation,
		age,
		conditions,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}
