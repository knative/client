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

	"github.com/spf13/cobra"

	clientv1beta1 "knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/kn/commands"
)

var create_example = `
# Create a broker 'mybroker' in the current namespace
  kn broker create mybroker
# Create a broker 'mybroker' in the 'myproject' namespace
  kn broker create mybroker --namespace myproject
`

func NewBrokerCreateCommand(p *commands.KnParams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "create NAME",
		Short:   "Create a broker.",
		Example: create_example,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'broker create' requires the broker name given as single argument")
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

			brokerBuilder := clientv1beta1.
				NewBrokerBuilder(name).
				Namespace(namespace)

			err = eventingClient.CreateBroker(brokerBuilder.Build())
			if err != nil {
				return fmt.Errorf(
					"cannot create broker '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Broker '%s' successfully created in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	return cmd
}
