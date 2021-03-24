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

package apiserver

import (
	context2 "context"
	"errors"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"knative.dev/eventing/pkg/apis/sources/v1alpha2"

	"knative.dev/client/lib/printing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

var describeExample = `
  # Describe an api-server source with name 'k8sevents'
  kn source apiserver describe k8sevents

  # Describe an api-server source with name 'k8sevents' in YAML format
  kn source apiserver describe k8sevents -o yaml`

// NewAPIServerDescribeCommand to describe an ApiServer source object
func NewAPIServerDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:     "describe NAME",
		Short:   "Show details of an api-server source",
		Example: describeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source apiserver describe' requires name of the source as single argument")
			}
			name := args[0]

			apiSourceClient, err := newAPIServerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			apiSource, err := apiSourceClient.GetAPIServerSource(context2.TODO(), name)
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
				return printer.PrintObj(apiSource, out)
			}
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeAPIServerSource(dw, apiSource, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			printing.DescribeSink(dw, "Sink", apiSource.Namespace, &apiSource.Spec.Sink)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			if apiSource.Spec.CloudEventOverrides != nil && apiSource.Spec.CloudEventOverrides.Extensions != nil {
				writeCeOverrides(dw, apiSource.Spec.CloudEventOverrides.Extensions)
			}

			writeResources(dw, apiSource.Spec.Resources)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, apiSource.Status.Conditions, printDetails)
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

func writeResources(dw printers.PrefixWriter, apiVersionKindSelectors []v1alpha2.APIVersionKindSelector) {
	subWriter := dw.WriteAttribute("Resources", "")
	for _, resource := range apiVersionKindSelectors {
		subWriter.WriteAttribute("Kind", fmt.Sprintf("%s (%s)", resource.Kind, resource.APIVersion))
		if resource.LabelSelector != nil {
			subWriter.WriteAttribute("Selector", labelSelectorToString(resource.LabelSelector))
		}
	}
}

func writeAPIServerSource(dw printers.PrefixWriter, source *v1alpha2.ApiServerSource, printDetails bool) {
	commands.WriteMetadata(dw, &source.ObjectMeta, printDetails)
	dw.WriteAttribute("ServiceAccountName", source.Spec.ServiceAccountName)
	dw.WriteAttribute("EventMode", source.Spec.EventMode)
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
