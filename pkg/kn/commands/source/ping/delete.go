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
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

// NewPingDeleteCommand is for deleting a Ping source
func NewPingDeleteCommand(p *commands.KnParams) *cobra.Command {
	pingDeleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a Ping source.",
		Example: `
  # Delete a Ping source 'my-ping'
  kn source ping delete my-ping`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'requires the name of the Ping source to delete as single argument")
			}
			name := args[0]

			pingClient, err := newPingSourceClient(p, cmd)
			if err != nil {
				return err
			}

			err = pingClient.DeletePingSource(name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Ping source '%s' deleted in namespace '%s'.\n", name, pingClient.Namespace())
			return nil
		},
	}
	commands.AddNamespaceFlags(pingDeleteCommand.Flags(), false)
	return pingDeleteCommand
}
