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
	v1 "knative.dev/eventing/pkg/apis/duck/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	clientv1beta1 "knative.dev/client-pkg/pkg/eventing/v1"
	"knative.dev/client/pkg/commands"
)

var createExample = `
  # Create a broker 'mybroker' in the current namespace
  kn broker create mybroker

  # Create a broker 'mybroker' in the 'myproject' namespace and with a broker class of 'Kafka'
  kn broker create mybroker --namespace myproject --class Kafka

  # Create a broker 'mybroker' in the myproject namespace with config referencing a configmap in current namespace
  kn broker create mybroker --namespace myproject --class Kafka --broker-config cm:spec-cm
  OR
  kn broker create mybroker --namespace myproject --class Kafka --broker-config spec-cm

  # Create a broker 'mybroker' in the myproject namespace with config referencing secret named spec-sc in test namespace
  kn broker create mybroker --namespace myproject --class Kafka --broker-config sc:spec-sc:test

  # Create a broker 'mybroker' in the myproject namespace with config referencing RabbitmqCluster mycluster in test namespace
  kn broker create mybroker --namespace myproject --class Kafka --broker-config rabbitmq.com/v1beta1:RabbitmqCluster:mycluster:test
`

// NewBrokerCreateCommand represents command to create new broker instance
func NewBrokerCreateCommand(p *commands.KnParams) *cobra.Command {

	var className string

	var deliveryFlags DeliveryOptionFlags
	var configFlags ConfigFlags
	cmd := &cobra.Command{
		Use:     "create NAME",
		Short:   "Create a broker",
		Example: createExample,
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

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			destination, err := deliveryFlags.GetDlSink(cmd, dynamicClient, namespace)
			if err != nil {
				return err
			}

			backoffPolicy := v1.BackoffPolicyType(deliveryFlags.BackoffPolicy)

			var configReference *duckv1.KReference

			if cmd.Flags().Changed("broker-config") {
				if !cmd.Flags().Changed("class") {
					return fmt.Errorf("cannot set broker-config without setting class")
				}

				configReference, err = configFlags.GetBrokerConfigReference()
				if err != nil {
					return err
				}
			}

			brokerBuilder := clientv1beta1.
				NewBrokerBuilder(name).
				Namespace(namespace).
				Class(className).
				DlSink(destination).
				Retry(&deliveryFlags.RetryCount).
				Timeout(&deliveryFlags.Timeout).
				BackoffPolicy(&backoffPolicy).
				BackoffDelay(&deliveryFlags.BackoffDelay).
				RetryAfterMax(&deliveryFlags.RetryAfterMax).
				Config(configReference)

			err = eventingClient.CreateBroker(cmd.Context(), brokerBuilder.Build())
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
	cmd.Flags().StringVar(&className, "class", "", "Broker class like 'MTChannelBasedBroker' or 'Kafka' (if available).")
	configFlags.Add(cmd)
	deliveryFlags.Add(cmd)
	return cmd
}
