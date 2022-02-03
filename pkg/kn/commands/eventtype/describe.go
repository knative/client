/*
Copyright 2022 The Knative Authors

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

package eventtype

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

var describeExample = `
  # Describe eventtype 'myeventtype' in the current namespace
  kn eventtype describe myeventtype

  # Describe eventtype 'myeventtype' in the 'myproject' namespace
  kn eventtype describe myeventtype --namespace myproject

  # Describe eventtype 'myeventtype' in YAML format
  kn eventtype describe myeventtype -o yaml`

// NewEventtypeDescribeCommand represents command to describe the details of an eventtype instance
func NewEventtypeDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	cmd := &cobra.Command{
		Use:     "describe",
		Short:   "Describe eventtype",
		Example: describeExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'broker describe' requires the broker name given as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			eventingV1Beta1Client, err := p.NewEventingV1beta1Client(namespace)
			if err != nil {
				return err
			}

			eventtype, err := eventingV1Beta1Client.GetEventtype(cmd.Context(), name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()

			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(eventtype, out)
			}
			return describeEventtype(out, eventtype, false)
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	machineReadablePrintFlags.AddFlags(cmd)
	cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(machineReadablePrintFlags.AllowedFormats(), "|"))
	return cmd
}

// describeEventtype prints eventtype details to the provided output writer
func describeEventtype(out io.Writer, eventtype *eventingv1beta1.EventType, printDetails bool) error {
	dw := printers.NewPrefixWriter(out)
	commands.WriteMetadata(dw, &eventtype.ObjectMeta, printDetails)
	dw.WriteLine()
	dw.WriteLine()
	commands.WriteConditions(dw, eventtype.Status.Conditions, printDetails)
	if err := dw.Flush(); err != nil {
		return err
	}
	return nil
}
