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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
	clientv1beta1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1beta1"

	"knative.dev/client/pkg/util"
)

// KnMessagingClient to Eventing Messaging. All methods are relative to
// the namespace specified during construction
type KnMessagingClient interface {
	// Get the Channels client
	ChannelsClient() KnChannelsClient

	// Get the Subscriptions client
	SubscriptionsClient() KnSubscriptionsClient
}

// messagingClient holds Messaging client interface and namespace
type messagingClient struct {
	client    clientv1beta1.MessagingV1beta1Interface
	namespace string
}

// NewKnMessagingClient for managing all eventing messaging types
func NewKnMessagingClient(client clientv1beta1.MessagingV1beta1Interface, namespace string) KnMessagingClient {
	return &messagingClient{
		client:    client,
		namespace: namespace,
	}
}

// ChannelsClient for working with Channels
func (c *messagingClient) ChannelsClient() KnChannelsClient {
	return newKnChannelsClient(c.client.Channels(c.namespace), c.namespace)
}

// SubscriptionsClient for working with Subscriptions
func (c *messagingClient) SubscriptionsClient() KnSubscriptionsClient {
	return newKnSubscriptionsClient(c.client.Subscriptions(c.namespace), c.namespace)
}

// update GVK of object
func updateMessagingGVK(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, messagingv1beta1.SchemeGroupVersion, scheme.Scheme)
}

// BuiltInChannelGVKs returns the GVKs for built in channel
func BuiltInChannelGVKs() []schema.GroupVersionKind {
	return []schema.GroupVersionKind{
		messagingv1beta1.SchemeGroupVersion.WithKind("InMemoryChannel"),
	}
}
