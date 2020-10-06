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
	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"
	clientv1beta1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1beta1"

	"knative.dev/client/pkg/kn/commands"
	messagingv1beta1 "knative.dev/client/pkg/messaging/v1beta1"
)

// NewChannelCommand to manage event channels
func NewChannelCommand(p *commands.KnParams) *cobra.Command {
	channelCmd := &cobra.Command{
		Use:     "channel COMMAND",
		Short:   "Manage event channels (alias: channels)",
		Aliases: []string{"channels"},
	}
	channelCmd.AddCommand(NewChannelCreateCommand(p))
	channelCmd.AddCommand(NewChannelListCommand(p))
	channelCmd.AddCommand(NewChannelDeleteCommand(p))
	channelCmd.AddCommand(NewChannelDescribeCommand(p))
	return channelCmd
}

var channelClientFactory func(config clientcmd.ClientConfig, namespace string) (messagingv1beta1.KnChannelsClient, error)

func newChannelClient(p *commands.KnParams, cmd *cobra.Command) (messagingv1beta1.KnChannelsClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if channelClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return channelClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := clientv1beta1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return messagingv1beta1.NewKnMessagingClient(client, namespace).ChannelsClient(), nil
}
