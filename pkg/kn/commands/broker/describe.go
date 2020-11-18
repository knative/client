/*
Copyright 2020 The Knative Authors

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

package broker

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/spf13/cobra"

	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

var describeExample = `
  # Describe broker 'mybroker' in the current namespace
  kn broker describe mybroker

  # Describe broker 'mybroker' in the 'myproject' namespace
  kn broker describe mybroker --namespace myproject

  # Describe broker 'mybroker' in YAML format
  kn broker describe mybroker -o yaml

  # Print only broker URL
  kn broker describe mybroker -o url`

// NewBrokerDescribeCommand represents command to describe details of broker instance
func NewBrokerDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	cmd := &cobra.Command{
		Use:     "describe NAME",
		Short:   "Describe broker",
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

			eventingClient, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			broker, err := eventingClient.GetBroker(name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()

			if machineReadablePrintFlags.OutputFlagSpecified() {
				if strings.ToLower(*machineReadablePrintFlags.OutputFormat) == "url" {
					fmt.Fprintf(out, "%s\n", extractURL(broker))
					return nil
				}
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(broker, out)
			}
			return describeBroker(out, broker, false)
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	machineReadablePrintFlags.AddFlags(cmd)
	cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(append(machineReadablePrintFlags.AllowedFormats(), "url"), "|"))
	return cmd
}

// describeBroker print broker details to the provided output writer
func describeBroker(out io.Writer, broker *v1beta1.Broker, printDetails bool) error {
	dw := printers.NewPrefixWriter(out)
	commands.WriteMetadata(dw, &broker.ObjectMeta, printDetails)
	dw.WriteLine()
	dw.WriteAttribute("Address", "").WriteAttribute("URL", extractURL(broker))
	dw.WriteLine()
	commands.WriteConditions(dw, broker.Status.Conditions, printDetails)
	if err := dw.Flush(); err != nil {
		return err
	}
	return nil
}

func extractURL(broker *v1beta1.Broker) string {
	return broker.Status.Address.URL.String()
}
