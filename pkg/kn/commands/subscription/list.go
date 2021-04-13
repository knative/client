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

package subscription

import (
	"fmt"

	"knative.dev/client/pkg/util"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewSubscriptionListCommand is for listing subscription objects
func NewSubscriptionListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(ListHandlers)

	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List subscriptions",
		Example: `
  # List all subscriptions
  kn subscription list

  # List subscriptions in YAML format
  kn subscription list -o yaml`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// TODO: filter list by given subscription name

			client, err := newSubscriptionClient(p, cmd)
			if err != nil {
				return err
			}

			subscriptionList, err := client.ListSubscription(cmd.Context())
			if err != nil {
				return err
			}

			if subscriptionList == nil {
				subscriptionList = &messagingv1.SubscriptionList{}
				err := util.UpdateGroupVersionKindWithScheme(subscriptionList, messagingv1.SchemeGroupVersion, scheme.Scheme)
				if err != nil {
					return err
				}
			}
			if !listFlags.GenericPrintFlags.OutputFlagSpecified() && len(subscriptionList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No subscriptions found.\n")
				return nil
			}

			if client.Namespace() == "" {
				listFlags.EnsureWithNamespace()
			}

			err = listFlags.Print(subscriptionList, cmd.OutOrStdout())
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
