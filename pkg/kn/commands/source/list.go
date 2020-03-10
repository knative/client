// Copyright © 2020 The Knative Authors
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

package source

import (
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	"knative.dev/client/pkg/kn/commands/source/duck"
)

var listExample = `
  # List available eventing sources
  kn source list

  # List PingSource type sources
  kn source list --type=PingSource

  # List PingSource and ApiServerSource types sources
  kn source list --type=PingSource --type=apiserversource`

// NewListCommand defines and processes `kn source list`
func NewListCommand(p *commands.KnParams) *cobra.Command {
	filterFlags := &flags.SourceTypeFilters{}
	listFlags := flags.NewListPrintFlags(ListHandlers)
	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "List available sources",
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}
			var filters dynamic.WithTypes
			for _, filter := range filterFlags.Filters {
				filters = append(filters, dynamic.WithTypeFilter(filter))
			}
			sourceList, err := dynamicClient.ListSources(filters...)
			if err != nil {
				return err
			}
			if len(sourceList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No sources found in %s namespace.\n", namespace)
				return nil
			}
			// empty namespace indicates all namespaces flag is specified
			if namespace == "" {
				listFlags.EnsureWithNamespace()
			}
			printer, err := listFlags.ToPrinter()
			if err != nil {
				return nil
			}
			if listFlags.GenericPrintFlags.OutputFlagSpecified() {
				return printer.PrintObj(sourceList, cmd.OutOrStdout())
			}
			// Convert the source list to DuckSourceList only if human readable table printing requested
			sourceDuckList := duck.ToSourceList(sourceList)
			err = printer.PrintObj(sourceDuckList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(listCommand.Flags(), true)
	listFlags.AddFlags(listCommand)
	filterFlags.Add(listCommand, "source type")
	return listCommand
}
