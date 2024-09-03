/*
Copyright 2022 The Knative Authors

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
	"knative.dev/client/pkg/config"
	v1 "knative.dev/client/pkg/eventing/v1"
	duckv1 "knative.dev/eventing/pkg/apis/duck/v1"

	"knative.dev/client/pkg/commands"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

var updateExample = `
  # Update a broker 'mybroker' in the current namespace with delivery sink svc1
  kn broker update mybroker --dl-sink svc1

  # Update a broker 'mybroker' in the 'myproject' namespace and with retry 2 seconds
  kn broker update mybroker --namespace myproject --retry 2
`

func NewBrokerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var deliveryFlags DeliveryOptionFlags

	cmd := &cobra.Command{
		Use:     "update NAME",
		Short:   "Update a broker",
		Example: updateExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'broker update' requires the broker name given as single argument")
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

			updateFunc := func(origBroker *eventingv1.Broker) (*eventingv1.Broker, error) {
				b := v1.NewBrokerBuilderFromExisting(origBroker)
				if cmd.Flags().Changed("dl-sink") {
					destination, err := deliveryFlags.GetDlSink(cmd, dynamicClient, namespace)
					if err != nil {
						return nil, err
					}
					b.DlSink(destination)
				}
				if cmd.Flags().Changed("retry") {
					b.Retry(&deliveryFlags.RetryCount)
				}
				if cmd.Flags().Changed("timeout") {
					b.Timeout(&deliveryFlags.Timeout)
				}
				if cmd.Flags().Changed("backoff-policy") {
					backoffPolicy := duckv1.BackoffPolicyType(deliveryFlags.BackoffPolicy)
					b.BackoffPolicy(&backoffPolicy)
				}
				if cmd.Flags().Changed("backoff-delay") {
					b.BackoffDelay(&deliveryFlags.BackoffDelay)
				}
				if cmd.Flags().Changed("retry-after-max") {
					b.RetryAfterMax(&deliveryFlags.RetryAfterMax)
				}
				return b.Build(), nil
			}
			err = eventingClient.UpdateBrokerWithRetry(cmd.Context(), name, updateFunc, config.DefaultRetry.Steps)
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Broker '%s' updated in namespace '%s'.\n", name, namespace)
			}
			return err
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preCheck(cmd)
		}}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	deliveryFlags.Add(cmd)
	return cmd
}

func preCheck(cmd *cobra.Command) error {
	if cmd.Flags().NFlag() == 0 {
		return fmt.Errorf("flag(s) not set\nUsage: %s", cmd.Use)
	}

	return nil
}
