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
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

// NewTriggerDeleteCommand represent 'revision delete' command
func NewTriggerDeleteCommand(p *commands.KnParams) *cobra.Command {
	TriggerDeleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a trigger",
		Example: `
  # Delete a trigger 'mytrigger' in default namespace
  kn trigger delete mytrigger`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'trigger delete' requires the name of the trigger as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			eventingClient, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			err = eventingClient.DeleteTrigger(name)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Trigger '%s' deleted in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(TriggerDeleteCommand.Flags(), false)
	return TriggerDeleteCommand
}
