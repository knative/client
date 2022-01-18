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
	"errors"
	"fmt"

	"knative.dev/client/pkg/config"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"

	"github.com/spf13/cobra"

	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	knmessagingv1 "knative.dev/client/pkg/messaging/v1"
)

// NewSubscriptionUpdateCommand to update event subscriptions
func NewSubscriptionUpdateCommand(p *commands.KnParams) *cobra.Command {
	var subscriberFlag, replyFlag, dlsFlag flags.SinkFlags
	cmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update an event subscription",
		Example: `
  # Update a subscription 'sub0' with a subscriber ksvc 'receiver'
  kn subscription update sub0 --sink ksvc:receiver

  # Update a subscription 'sub1' with subscriber ksvc 'mirror', reply to a broker 'nest' and DeadLetterSink to a ksvc 'bucket'
  kn subscription update sub1 --sink mirror --sink-reply broker:nest --sink-dead-letter bucket`,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn subscription update' requires the subscription name given as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			client, err := newSubscriptionClient(p, cmd)
			if err != nil {
				return err
			}

			updateFunc := func(origSub *messagingv1.Subscription) (*messagingv1.Subscription, error) {
				sb := knmessagingv1.NewSubscriptionBuilderFromExisting(origSub)

				sub, err := subscriberFlag.ResolveSink(cmd.Context(), dynamicClient, namespace)
				if err != nil {
					return nil, err
				}
				sb.Subscriber(sub)

				rep, err := replyFlag.ResolveSink(cmd.Context(), dynamicClient, namespace)
				if err != nil {
					return nil, err
				}
				sb.Reply(rep)

				ds, err := dlsFlag.ResolveSink(cmd.Context(), dynamicClient, namespace)
				if err != nil {
					return nil, err
				}
				sb.DeadLetterSink(ds)
				return sb.Build(), nil
			}
			err = client.UpdateSubscriptionWithRetry(cmd.Context(), name, updateFunc, config.DefaultRetry.Steps)
			if err != nil {
				return knerrors.GetError(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Subscription '%s' updated in namespace '%s'.\n", name, namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	// add subscriber flag as `--sink`
	subscriberFlag.Add(cmd)
	replyFlag.AddWithFlagName(cmd, "sink-reply", "")
	dlsFlag.AddWithFlagName(cmd, "sink-dead-letter", "")
	return cmd
}
