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
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewTriggerListCommand represents 'kn trigger list' command
func NewTriggerListCommand(p *commands.KnParams) *cobra.Command {
	triggerListFlags := flags.NewListPrintFlags(TriggerListHandlers)

	triggerListCommand := &cobra.Command{
		Use:     "list",
		Short:   "List triggers",
		Aliases: []string{"ls"},
		Example: `
  # List all triggers
  kn trigger list

  # List all triggers in JSON output format
  kn trigger list -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}
			triggerList, err := client.ListTriggers()
			if err != nil {
				return err
			}
			if len(triggerList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No triggers found.\n")
				return nil
			}

			// empty namespace indicates all-namespaces flag is specified
			if namespace == "" {
				triggerListFlags.EnsureWithNamespace()
			}

			err = triggerListFlags.Print(triggerList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(triggerListCommand.Flags(), true)
	triggerListFlags.AddFlags(triggerListCommand)
	return triggerListCommand
}
