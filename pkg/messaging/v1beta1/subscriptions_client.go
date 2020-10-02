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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingduckv1beta1 "knative.dev/eventing/pkg/apis/duck/v1beta1"
	"knative.dev/eventing/pkg/apis/messaging/v1beta1"
	clientv1beta1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1beta1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	knerrors "knative.dev/client/pkg/errors"
)

// KnSubscriptionsClient for interacting with Subscriptions
type KnSubscriptionsClient interface {

	// GetSubscription returns a Subscription by its name
	GetSubscription(name string) (*v1beta1.Subscription, error)

	// CreteSubscription creates a Subscription with given spec
	CreateSubscription(subscription *v1beta1.Subscription) error

	// UpdateSubscription updates a Subscription with given spec
	UpdateSubscription(subscription *v1beta1.Subscription) error

	// DeleteSubscription deletes a Subscription by its name
	DeleteSubscription(name string) error

	// ListSubscription lists all Subscriptions
	ListSubscription() (*v1beta1.SubscriptionList, error)

	// Namespace returns the namespace for this subscription client
	Namespace() string
}

// subscriptionsClient struct holds the client interface and namespace
type subscriptionsClient struct {
	client    clientv1beta1.SubscriptionInterface
	namespace string
}

// newKnSubscriptionsClient returns kn subscriptions client
func newKnSubscriptionsClient(client clientv1beta1.SubscriptionInterface, namespace string) KnSubscriptionsClient {
	return &subscriptionsClient{
		client:    client,
		namespace: namespace,
	}
}

// Get the namespace for which this client is created
func (c *subscriptionsClient) Namespace() string {
	return c.namespace
}

// GetSubscription gets Subscription by its name
func (c *subscriptionsClient) GetSubscription(name string) (*v1beta1.Subscription, error) {
	subscription, err := c.client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	err = updateMessagingGVK(subscription)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

// CreateSubscription creates Subscription with given spec
func (c *subscriptionsClient) CreateSubscription(subscription *v1beta1.Subscription) error {
	_, err := c.client.Create(subscription)
	return knerrors.GetError(err)
}

// UpdateSubscription creates Subscription with given spec
func (c *subscriptionsClient) UpdateSubscription(subscription *v1beta1.Subscription) error {
	_, err := c.client.Update(subscription)
	return knerrors.GetError(err)
}

// DeleteSubscription deletes Subscription by its name
func (c *subscriptionsClient) DeleteSubscription(name string) error {
	return knerrors.GetError(c.client.Delete(name, &metav1.DeleteOptions{}))
}

// ListSubscription lists subscriptions in configured namespace
func (c *subscriptionsClient) ListSubscription() (*v1beta1.SubscriptionList, error) {
	subscriptionList, err := c.client.List(metav1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}

	return updateSubscriptionListGVK(subscriptionList)
}

func updateSubscriptionListGVK(subscriptionList *v1beta1.SubscriptionList) (*v1beta1.SubscriptionList, error) {
	subscriptionListNew := subscriptionList.DeepCopy()
	err := updateMessagingGVK(subscriptionListNew)
	if err != nil {
		return nil, err
	}

	subscriptionListNew.Items = make([]v1beta1.Subscription, len(subscriptionList.Items))
	for idx, subscription := range subscriptionList.Items {
		subscriptionClone := subscription.DeepCopy()
		err := updateMessagingGVK(subscriptionClone)
		if err != nil {
			return nil, err
		}
		subscriptionListNew.Items[idx] = *subscriptionClone
	}
	return subscriptionListNew, nil
}

// SubscriptionBuilder is for building the Subscription object
type SubscriptionBuilder struct {
	subscription *v1beta1.Subscription
}

// NewSubscriptionBuilder for building Subscription object
func NewSubscriptionBuilder(name string) *SubscriptionBuilder {
	return &SubscriptionBuilder{subscription: &v1beta1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

// NewSubscriptionBuilderFromExisting for building Subscription object from existing Subscription object
func NewSubscriptionBuilderFromExisting(subs *v1beta1.Subscription) *SubscriptionBuilder {
	return &SubscriptionBuilder{subscription: subs.DeepCopy()}
}

// Channel sets the channel reference for this subscription
func (s *SubscriptionBuilder) Channel(channel *corev1.ObjectReference) *SubscriptionBuilder {
	if channel == nil {
		return s
	}

	s.subscription.Spec.Channel = *channel
	return s
}

func (s *SubscriptionBuilder) Subscriber(subs *duckv1.Destination) *SubscriptionBuilder {
	if subs == nil {
		return s
	}

	s.subscription.Spec.Subscriber = subs
	return s
}

func (s *SubscriptionBuilder) Reply(reply *duckv1.Destination) *SubscriptionBuilder {
	if reply == nil {
		return s
	}

	s.subscription.Spec.Reply = reply
	return s
}

func (s *SubscriptionBuilder) DeadLetterSink(dls *duckv1.Destination) *SubscriptionBuilder {
	if dls == nil {
		return s
	}

	ds := &eventingduckv1beta1.DeliverySpec{}
	ds.DeadLetterSink = dls
	s.subscription.Spec.Delivery = ds
	return s
}

// Build returns the Subscription object from the builder
func (s *SubscriptionBuilder) Build() *v1beta1.Subscription {
	return s.subscription
}
