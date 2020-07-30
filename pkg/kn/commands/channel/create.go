// Copyright © 2020 The Knative Authors
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

package channel

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	knflags "knative.dev/client/pkg/kn/flags"
	knmessagingv1beta1 "knative.dev/client/pkg/messaging/v1beta1"
)

// NewChannelCreateCommand to create event channels
func NewChannelCreateCommand(p *commands.KnParams) *cobra.Command {
	var ctypeFlags knflags.ChannelTypeFlags
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create an event channel",
		Example: `
  # Create a channel 'pipe' with default setting for channel configuration
  kn channel create pipe

  # Create a channel 'imc1' of type InMemoryChannel using inbuilt alias 'imc'
  kn channel create imc1 --type imc
  # same as above without using inbuilt alias but providing explicit GVK
  kn channel create imc1 --type messaging.knative.dev:v1beta1:InMemoryChannel

  # Create a channel 'k1' of type KafkaChannel
  kn channel create k1 --type messaging.knative.dev:v1alpha1:KafkaChannel`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn channel create' requires the channel name given as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := newChannelClient(p, cmd)
			if err != nil {
				return err
			}

			cb := knmessagingv1beta1.NewChannelBuilder(name)

			if cmd.Flag("type").Changed {
				gvk, err := ctypeFlags.Parse()
				if err != nil {
					return err
				}
				cb.Type(gvk)
			}

			err = client.CreateChannel(cb.Build())
			if err != nil {
				return knerrors.GetError(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Channel '%s' created in namespace '%s'.\n", name, namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	ctypeFlags.Add(cmd.Flags())
	return cmd
}
