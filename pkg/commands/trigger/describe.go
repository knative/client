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
	"knative.dev/client-pkg/pkg/printers"
	"knative.dev/client/pkg/commands"
	"knative.dev/client/pkg/describe"
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
		Use:               "describe NAME",
		Short:             "Show details of a trigger",
		Example:           describeExample,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
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
			describe.Sink(dw, "Sink", trigger.Namespace, &trigger.Spec.Subscriber)
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
	if len(trigger.Spec.Filters) > 0 {
		// Split 'Filter' and 'Filters (experimental)' with new line
		dw.WriteLine()
		subWriter := dw.WriteAttribute("Filters (experimental)", "")
		for _, filter := range trigger.Spec.Filters {
			writeNestedFilters(subWriter, filter)
		}
	}
}

// writeNestedFilters goes through SubscriptionsAPIFilter and writes its content accordingly
func writeNestedFilters(dw printers.PrefixWriter, filter v1beta1.SubscriptionsAPIFilter) {
	// All []SubscriptionsAPIFilter
	if len(filter.All) > 0 {
		// create new indentation after name
		subWriter := dw.WriteAttribute("all", "")
		for _, nestedFilter := range filter.All {
			writeNestedFilters(subWriter, nestedFilter)
		}
	}
	// Any []SubscriptionsAPIFilter
	if len(filter.Any) > 0 {
		// create new indentation after name
		subWriter := dw.WriteAttribute("any", "")
		for _, nestedFilter := range filter.Any {
			writeNestedFilters(subWriter, nestedFilter)
		}
	}
	// Not *SubscriptionsAPIFilter
	if filter.Not != nil {
		subWriter := dw.WriteAttribute("not", "")
		writeNestedFilters(subWriter, *filter.Not)
	}
	// Exact map[string]string
	if len(filter.Exact) > 0 {
		// create new indentation after name
		subWriter := dw.WriteAttribute("exact", "")
		for k, v := range filter.Exact {
			subWriter.WriteAttribute(k, v)
		}
	}
	// Prefix map[string]string
	if len(filter.Prefix) > 0 {
		// create new indentation after name
		subWriter := dw.WriteAttribute("prefix", "")
		for k, v := range filter.Prefix {
			subWriter.WriteAttribute(k, v)
		}
	}
	// Suffix map[string]string
	if len(filter.Suffix) > 0 {
		// create new indentation after name
		subWriter := dw.WriteAttribute("suffix", "")
		for k, v := range filter.Suffix {
			subWriter.WriteAttribute(k, v)
		}
	}
	// CESQL string
	if filter.CESQL != "" {
		dw.WriteAttribute("cesql", filter.CESQL)
	}
}
