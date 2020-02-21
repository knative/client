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
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewBindingListCommand is for listing sink bindings
func NewBindingListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(BindingListHandlers)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sink bindings.",
		Example: `
  # List all sink binidngs
  kn source binding list

  # List all sink bindings in YAML format
  kn source binding list -o yaml`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// TODO: filter list by given source name

			bindingClient, err := newSinkBindingClient(p, cmd)
			if err != nil {
				return err
			}

			sourceList, err := bindingClient.ListSinkBindings()
			if err != nil {
				return err
			}

			if len(sourceList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No sink binding found.\n")
				return nil
			}

			if bindingClient.Namespace() == "" {
				listFlags.EnsureWithNamespace()
			}

			printer, err := listFlags.ToPrinter()
			if err != nil {
				return nil
			}

			err = printer.PrintObj(sourceList, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), true)
	listFlags.AddFlags(cmd)
	return cmd
}
