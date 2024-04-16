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

package v1

import (
	"context"

	"knative.dev/client-pkg/pkg/util"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	clientmessagingv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1"

	knerrors "knative.dev/client-pkg/pkg/errors"
)

// KnChannelsClient for interacting with Channels
type KnChannelsClient interface {

	// GetChannel returns a Channel by its name
	GetChannel(ctx context.Context, name string) (*messagingv1.Channel, error)

	// CreteChannel creates a Channel with given spec
	CreateChannel(ctx context.Context, channel *messagingv1.Channel) error

	// DeleteChannel deletes a Channel by its name
	DeleteChannel(ctx context.Context, name string) error

	// ListChannel lists all Channels
	ListChannel(ctx context.Context) (*messagingv1.ChannelList, error)

	// Namespace returns the namespace for this channel client
	Namespace() string
}

// channelsClient struct holds the client interface and namespace
type channelsClient struct {
	client    clientmessagingv1.ChannelInterface
	namespace string
}

// newKnChannelsClient returns kn channels client
func newKnChannelsClient(client clientmessagingv1.ChannelInterface, namespace string) KnChannelsClient {
	return &channelsClient{
		client:    client,
		namespace: namespace,
	}
}

// Get the namespace for which this client is created
func (c *channelsClient) Namespace() string {
	return c.namespace
}

// GetChannel gets Channel by its name
func (c *channelsClient) GetChannel(ctx context.Context, name string) (*messagingv1.Channel, error) {
	channel, err := c.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	err = updateMessagingGVK(channel)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// CreateChannel creates Channel with given spec
func (c *channelsClient) CreateChannel(ctx context.Context, channel *messagingv1.Channel) error {
	_, err := c.client.Create(ctx, channel, metav1.CreateOptions{})
	return knerrors.GetError(err)
}

// DeleteChannel deletes Channel by its name
func (c *channelsClient) DeleteChannel(ctx context.Context, name string) error {
	return knerrors.GetError(c.client.Delete(ctx, name, metav1.DeleteOptions{}))
}

// ListChannel lists channels in configured namespace
func (c *channelsClient) ListChannel(ctx context.Context) (*messagingv1.ChannelList, error) {
	channelList, err := c.client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}

	return updateChannelListGVK(channelList)
}

func updateChannelListGVK(channelList *messagingv1.ChannelList) (*messagingv1.ChannelList, error) {
	channelListNew := channelList.DeepCopy()
	err := updateMessagingGVK(channelListNew)
	if err != nil {
		return nil, err
	}

	channelListNew.Items = make([]messagingv1.Channel, len(channelList.Items))
	for idx, channel := range channelList.Items {
		channelClone := channel.DeepCopy()
		err := updateMessagingGVK(channelClone)
		if err != nil {
			return nil, err
		}
		channelListNew.Items[idx] = *channelClone
	}
	return channelListNew, nil
}

// ChannelBuilder is for building the Channel object
type ChannelBuilder struct {
	channel *messagingv1.Channel
}

// NewChannelBuilder for building Channel object
func NewChannelBuilder(name, namespace string) *ChannelBuilder {
	return &ChannelBuilder{channel: &messagingv1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}}
}

// WithGvk sets the GVK on the channel
func (c *ChannelBuilder) WithGvk() *ChannelBuilder {
	_ = util.UpdateGroupVersionKindWithScheme(c.channel, eventingv1.SchemeGroupVersion, scheme.Scheme)
	return c
}

// Type sets the type of the channel to create
func (c *ChannelBuilder) Type(gvk *schema.GroupVersionKind) *ChannelBuilder {
	if gvk == nil {
		return c
	}

	c.channel.TypeMeta = metav1.TypeMeta{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
	}

	spec := &messagingv1.ChannelTemplateSpec{}
	spec.Kind = gvk.Kind
	spec.APIVersion = gvk.GroupVersion().String()
	c.channel.Spec.ChannelTemplate = spec
	return c
}

// Build returns the Channel object from the builder
func (c *ChannelBuilder) Build() *messagingv1.Channel {
	return c.channel
}
