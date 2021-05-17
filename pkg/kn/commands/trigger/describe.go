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

package trigger

import (
	"errors"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"knative.dev/client/lib/printing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

var describeExample = `
  # Describe a trigger with name 'my-trigger'
  kn trigger describe my-trigger

  # Describe a trigger 'my-trigger' in YAML format
  kn trigger describe my-trigger -o yaml`

// NewTriggerDescribeCommand returns a new command for describe a trigger
func NewTriggerDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:     "describe NAME",
		Short:   "Show details of a trigger",
		Example: describeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn trigger describe' requires name of the trigger as single argument")
			}
			name := args[0]

			// get namespace
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			// get client
			eventingClient, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			trigger, err := eventingClient.GetTrigger(cmd.Context(), name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()

			// Print out machine readable output if requested
			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(trigger, out)
			}

			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeTrigger(dw, trigger, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Revisions summary info
			printing.DescribeSink(dw, "Sink", trigger.Namespace, &trigger.Spec.Subscriber)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, trigger.Status.Conditions, printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")
	machineReadablePrintFlags.AddFlags(command)
	return command
}

func writeTrigger(dw printers.PrefixWriter, trigger *v1beta1.Trigger, printDetails bool) {
	commands.WriteMetadata(dw, &trigger.ObjectMeta, printDetails)
	dw.WriteAttribute("Broker", trigger.Spec.Broker)
	if trigger.Spec.Filter != nil && trigger.Spec.Filter.Attributes != nil {
		subWriter := dw.WriteAttribute("Filter", "")
		for key, value := range trigger.Spec.Filter.Attributes {
			subWriter.WriteAttribute(key, value)
		}
	}
}
