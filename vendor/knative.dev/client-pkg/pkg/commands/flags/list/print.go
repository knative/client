/*
 Copyright 2024 The Knative Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package list

import (
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"knative.dev/client-pkg/pkg/commands"

	hprinters "knative.dev/client-pkg/pkg/printers"
	"knative.dev/client-pkg/pkg/util"
)

// PrintFlags composes common printer flag structs used in the list command.
type PrintFlags struct {
	GenericPrintFlags  *genericclioptions.PrintFlags
	HumanReadableFlags *commands.HumanPrintFlags
	PrinterHandler     func(h hprinters.PrintHandler)
}

// AllowedFormats is the list of formats in which data can be displayed
func (f *PrintFlags) AllowedFormats() []string {
	formats := f.GenericPrintFlags.AllowedFormats()
	formats = append(formats, f.HumanReadableFlags.AllowedFormats()...)
	return formats
}

// ToPrinter attempts to find a composed set of ListTypesFlags suitable for
// returning a printer based on current flag values.
func (f *PrintFlags) ToPrinter() (hprinters.ResourcePrinter, error) {
	// if there are flags specified for generic printing
	if f.GenericPrintFlags.OutputFlagSpecified() {
		p, err := f.GenericPrintFlags.ToPrinter()
		if err != nil {
			return nil, err
		}
		return p, nil
	}

	p, err := f.HumanReadableFlags.ToPrinter(f.PrinterHandler)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Print is to print an Object to a Writer
func (f *PrintFlags) Print(obj runtime.Object, w io.Writer) error {
	printer, err := f.ToPrinter()
	if err != nil {
		return err
	}

	if f.GenericPrintFlags.OutputFlagSpecified() {
		unstructuredList, err := util.ToUnstructuredList(obj)
		if err != nil {
			return err
		}
		return printer.PrintObj(unstructuredList, w)
	}

	return printer.PrintObj(obj, w)
}

// AddFlags receives a *cobra.Command reference and binds
// flags related to humanreadable and template printing.
func (f *PrintFlags) AddFlags(cmd *cobra.Command) {
	f.GenericPrintFlags.AddFlags(cmd)
	f.HumanReadableFlags.AddFlags(cmd)
}

// NewPrintFlags returns flags associated with humanreadable, template, and
// "name" printing, with default values set.
func NewPrintFlags(printer func(h hprinters.PrintHandler)) *PrintFlags {
	return &PrintFlags{
		GenericPrintFlags:  genericclioptions.NewPrintFlags(""),
		HumanReadableFlags: commands.NewHumanPrintFlags(),
		PrinterHandler:     printer,
	}
}

// EnsureWithNamespace ensures that humanreadable flags return
// a printer capable of printing with a "namespace" column.
func (f *PrintFlags) EnsureWithNamespace() {
	f.HumanReadableFlags.EnsureWithNamespace()
}
