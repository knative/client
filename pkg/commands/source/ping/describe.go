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

package ping

import (
	"errors"
	"sort"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"knative.dev/client/pkg/describe"

	"knative.dev/client-pkg/pkg/printers"
	"knative.dev/client/pkg/commands"
	clientsourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"
)

var describeExample = `
  # Describe a ping source 'myping'
  kn source ping describe myping

  # Describe a ping source 'myping' in YAML format
  kn source ping describe myping -o yaml`

// NewPingDescribeCommand returns a new command for describe a Ping source object
func NewPingDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:               "describe NAME",
		Short:             "Show details of a ping source",
		Example:           describeExample,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source ping describe' requires name of the source as single argument")
			}
			name := args[0]

			pingSourceClient, err := newPingSourceClient(p, cmd)
			if err != nil {
				return err
			}

			pingSource, err := pingSourceClient.GetPingSource(cmd.Context(), name)
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
				return printer.PrintObj(pingSource, out)
			}
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writePingSource(dw, pingSource, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Revisions summary info
			describe.Sink(dw, "Sink", pingSource.Namespace, &pingSource.Spec.Sink)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			if pingSource.Spec.CloudEventOverrides != nil && pingSource.Spec.CloudEventOverrides.Extensions != nil {
				writeCeOverrides(dw, pingSource.Spec.CloudEventOverrides.Extensions)
				dw.WriteLine()
				if err := dw.Flush(); err != nil {
					return err
				}
			}

			// Condition info
			commands.WriteConditions(dw, pingSource.Status.Conditions, printDetails)
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

func writePingSource(dw printers.PrefixWriter, source *clientsourcesv1beta2.PingSource, printDetails bool) {
	commands.WriteMetadata(dw, &source.ObjectMeta, printDetails)
	dw.WriteAttribute("Schedule", source.Spec.Schedule)
	if source.Spec.DataBase64 != "" {
		dw.WriteAttribute("DataBase64", source.Spec.DataBase64)
	} else {
		dw.WriteAttribute("Data", source.Spec.Data)
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
