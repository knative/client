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
	hprinters "github.com/knative/client/pkg/printers"
	"github.com/spf13/cobra"
)

// HumanPrintFlags provides default flags necessary for printing.
// Given the following flag values, a printer can be requested that knows
// how to handle printing based on these values.
type HumanPrintFlags struct {
	//TODO: Add more flags as required
}

// AllowedFormats returns more customized formating options
func (f *HumanPrintFlags) AllowedFormats() []string {
	// TODO: Add more formats eg: wide
	return []string{""}
}

// ToPrinter receives returns a printer capable of
// handling human-readable output.
func (f *HumanPrintFlags) ToPrinter() (hprinters.ResourcePrinter, error) {
	p := hprinters.NewTablePrinter(hprinters.PrintOptions{})
	// Add the column definitions and respective functions
	ServiceGetHandlers(p)
	return p, nil
}

// AddFlags receives a *cobra.Command reference and binds
// flags related to human-readable printing to it
func (f *HumanPrintFlags) AddFlags(c *cobra.Command) {
	//TODO: Add more flags as required
}

// NewHumanPrintFlags returns flags associated with
// human-readable printing, with default values set.
func NewHumanPrintFlags() *HumanPrintFlags {
	return &HumanPrintFlags{}
}
