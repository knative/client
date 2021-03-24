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
	context2 "context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

// NewSubscriptionDeleteCommand is for deleting a Subscription
func NewSubscriptionDeleteCommand(p *commands.KnParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a subscription",
		Example: `
  # Delete a subscription 'sub0'
  kn subscription delete sub0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn subscription delete' requires the subscription name as single argument")
			}
			name := args[0]

			subscriptionClient, err := newSubscriptionClient(p, cmd)
			if err != nil {
				return err
			}

			err = subscriptionClient.DeleteSubscription(context2.TODO(), name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Subscription '%s' deleted in namespace '%s'.\n", name, subscriptionClient.Namespace(context2.TODO()))
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	return cmd
}
