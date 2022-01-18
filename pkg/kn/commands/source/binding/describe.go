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

package binding

import (
	"errors"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"
	"knative.dev/pkg/tracker"

	"knative.dev/client/lib/printing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

var describeExample = `
  # Describe a sink binding 'mysinkbinding'
  kn source binding describe mysinkbinding

  # Describe a sink binding 'mysinkbinding' in YAML format
  kn source binding describe mysinkbinding -o yaml`

// NewBindingDescribeCommand returns a new command for describe a sink binding object
func NewBindingDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:               "describe NAME",
		Short:             "Show details of a sink binding",
		Example:           describeExample,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source binding describe' requires name of the sink binding as single argument")
			}
			name := args[0]

			bindingClient, err := newSinkBindingClient(p, cmd)
			if err != nil {
				return err
			}

			binding, err := bindingClient.GetSinkBinding(cmd.Context(), name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			// Print out machine readable output if requested
			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(binding, out)
			}

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeSinkBinding(dw, binding, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, binding.Status.Conditions, printDetails)
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

func writeSinkBinding(dw printers.PrefixWriter, binding *v1.SinkBinding, printDetails bool) {
	commands.WriteMetadata(dw, &binding.ObjectMeta, printDetails)
	writeSubject(dw, binding.Namespace, &binding.Spec.Subject)
	printing.DescribeSink(dw, "Sink", binding.Namespace, &binding.Spec.Sink)
	if binding.Spec.CloudEventOverrides != nil && binding.Spec.CloudEventOverrides.Extensions != nil {
		writeCeOverrides(dw, binding.Spec.CloudEventOverrides.Extensions)
	}
}

func writeCeOverrides(dw printers.PrefixWriter, ceOverrides map[string]string) {
	subDw := dw.WriteAttribute("CloudEvent Overrides", "")
	keys := make([]string, 0, len(ceOverrides))
	for k := range ceOverrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		subDw.WriteAttribute(k, ceOverrides[k])
	}
}

func writeSubject(dw printers.PrefixWriter, namespace string, subject *tracker.Reference) {
	subjectDw := dw.WriteAttribute("Subject", "")
	if subject.Namespace != "" && subject.Namespace != namespace {
		subjectDw.WriteAttribute("Namespace", subject.Namespace)
	}
	if subject.Name != "" {
		subjectDw.WriteAttribute("Name", subject.Name)
	}
	subjectDw.WriteAttribute("Resource", fmt.Sprintf("%s (%s)", subject.Kind, subject.APIVersion))
	if subject.Selector != nil {
		matchDw := subjectDw.WriteAttribute("Selector", "")
		selector := subject.Selector
		if len(selector.MatchLabels) > 0 {
			var lKeys []string
			for k := range selector.MatchLabels {
				lKeys = append(lKeys, k)
			}
			sort.Strings(lKeys)
			for _, k := range lKeys {
				matchDw.WriteAttribute(k, selector.MatchLabels[k])
			}
		}
		// TODO: Print out selector.MatchExpressions
	}
}
