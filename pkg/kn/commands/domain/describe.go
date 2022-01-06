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

package domain

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// NewDomainMappingDescribeCommand represents 'kn route describe' command
func NewDomainMappingDescribeCommand(p *commands.KnParams) *cobra.Command {
	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")
	cmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a domain mapping",
		Example: `
  # Show details of for the domain 'hello.example.com'
  kn domain describe hello.example.com`,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn domain describe' requires name of the domain mapping as single argument")
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewServingV1alpha1Client(namespace)
			if err != nil {
				return err
			}

			domainMapping, err := client.GetDomainMapping(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if machineReadablePrintFlags.OutputFlagSpecified() {
				if strings.ToLower(*machineReadablePrintFlags.OutputFormat) == "url" {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", domainMapping.Status.URL)
					return nil
				}
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(domainMapping, cmd.OutOrStdout())
			}
			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			return describe(cmd.OutOrStdout(), domainMapping, printDetails)
		},
	}
	flags := cmd.Flags()
	commands.AddNamespaceFlags(flags, false)
	machineReadablePrintFlags.AddFlags(cmd)
	cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(append(machineReadablePrintFlags.AllowedFormats(), "url"), "|"))
	flags.BoolP("verbose", "v", false, "More output.")
	return cmd
}

func describe(w io.Writer, domainMapping *v1alpha1.DomainMapping, printDetails bool) error {
	dw := printers.NewPrefixWriter(w)
	commands.WriteMetadata(dw, &domainMapping.ObjectMeta, printDetails)
	dw.WriteLine()
	dw.WriteAttribute("URL", domainMapping.Status.URL.String())
	dw.WriteLine()
	ref := dw.WriteAttribute("Reference", "")
	ref.WriteAttribute("APIVersion", domainMapping.Spec.Ref.APIVersion)
	ref.WriteAttribute("Kind", domainMapping.Spec.Ref.Kind)
	ref.WriteAttribute("Name", domainMapping.Spec.Ref.Name)
	if domainMapping.Namespace != domainMapping.Spec.Ref.Namespace {
		ref.WriteAttribute("Namespace", domainMapping.Spec.Ref.Namespace)
	}
	dw.WriteLine()
	commands.WriteConditions(dw, domainMapping.Status.Conditions, printDetails)
	if err := dw.Flush(); err != nil {
		return err
	}
	return nil
}
