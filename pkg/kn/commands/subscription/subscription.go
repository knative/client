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
	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"
	clientv1beta1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1beta1"

	"knative.dev/client/pkg/kn/commands"
	messagingv1beta1 "knative.dev/client/pkg/messaging/v1beta1"
)

// NewSubscriptionCommand to manage event subscriptions
func NewSubscriptionCommand(p *commands.KnParams) *cobra.Command {
	subscriptionCmd := &cobra.Command{
		Use:     "subscription COMMAND",
		Short:   "Manage event subscriptions (aliases: subscriptions, sub)",
		Aliases: []string{"subscriptions", "sub"},
	}
	subscriptionCmd.AddCommand(NewSubscriptionCreateCommand(p))
	subscriptionCmd.AddCommand(NewSubscriptionUpdateCommand(p))
	subscriptionCmd.AddCommand(NewSubscriptionListCommand(p))
	subscriptionCmd.AddCommand(NewSubscriptionDeleteCommand(p))
	subscriptionCmd.AddCommand(NewSubscriptionDescribeCommand(p))
	return subscriptionCmd
}

var subscriptionClientFactory func(config clientcmd.ClientConfig, namespace string) (messagingv1beta1.KnSubscriptionsClient, error)

func newSubscriptionClient(p *commands.KnParams, cmd *cobra.Command) (messagingv1beta1.KnSubscriptionsClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if subscriptionClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return subscriptionClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := clientv1beta1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return messagingv1beta1.NewKnMessagingClient(client, namespace).SubscriptionsClient(), nil
}
