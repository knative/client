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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

// NewChannelDeleteCommand is for deleting a Channel
func NewChannelDeleteCommand(p *commands.KnParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a channel",
		Example: `
  # Delete a channel 'pipe'
  kn channel delete pipe`,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn channel delete' requires the channel name as single argument")
			}
			name := args[0]

			channelClient, err := newChannelClient(p, cmd)
			if err != nil {
				return err
			}

			err = channelClient.DeleteChannel(cmd.Context(), name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Channel '%s' deleted in namespace '%s'.\n", name, channelClient.Namespace())
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	return cmd
}
