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

package v1

import (
	"context"
	"fmt"

	"knative.dev/client-pkg/pkg/config"

	"k8s.io/client-go/util/retry"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	clientmessagingv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	knerrors "knative.dev/client-pkg/pkg/errors"
)

type SubscriptionUpdateFunc func(origSub *messagingv1.Subscription) (*messagingv1.Subscription, error)

// KnSubscriptionsClient for interacting with Subscriptions
type KnSubscriptionsClient interface {

	// GetSubscription returns a Subscription by its name
	GetSubscription(ctx context.Context, name string) (*messagingv1.Subscription, error)

	// CreateSubscription creates a Subscription with given spec
	CreateSubscription(ctx context.Context, subscription *messagingv1.Subscription) error

	// UpdateSubscription updates a Subscription with given spec
	UpdateSubscription(ctx context.Context, subscription *messagingv1.Subscription) error

	// UpdateSubscriptionWithRetry updates a Subscription and retries on conflict error
	UpdateSubscriptionWithRetry(ctx context.Context, name string, updateFunc SubscriptionUpdateFunc, nrRetries int) error

	// DeleteSubscription deletes a Subscription by its name
	DeleteSubscription(ctx context.Context, name string) error

	// ListSubscription lists all Subscriptions
	ListSubscription(ctx context.Context) (*messagingv1.SubscriptionList, error)

	// Namespace returns the namespace for this subscription client
	Namespace() string
}

// subscriptionsClient struct holds the client interface and namespace
type subscriptionsClient struct {
	client    clientmessagingv1.SubscriptionInterface
	namespace string
}

// newKnSubscriptionsClient returns kn subscriptions client
func newKnSubscriptionsClient(client clientmessagingv1.SubscriptionInterface, namespace string) KnSubscriptionsClient {
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
func (c *subscriptionsClient) GetSubscription(ctx context.Context, name string) (*messagingv1.Subscription, error) {
	subscription, err := c.client.Get(ctx, name, metav1.GetOptions{})
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
func (c *subscriptionsClient) CreateSubscription(ctx context.Context, subscription *messagingv1.Subscription) error {
	_, err := c.client.Create(ctx, subscription, metav1.CreateOptions{})
	return knerrors.GetError(err)
}

// UpdateSubscription creates Subscription with given spec
func (c *subscriptionsClient) UpdateSubscription(ctx context.Context, subscription *messagingv1.Subscription) error {
	_, err := c.client.Update(ctx, subscription, metav1.UpdateOptions{})
	return knerrors.GetError(err)
}

func (c *subscriptionsClient) UpdateSubscriptionWithRetry(ctx context.Context, name string, updateFunc SubscriptionUpdateFunc, nrRetries int) error {
	return updateSubscriptionWithRetry(ctx, c, name, updateFunc, nrRetries)
}

func updateSubscriptionWithRetry(ctx context.Context, c KnSubscriptionsClient, name string, updateFunc SubscriptionUpdateFunc, nrRetries int) error {
	b := config.DefaultRetry
	b.Steps = nrRetries
	err := retry.RetryOnConflict(b, func() error {
		return updateSubscription(ctx, c, name, updateFunc)
	})
	return err
}

func updateSubscription(ctx context.Context, c KnSubscriptionsClient, name string, updateFunc SubscriptionUpdateFunc) error {
	sub, err := c.GetSubscription(ctx, name)
	if err != nil {
		return err
	}
	if sub.GetDeletionTimestamp() != nil {
		return fmt.Errorf("can't update subscription %s because it has been marked for deletion", name)
	}
	updatedSub, err := updateFunc(sub.DeepCopy())
	if err != nil {
		return err
	}

	return c.UpdateSubscription(ctx, updatedSub)
}

// DeleteSubscription deletes Subscription by its name
func (c *subscriptionsClient) DeleteSubscription(ctx context.Context, name string) error {
	return knerrors.GetError(c.client.Delete(ctx, name, metav1.DeleteOptions{}))
}

// ListSubscription lists subscriptions in configured namespace
func (c *subscriptionsClient) ListSubscription(ctx context.Context) (*messagingv1.SubscriptionList, error) {
	subscriptionList, err := c.client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}

	return updateSubscriptionListGVK(subscriptionList)
}

func updateSubscriptionListGVK(subscriptionList *messagingv1.SubscriptionList) (*messagingv1.SubscriptionList, error) {
	subscriptionListNew := subscriptionList.DeepCopy()
	err := updateMessagingGVK(subscriptionListNew)
	if err != nil {
		return nil, err
	}

	subscriptionListNew.Items = make([]messagingv1.Subscription, len(subscriptionList.Items))
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
	subscription *messagingv1.Subscription
}

// NewSubscriptionBuilder for building Subscription object
func NewSubscriptionBuilder(name string) *SubscriptionBuilder {
	return &SubscriptionBuilder{subscription: &messagingv1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: messagingv1.SchemeGroupVersion.String(),
			Kind:       "Subscription",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

// NewSubscriptionBuilderFromExisting for building Subscription object from existing Subscription object
func NewSubscriptionBuilderFromExisting(subs *messagingv1.Subscription) *SubscriptionBuilder {
	return &SubscriptionBuilder{subscription: subs.DeepCopy()}
}

// Channel sets the channel reference for this subscription
func (s *SubscriptionBuilder) Channel(channel *duckv1.KReference) *SubscriptionBuilder {
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

	ds := &eventingduckv1.DeliverySpec{}
	ds.DeadLetterSink = dls
	s.subscription.Spec.Delivery = ds
	return s
}

// Build returns the Subscription object from the builder
func (s *SubscriptionBuilder) Build() *messagingv1.Subscription {
	return s.subscription
}
