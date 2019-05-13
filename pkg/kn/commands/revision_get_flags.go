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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	hprinters "github.com/knative/client/pkg/printers"
	serving "github.com/knative/serving/pkg/apis/serving"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// RevisionGetFlags composes common printer flag structs
// used in the Get command.
type RevisionGetFlags struct {
	GenericPrintFlags  *genericclioptions.PrintFlags
	HumanReadableFlags *HumanPrintFlags
}

// AllowedFormats is the list of formats in which data can be displayed
func (f *RevisionGetFlags) AllowedFormats() []string {
	formats := f.GenericPrintFlags.AllowedFormats()
	formats = append(formats, f.HumanReadableFlags.AllowedFormats()...)
	return formats
}

// ToPrinter attempts to find a composed set of RevisionGetFlags suitable for
// returning a printer based on current flag values.
func (f *RevisionGetFlags) ToPrinter() (hprinters.ResourcePrinter, error) {
	// if there are flags specified for generic printing
	if f.GenericPrintFlags.OutputFlagSpecified() {
		p, err := f.GenericPrintFlags.ToPrinter()
		if err != nil {
			return nil, err
		}
		return p, nil
	}
	// if no flags specified, use the table printing
	p, err := f.HumanReadableFlags.ToPrinter()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// AddFlags receives a *cobra.Command reference and binds
// flags related to humanreadable and template printing.
func (f *RevisionGetFlags) AddFlags(cmd *cobra.Command) {
	f.GenericPrintFlags.AddFlags(cmd)
	f.HumanReadableFlags.AddFlags(cmd)
}

// NewGetPrintFlags returns flags associated with humanreadable,
// template, and "name" printing, with default values set.
func NewRevisionGetFlags() *RevisionGetFlags {
	return &RevisionGetFlags{
		GenericPrintFlags:  genericclioptions.NewPrintFlags(""),
		HumanReadableFlags: NewHumanPrintFlags(),
	}
}

// RevisionGetHandlers adds print handlers for revision get command
func RevisionGetHandlers(h hprinters.PrintHandler) {
	RevisionColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Service", Type: "string", Description: "Name of the knative service."},
		{Name: "Name", Type: "string", Description: "Name of the revision."},
		{Name: "Age", Type: "string", Description: "Age of the revision."},
		{Name: "Conditions", Type: "string", Description: "Conditions describing statuses of revision."},
		{Name: "Ready", Type: "string", Description: "Ready condition status of the revision."},
		{Name: "Reason", Type: "string", Description: "Reason for non-ready condition of the revision."},
	}
	h.TableHandler(RevisionColumnDefinitions, printRevision)
	h.TableHandler(RevisionColumnDefinitions, printRevisionList)
}

// printRevisionList populates the knative revision list table rows
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

// printRevision populates the knative revision table rows
func printRevision(revision *servingv1alpha1.Revision, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	service := revision.Labels[serving.ConfigurationLabelKey]
	name := revision.Name
	age := translateTimestampSince(revision.CreationTimestamp)
	conditions := conditionsValue(revision.Status.Conditions)
	ready := readyCondition(revision.Status.Conditions)
	reason := nonReadyConditionReason(revision.Status.Conditions)
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: revision},
	}
	row.Cells = append(row.Cells,
		service,
		name,
		age,
		conditions,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}
