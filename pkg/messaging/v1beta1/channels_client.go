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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/eventing/pkg/apis/messaging/v1beta1"
	clientv1beta1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1beta1"

	knerrors "knative.dev/client/pkg/errors"
)

// KnChannelsClient for interacting with Channels
type KnChannelsClient interface {

	// GetChannel returns a Channel by its name
	GetChannel(name string) (*v1beta1.Channel, error)

	// CreteChannel creates a Channel with given spec
	CreateChannel(channel *v1beta1.Channel) error

	// DeleteChannel deletes a Channel by its name
	DeleteChannel(name string) error

	// ListChannel lists all Channels
	ListChannel() (*v1beta1.ChannelList, error)

	// Namespace returns the namespace for this channel client
	Namespace() string
}

// channelsClient struct holds the client interface and namespace
type channelsClient struct {
	client    clientv1beta1.ChannelInterface
	namespace string
}

// newKnChannelsClient returns kn channels client
func newKnChannelsClient(client clientv1beta1.ChannelInterface, namespace string) KnChannelsClient {
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
func (c *channelsClient) GetChannel(name string) (*v1beta1.Channel, error) {
	channel, err := c.client.Get(name, metav1.GetOptions{})
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
func (c *channelsClient) CreateChannel(channel *v1beta1.Channel) error {
	_, err := c.client.Create(channel)
	return knerrors.GetError(err)
}

// DeleteChannel deletes Channel by its name
func (c *channelsClient) DeleteChannel(name string) error {
	return knerrors.GetError(c.client.Delete(name, &metav1.DeleteOptions{}))
}

// ListChannel lists channels in configured namespace
func (c *channelsClient) ListChannel() (*v1beta1.ChannelList, error) {
	channelList, err := c.client.List(metav1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}

	return updateChannelListGVK(channelList)
}

func updateChannelListGVK(channelList *v1beta1.ChannelList) (*v1beta1.ChannelList, error) {
	channelListNew := channelList.DeepCopy()
	err := updateMessagingGVK(channelListNew)
	if err != nil {
		return nil, err
	}

	channelListNew.Items = make([]v1beta1.Channel, len(channelList.Items))
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
	channel *v1beta1.Channel
}

// NewChannelBuilder for building Channel object
func NewChannelBuilder(name string) *ChannelBuilder {
	return &ChannelBuilder{channel: &v1beta1.Channel{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

// Type sets the type of the channel to create
func (c *ChannelBuilder) Type(gvk *schema.GroupVersionKind) *ChannelBuilder {
	if gvk == nil {
		return c
	}

	spec := &v1beta1.ChannelTemplateSpec{}
	spec.Kind = gvk.Kind
	spec.APIVersion = gvk.GroupVersion().String()
	c.channel.Spec.ChannelTemplate = spec
	return c
}

// Build returns the Channel object from the builder
func (c *ChannelBuilder) Build() *v1beta1.Channel {
	return c.channel
}
