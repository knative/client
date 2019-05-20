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
	"fmt"
	hprinters "github.com/knative/client/pkg/printers"
	"github.com/knative/pkg/apis"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"time"
)

// ServiceGetFlags composes common printer flag structs
// used in the Get command.
type ServiceGetFlags struct {
	GenericPrintFlags  *genericclioptions.PrintFlags
	HumanReadableFlags *HumanPrintFlags
}

// AllowedFormats is the list of formats in which data can be displayed
func (f *ServiceGetFlags) AllowedFormats() []string {
	formats := f.GenericPrintFlags.AllowedFormats()
	formats = append(formats, f.HumanReadableFlags.AllowedFormats()...)
	return formats
}

// ToPrinter attempts to find a composed set of ServiceGetFlags suitable for
// returning a printer based on current flag values.
func (f *ServiceGetFlags) ToPrinter() (hprinters.ResourcePrinter, error) {
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
func (f *ServiceGetFlags) AddFlags(cmd *cobra.Command) {
	f.GenericPrintFlags.AddFlags(cmd)
	f.HumanReadableFlags.AddFlags(cmd)
}

// NewGetPrintFlags returns flags associated with humanreadable,
// template, and "name" printing, with default values set.
func NewServiceGetFlags() *ServiceGetFlags {
	return &ServiceGetFlags{
		GenericPrintFlags:  genericclioptions.NewPrintFlags(""),
		HumanReadableFlags: NewHumanPrintFlags(),
	}
}

// ServiceGetHandlers adds print handlers for service get command
func ServiceGetHandlers(h hprinters.PrintHandler) {
	kServiceColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Name", Type: "string", Description: "Name of the knative service."},
		{Name: "Domain", Type: "string", Description: "Domain name of the knative service."},
		//{Name: "LastCreatedRevision", Type: "string", Description: "Name of last revision created."},
		//{Name: "LastReadyRevision", Type: "string", Description: "Name of last ready revision."},
		{Name: "Generation", Type: "integer", Description: "Sequence number of 'Generation' of the service that was last processed by the controller."},
		{Name: "Age", Type: "string", Description: "Age of the service."},
		{Name: "Conditions", Type: "string", Description: "Conditions describing statuses of service components."},
		{Name: "Ready", Type: "string", Description: "Ready condition status of the service."},
		{Name: "Reason", Type: "string", Description: "Reason for non-ready condition of the service."},
	}
	h.TableHandler(kServiceColumnDefinitions, printKService)
	h.TableHandler(kServiceColumnDefinitions, printKServiceList)
}

// conditionsValue returns the True conditions count among total conditions
func conditionsValue(conditions duckv1beta1.Conditions) string {
	var ok int
	for _, condition := range conditions {
		if condition.Status == "True" {
			ok++
		}
	}
	return fmt.Sprintf("%d OK / %d", ok, len(conditions))
}

// readyCondition returns status of resource's Ready type condition
func readyCondition(conditions duckv1beta1.Conditions) string {
	for _, condition := range conditions {
		if condition.Type == apis.ConditionReady {
			return string(condition.Status)
		}
	}
	return "<unknown>"
}

func nonReadyConditionReason(conditions duckv1beta1.Conditions) string {
	for _, condition := range conditions {
		if condition.Type == apis.ConditionReady {
			if string(condition.Status) == "True" {
				return ""
			}
			if condition.Message != "" {
				return fmt.Sprintf("%s : %s", condition.Reason, condition.Message)
			}
			return string(condition.Reason)
		}
	}
	return "<unknown>"
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}

// printKServiceList populates the knative service list table rows
func printKServiceList(kServiceList *servingv1alpha1.ServiceList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
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
func printKService(kService *servingv1alpha1.Service, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := kService.Name
	domain := kService.Status.RouteStatusFields.DeprecatedDomain
	//lastCreatedRevision := kService.Status.LatestCreatedRevisionName
	//lastReadyRevision := kService.Status.LatestReadyRevisionName
	generation := kService.Status.ObservedGeneration
	age := translateTimestampSince(kService.CreationTimestamp)
	conditions := conditionsValue(kService.Status.Conditions)
	ready := readyCondition(kService.Status.Conditions)
	reason := nonReadyConditionReason(kService.Status.Conditions)

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: kService},
	}
	row.Cells = append(row.Cells,
		name,
		domain,
		//lastCreatedRevision,
		//lastReadyRevision,
		generation,
		age,
		conditions,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}
