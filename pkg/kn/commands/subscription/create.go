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

	"github.com/spf13/cobra"

	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	knflags "knative.dev/client/pkg/kn/flags"
	knmessagingv1beta1 "knative.dev/client/pkg/messaging/v1beta1"
)

// NewSubscriptionCreateCommand to create event subscriptions
func NewSubscriptionCreateCommand(p *commands.KnParams) *cobra.Command {
	var (
		crefFlag                           knflags.ChannelRef
		subscriberFlag, replyFlag, dlsFlag flags.SinkFlags
	)

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a subscription",
		Example: `
  # Create a subscription 'sub0' from InMemoryChannel 'pipe0' to a subscriber ksvc 'receiver'
  kn subscription create sub0 --channel imcv1beta1:pipe0 --sink ksvc:receiver

  # Create a subscription 'sub1' from KafkaChannel 'k1' to ksvc 'mirror', reply to a broker 'nest' and DeadLetterSink to a ksvc 'bucket'
  kn subscription create sub1 --channel messaging.knative.dev:v1alpha1:KafkaChannel:k1 --sink mirror --sink-reply broker:nest --sink-dead-letter bucket`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn subscription create' requires the subscription name given as single argument")
			}
			name := args[0]

			if crefFlag.Cref == "" {
				return errors.New("'kn subscription create' requires the channel reference provided with --channel flag")
			}

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

			sb := knmessagingv1beta1.NewSubscriptionBuilder(name)

			cref, err := crefFlag.Parse()
			if err != nil {
				return err
			}
			sb.Channel(cref)

			sub, err := subscriberFlag.ResolveSink(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return err
			}
			sb.Subscriber(sub)

			rep, err := replyFlag.ResolveSink(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return err
			}
			sb.Reply(rep)

			ds, err := dlsFlag.ResolveSink(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return err
			}
			sb.DeadLetterSink(ds)

			err = client.CreateSubscription(sb.Build())
			if err != nil {
				return knerrors.GetError(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Subscription '%s' created in namespace '%s'.\n", name, namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	crefFlag.Add(cmd.Flags())
	// add subscriber flag as `--sink`
	subscriberFlag.Add(cmd)
	replyFlag.AddWithFlagName(cmd, "sink-reply", "")
	dlsFlag.AddWithFlagName(cmd, "sink-dead-letter", "")
	return cmd
}
