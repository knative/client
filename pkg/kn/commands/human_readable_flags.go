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

package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	hprinters "knative.dev/client/pkg/printers"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// HumanPrintFlags provides default flags necessary for printing.
// Given the following flag values, a printer can be requested that knows
// how to handle printing based on these values.
type HumanPrintFlags struct {
	WithNamespace bool
	NoHeaders     bool
	//TODO: Add more flags as required
}

// AllowedFormats returns more customized formating options
func (f *HumanPrintFlags) AllowedFormats() []string {
	// TODO: Add more formats eg: wide
	return []string{"no-headers"}
}

// ToPrinter receives returns a printer capable of
// handling human-readable output.
func (f *HumanPrintFlags) ToPrinter(getHandlerFunc func(h hprinters.PrintHandler)) (hprinters.ResourcePrinter, error) {
	p := hprinters.NewTablePrinter(hprinters.PrintOptions{AllNamespaces: f.WithNamespace, NoHeaders: f.NoHeaders})
	getHandlerFunc(p)
	return p, nil
}

// AddFlags receives a *cobra.Command reference and binds
// flags related to human-readable printing to it
func (f *HumanPrintFlags) AddFlags(c *cobra.Command) {
	c.Flags().BoolVar(&f.NoHeaders, "no-headers", false, "When using the default output format, don't print headers (default: print headers).")
	//TODO: Add more flags as required
}

// NewHumanPrintFlags returns flags associated with
// human-readable printing, with default values set.
func NewHumanPrintFlags() *HumanPrintFlags {
	return &HumanPrintFlags{}
}

// EnsureWithNamespace sets the "WithNamespace" humanreadable option to true.
func (f *HumanPrintFlags) EnsureWithNamespace() {
	f.WithNamespace = true
}

// conditionsValue returns the True conditions count among total conditions
func ConditionsValue(conditions duckv1.Conditions) string {
	var ok int
	for _, condition := range conditions {
		if condition.Status == "True" {
			ok++
		}
	}
	return fmt.Sprintf("%d OK / %d", ok, len(conditions))
}

// readyCondition returns status of resource's Ready type condition
func ReadyCondition(conditions duckv1.Conditions) string {
	for _, condition := range conditions {
		if condition.Type == apis.ConditionReady {
			return string(condition.Status)
		}
	}
	return "<unknown>"
}

// NonReadyConditionReason returns formatted string of
// reason and message for non ready conditions
func NonReadyConditionReason(conditions duckv1.Conditions) string {
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
func TranslateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(timestamp.Time))
}
