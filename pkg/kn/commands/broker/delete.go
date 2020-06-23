/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package broker

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
)

var deleteExample = `
  # Delete a broker 'mybroker' in the current namespace
  kn broker create mybroker

  # Delete a broker 'mybroker' in the 'myproject' namespace
  kn broker create mybroker --namespace myproject`

// NewBrokerDeleteCommand represents command to existing delete broker
func NewBrokerDeleteCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags

	cmd := &cobra.Command{
		Use:     "delete NAME",
		Short:   "Delete a broker",
		Example: deleteExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'broker delete' requires the broker name given as single argument")
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

			timeout := time.Duration(0)
			if waitFlags.Wait {
				timeout = time.Duration(waitFlags.TimeoutInSeconds) * time.Second
			}
			err = eventingClient.DeleteBroker(name, timeout)
			if err != nil {
				return fmt.Errorf(
					"cannot delete broker '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Broker '%s' successfully deleted in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	waitFlags.AddConditionWaitFlags(cmd, commands.WaitDefaultTimeout, "delete", "broker", "deleted")
	return cmd
}
