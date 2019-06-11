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

package service

import (
	"github.com/knative/client/pkg/kn/commands"
	hprinters "github.com/knative/client/pkg/printers"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ServiceGetFlags composes common printer flag structs
// used in the Get command.
type ServiceGetFlags struct {
	GenericPrintFlags  *genericclioptions.PrintFlags
	HumanReadableFlags *commands.HumanPrintFlags
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
	p, err := f.HumanReadableFlags.ToPrinter(ServiceGetHandlers)
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
		HumanReadableFlags: commands.NewHumanPrintFlags(),
	}
}
