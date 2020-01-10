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

package revision

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"knative.dev/client/pkg/kn/commands"
	hprinters "knative.dev/client/pkg/printers"
)

// RevisionListFlags composes common printer flag structs
// used in the List command.
type RevisionListFlags struct {
	GenericPrintFlags  *genericclioptions.PrintFlags
	HumanReadableFlags *commands.HumanPrintFlags
	ServiceRefFlags    *ServiceReferenceFlags
}

// AllowedFormats is the list of formats in which data can be displayed
func (f *RevisionListFlags) AllowedFormats() []string {
	formats := f.GenericPrintFlags.AllowedFormats()
	formats = append(formats, f.HumanReadableFlags.AllowedFormats()...)
	return formats
}

// ToPrinter attempts to find a composed set of RevisionListFlags suitable for
// returning a printer based on current flag values.
func (f *RevisionListFlags) ToPrinter() (hprinters.ResourcePrinter, error) {
	// if there are flags specified for generic printing
	if f.GenericPrintFlags.OutputFlagSpecified() {
		// we need to wrap for cleaning up any temporary annotations
		return wrapPrinterForAnnotationCleanup(f.GenericPrintFlags.ToPrinter())
	}
	// if no flags specified, use the table printing
	p, err := f.HumanReadableFlags.ToPrinter(RevisionListHandlers)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Flags receives a *cobra.Command reference and binds
// flags related to humanreadable and template printing
// as well as to reference a service
func (f *RevisionListFlags) AddFlags(cmd *cobra.Command) {
	f.GenericPrintFlags.AddFlags(cmd)
	f.HumanReadableFlags.AddFlags(cmd)
	f.ServiceRefFlags.SetOptional(cmd)
}

// NewRevisionListFlags returns flags associated with humanreadable,
// template, and "name" printing, with default values set.
func NewRevisionListFlags() *RevisionListFlags {
	return &RevisionListFlags{
		GenericPrintFlags:  genericclioptions.NewPrintFlags(""),
		HumanReadableFlags: commands.NewHumanPrintFlags(),
		ServiceRefFlags:    &ServiceReferenceFlags{},
	}
}

// ServiceReferenceFlags compose a set of flag(s) to reference a service
type ServiceReferenceFlags struct {
	Name string
}

// SetOptional receives a *cobra.Command reference and
// adds the ServiceReferenceFlags flags as optional flags
func (s *ServiceReferenceFlags) SetOptional(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.Name, "service", "s", "", "Service name")
}
