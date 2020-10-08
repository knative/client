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

package channel

import (
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewChannelListCommand is for listing channel objects
func NewChannelListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(ListHandlers)

	listCommand := &cobra.Command{
		Use:     "list",
		Short:   "List channels",
		Aliases: []string{"ls"},
		Example: `
  # List all channels
  kn channel list

  # List channels in YAML format
  kn channel ping list -o yaml`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// TODO: filter list by given channel name

			client, err := newChannelClient(p, cmd)
			if err != nil {
				return err
			}

			channelList, err := client.ListChannel()
			if err != nil {
				return err
			}

			if channelList == nil || len(channelList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No channels found.\n")
				return nil
			}

			if client.Namespace() == "" {
				listFlags.EnsureWithNamespace()
			}

			err = listFlags.Print(channelList, cmd.OutOrStdout())
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
