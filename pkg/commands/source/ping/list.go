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
	"fmt"

	"github.com/spf13/cobra"
	listfl "knative.dev/client-pkg/pkg/commands/flags/list"

	"knative.dev/client/pkg/commands"
)

// NewPingListCommand is for listing Ping source COs
func NewPingListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := listfl.NewPrintFlags(PingSourceListHandlers)

	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List ping sources",
		Example: `
  # List all Ping sources
  kn source ping list

  # List all Ping sources in YAML format
  kn source ping list -o yaml`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// TODO: filter list by given source name

			pingClient, err := newPingSourceClient(p, cmd)
			if err != nil {
				return err
			}

			sourceList, err := pingClient.ListPingSource(cmd.Context())
			if err != nil {
				return err
			}

			if !listFlags.GenericPrintFlags.OutputFlagSpecified() && len(sourceList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No Ping source found.\n")
				return nil
			}

			if pingClient.Namespace() == "" {
				listFlags.EnsureWithNamespace()
			}

			err = listFlags.Print(sourceList, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(listCommand.Flags(), true)
	listFlags.AddFlags(listCommand)
	return listCommand
}
