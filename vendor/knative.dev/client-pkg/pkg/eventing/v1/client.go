// Copyright © 2019 The Knative Authors
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
	"fmt"
	"time"

	apis_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/util/retry"
	"knative.dev/client-pkg/pkg/config"
	v1 "knative.dev/eventing/pkg/apis/duck/v1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
	clientv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kn_errors "knative.dev/client-pkg/pkg/errors"
	"knative.dev/client-pkg/pkg/util"
	"knative.dev/client-pkg/pkg/wait"
)

type TriggerUpdateFunc func(origTrigger *eventingv1.Trigger) (*eventingv1.Trigger, error)
type BrokerUpdateFunc func(origBroker *eventingv1.Broker) (*eventingv1.Broker, error)

// KnEventingClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnEventingClient interface {
	// Namespace in which this client is operating for
	Namespace() string
	// CreateTrigger is used to create an instance of trigger
	CreateTrigger(ctx context.Context, trigger *eventingv1.Trigger) error
	// DeleteTrigger is used to delete an instance of trigger
	DeleteTrigger(ctx context.Context, name string) error
	// GetTrigger is used to get an instance of trigger
	GetTrigger(ctx context.Context, name string) (*eventingv1.Trigger, error)
	// ListTriggers returns list of trigger CRDs
	ListTriggers(ctx context.Context) (*eventingv1.TriggerList, error)
	// UpdateTrigger is used to update an instance of trigger
	UpdateTrigger(ctx context.Context, trigger *eventingv1.Trigger) error
	// UpdateTriggerWithRetry is used to update an instance of trigger
	UpdateTriggerWithRetry(ctx context.Context, name string, updateFunc TriggerUpdateFunc, nrRetries int) error
	// CreateBroker is used to create an instance of broker
	CreateBroker(ctx context.Context, broker *eventingv1.Broker) error
	// GetBroker is used to get an instance of broker
	GetBroker(ctx context.Context, name string) (*eventingv1.Broker, error)
	// DeleteBroker is used to delete an instance of broker
	DeleteBroker(ctx context.Context, name string, timeout time.Duration) error
	// ListBrokers returns list of broker CRDs
	ListBrokers(ctx context.Context) (*eventingv1.BrokerList, error)
	// UpdateBroker is used to update an instance of broker
	UpdateBroker(ctx context.Context, broker *eventingv1.Broker) error
	// UpdateBrokerWithRetry is used to update an instance of broker
	UpdateBrokerWithRetry(ctx context.Context, name string, updateFunc BrokerUpdateFunc, nrRetries int) error
}

// KnEventingClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type knEventingClient struct {
	client    clientv1.EventingV1Interface
	namespace string
}

// NewKnEventingClient is to invoke Eventing Sources Client API to create object
func NewKnEventingClient(client clientv1.EventingV1Interface, namespace string) KnEventingClient {
	return &knEventingClient{
		client:    client,
		namespace: namespace,
	}
}

// CreateTrigger is used to create an instance of trigger
func (c *knEventingClient) CreateTrigger(ctx context.Context, trigger *eventingv1.Trigger) error {
	_, err := c.client.Triggers(c.namespace).Create(ctx, trigger, meta_v1.CreateOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// DeleteTrigger is used to delete an instance of trigger
func (c *knEventingClient) DeleteTrigger(ctx context.Context, name string) error {
	err := c.client.Triggers(c.namespace).Delete(ctx, name, apis_v1.DeleteOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// GetTrigger is used to get an instance of trigger
func (c *knEventingClient) GetTrigger(ctx context.Context, name string) (*eventingv1.Trigger, error) {
	trigger, err := c.client.Triggers(c.namespace).Get(ctx, name, apis_v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	err = updateEventingGVK(trigger)
	if err != nil {
		return nil, err
	}
	return trigger, nil
}

func (c *knEventingClient) ListTriggers(ctx context.Context) (*eventingv1.TriggerList, error) {
	triggerList, err := c.client.Triggers(c.namespace).List(ctx, apis_v1.ListOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	triggerListNew := triggerList.DeepCopy()
	err = updateEventingGVK(triggerListNew)
	if err != nil {
		return nil, err
	}

	triggerListNew.Items = make([]eventingv1.Trigger, len(triggerList.Items))
	for idx, trigger := range triggerList.Items {
		triggerClone := trigger.DeepCopy()
		err := updateEventingGVK(triggerClone)
		if err != nil {
			return nil, err
		}
		triggerListNew.Items[idx] = *triggerClone
	}
	return triggerListNew, nil
}

// UpdateTrigger is used to update an instance of trigger
func (c *knEventingClient) UpdateTrigger(ctx context.Context, trigger *eventingv1.Trigger) error {
	_, err := c.client.Triggers(c.namespace).Update(ctx, trigger, meta_v1.UpdateOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

func (c *knEventingClient) UpdateTriggerWithRetry(ctx context.Context, name string, updateFunc TriggerUpdateFunc, nrRetries int) error {
	return updateTriggerWithRetry(ctx, c, name, updateFunc, nrRetries)
}

func updateTriggerWithRetry(ctx context.Context, c KnEventingClient, name string, updateFunc TriggerUpdateFunc, nrRetries int) error {
	b := config.DefaultRetry
	b.Steps = nrRetries
	updateTriggerFunc := func() error {
		return updateTrigger(ctx, c, name, updateFunc)
	}
	err := retry.RetryOnConflict(b, updateTriggerFunc)
	return err
}

func updateTrigger(ctx context.Context, c KnEventingClient, name string, updateFunc TriggerUpdateFunc) error {
	trigger, err := c.GetTrigger(ctx, name)
	if err != nil {
		return err
	}
	if trigger.GetDeletionTimestamp() != nil {
		return fmt.Errorf("can't update trigger %s because it has been marked for deletion", name)
	}
	updatedTrigger, err := updateFunc(trigger.DeepCopy())
	if err != nil {
		return err
	}

	return c.UpdateTrigger(ctx, updatedTrigger)
}

// Namespace returns the namespace this client is bound to
func (c *knEventingClient) Namespace() string {
	return c.namespace
}

// update with the eventingv1 group + version
func updateEventingGVK(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, eventingv1.SchemeGroupVersion, scheme.Scheme)
}

// TriggerBuilder is for building the trigger
type TriggerBuilder struct {
	trigger *eventingv1.Trigger
}

// NewTriggerBuilder for building trigger object
func NewTriggerBuilder(name string) *TriggerBuilder {
	return &TriggerBuilder{&eventingv1.Trigger{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
	}}
}

// WithGvk sets the GVK for the triggers (which otherwise remains empty
func (b *TriggerBuilder) WithGvk() *TriggerBuilder {
	_ = updateEventingGVK(b.trigger)
	return b
}

// NewTriggerBuilderFromExisting for building the object from existing Trigger object
func NewTriggerBuilderFromExisting(trigger *eventingv1.Trigger) *TriggerBuilder {
	return &TriggerBuilder{trigger: trigger.DeepCopy()}
}

// Namespace for this trigger
func (b *TriggerBuilder) Namespace(ns string) *TriggerBuilder {
	b.trigger.Namespace = ns
	return b
}

// Subscriber for the trigger to send to (it's a Sink actually)
func (b *TriggerBuilder) Subscriber(subscriber *duckv1.Destination) *TriggerBuilder {
	b.trigger.Spec.Subscriber = *subscriber
	return b
}

// Broker to set the broker of trigger object
func (b *TriggerBuilder) Broker(broker string) *TriggerBuilder {
	b.trigger.Spec.Broker = broker

	return b
}

// InjectBroker to add annotation to setup default broker
func (b *TriggerBuilder) InjectBroker(inject bool) *TriggerBuilder {
	if inject {
		meta_v1.SetMetaDataAnnotation(&b.trigger.ObjectMeta, eventingv1.InjectionAnnotation, "enabled")
	} else {
		if meta_v1.HasAnnotation(b.trigger.ObjectMeta, eventingv1.InjectionAnnotation) {
			delete(b.trigger.ObjectMeta.Annotations, eventingv1.InjectionAnnotation)
		}
	}
	return b
}

func (b *TriggerBuilder) Filters(filters map[string]string) *TriggerBuilder {
	if len(filters) == 0 {
		b.trigger.Spec.Filter = &eventingv1.TriggerFilter{}
		return b
	}
	filter := b.trigger.Spec.Filter
	if filter == nil {
		filter = &eventingv1.TriggerFilter{}
		b.trigger.Spec.Filter = filter
	}
	filter.Attributes = eventingv1.TriggerFilterAttributes{}
	for k, v := range filters {
		filter.Attributes[k] = v
	}
	return b
}

// Build to return an instance of trigger object
func (b *TriggerBuilder) Build() *eventingv1.Trigger {
	return b.trigger
}

// CreateBroker is used to create an instance of broker
func (c *knEventingClient) CreateBroker(ctx context.Context, broker *eventingv1.Broker) error {
	_, err := c.client.Brokers(c.namespace).Create(ctx, broker, meta_v1.CreateOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// GetBroker is used to get an instance of broker
func (c *knEventingClient) GetBroker(ctx context.Context, name string) (*eventingv1.Broker, error) {
	broker, err := c.client.Brokers(c.namespace).Get(ctx, name, apis_v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	err = updateEventingGVK(broker)
	if err != nil {
		return nil, err
	}
	return broker, nil
}

// WatchBroker is used to create watcher object
func (c *knEventingClient) WatchBroker(ctx context.Context, name string, initialVersion string, timeout time.Duration) (watch.Interface, error) {
	return wait.NewWatcherWithVersion(ctx, c.client.Brokers(c.namespace).Watch, c.client.RESTClient(), c.namespace, "brokers", name, initialVersion, timeout)
}

// DeleteBroker is used to delete an instance of broker and wait for completion until given timeout
// For `timeout == 0` delete is performed async without any wait
func (c *knEventingClient) DeleteBroker(ctx context.Context, name string, timeout time.Duration) error {
	broker, err := c.GetBroker(ctx, name)
	if err != nil {
		return err
	}
	if broker.GetDeletionTimestamp() != nil {
		return fmt.Errorf("can't delete broker '%s' because it has been already marked for deletion", name)
	}
	if timeout == 0 {
		return c.deleteBroker(ctx, name, apis_v1.DeletePropagationBackground)
	}
	waitC := make(chan error)
	go func() {
		waitForEvent := wait.NewWaitForEvent("broker", c.WatchBroker, func(evt *watch.Event) bool { return evt.Type == watch.Deleted })
		err, _ := waitForEvent.Wait(ctx, name, broker.ResourceVersion, wait.Options{Timeout: &timeout}, wait.NoopMessageCallback())
		waitC <- err
	}()
	err = c.deleteBroker(ctx, name, apis_v1.DeletePropagationForeground)
	if err != nil {
		return err
	}
	return <-waitC
}

// deleteBroker is used to delete an instance of broker
func (c *knEventingClient) deleteBroker(ctx context.Context, name string, propagationPolicy apis_v1.DeletionPropagation) error {
	err := c.client.Brokers(c.namespace).Delete(ctx, name, apis_v1.DeleteOptions{PropagationPolicy: &propagationPolicy})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// ListBrokers is used to retrieve the list of broker instances
func (c *knEventingClient) ListBrokers(ctx context.Context) (*eventingv1.BrokerList, error) {
	brokerList, err := c.client.Brokers(c.namespace).List(ctx, apis_v1.ListOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	brokerListNew := brokerList.DeepCopy()
	err = updateEventingGVK(brokerListNew)
	if err != nil {
		return nil, err
	}

	brokerListNew.Items = make([]eventingv1.Broker, len(brokerList.Items))
	for idx, trigger := range brokerList.Items {
		triggerClone := trigger.DeepCopy()
		err := updateEventingGVK(triggerClone)
		if err != nil {
			return nil, err
		}
		brokerListNew.Items[idx] = *triggerClone
	}
	return brokerListNew, nil
}

// UpdateBroker is used to update an instance of broker
func (c *knEventingClient) UpdateBroker(ctx context.Context, broker *eventingv1.Broker) error {
	_, err := c.client.Brokers(c.namespace).Update(ctx, broker, meta_v1.UpdateOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

func (c *knEventingClient) UpdateBrokerWithRetry(ctx context.Context, name string, updateFunc BrokerUpdateFunc, nrRetries int) error {
	return updateBrokerWithRetry(ctx, c, name, updateFunc, nrRetries)
}

func updateBrokerWithRetry(ctx context.Context, c KnEventingClient, name string, updateFunc BrokerUpdateFunc, nrRetries int) error {
	b := config.DefaultRetry
	b.Steps = nrRetries
	updateBrokerFunc := func() error {
		return updateBroker(ctx, c, name, updateFunc)
	}
	err := retry.RetryOnConflict(b, updateBrokerFunc)
	return err
}

func updateBroker(ctx context.Context, c KnEventingClient, name string, updateFunc BrokerUpdateFunc) error {
	broker, err := c.GetBroker(ctx, name)
	if err != nil {
		return err
	}
	if broker.GetDeletionTimestamp() != nil {
		return fmt.Errorf("can't update broker %s because it has been marked for deletion", name)
	}
	updatedBroker, err := updateFunc(broker.DeepCopy())
	if err != nil {
		return err
	}

	return c.UpdateBroker(ctx, updatedBroker)
}

// BrokerBuilder is for building the broker
type BrokerBuilder struct {
	broker *eventingv1.Broker
}

// NewBrokerBuilder for building broker object
func NewBrokerBuilder(name string) *BrokerBuilder {
	return &BrokerBuilder{broker: &eventingv1.Broker{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
	}}
}

// NewBrokerBuilderFromExisting returns broker builder from original broker
func NewBrokerBuilderFromExisting(broker *eventingv1.Broker) *BrokerBuilder {
	return &BrokerBuilder{
		broker: broker,
	}
}

// WithGvk add the GVK coordinates for read tests
func (b *BrokerBuilder) WithGvk() *BrokerBuilder {
	_ = updateEventingGVK(b.broker)
	return b
}

// Namespace for broker builder
func (b *BrokerBuilder) Namespace(ns string) *BrokerBuilder {
	b.broker.Namespace = ns
	return b
}

// Class for broker builder
func (b *BrokerBuilder) Class(class string) *BrokerBuilder {
	if class == "" {
		return b
	}
	if len(b.broker.Annotations) == 0 {
		b.broker.Annotations = make(map[string]string)
	}
	b.broker.Annotations[eventingv1.BrokerClassAnnotationKey] = class
	return b
}

// DlSink for the broker builder
func (b *BrokerBuilder) DlSink(dlSink *duckv1.Destination) *BrokerBuilder {
	empty := duckv1.Destination{}
	if dlSink == nil || *dlSink == empty {
		return b
	}
	if b.broker.Spec.Delivery == nil {
		b.broker.Spec.Delivery = &v1.DeliverySpec{}
	}
	b.broker.Spec.Delivery.DeadLetterSink = dlSink
	return b
}

// Retry for the broker builder
func (b *BrokerBuilder) Retry(retry *int32) *BrokerBuilder {
	if retry == nil || *retry == 0 {
		return b
	}
	if b.broker.Spec.Delivery == nil {
		b.broker.Spec.Delivery = &v1.DeliverySpec{}
	}
	b.broker.Spec.Delivery.Retry = retry
	return b
}

// Timeout for the broker builder
func (b *BrokerBuilder) Timeout(timeout *string) *BrokerBuilder {
	if timeout == nil || *timeout == "" {
		return b
	}
	if b.broker.Spec.Delivery == nil {
		b.broker.Spec.Delivery = &v1.DeliverySpec{}
	}
	b.broker.Spec.Delivery.Timeout = timeout
	return b
}

// BackoffPolicy for the broker builder
func (b *BrokerBuilder) BackoffPolicy(policyType *v1.BackoffPolicyType) *BrokerBuilder {
	if policyType == nil || *policyType == "" {
		return b
	}
	if b.broker.Spec.Delivery == nil {
		b.broker.Spec.Delivery = &v1.DeliverySpec{}
	}
	b.broker.Spec.Delivery.BackoffPolicy = policyType
	return b
}

// BackoffDelay for the broker builder
func (b *BrokerBuilder) BackoffDelay(backoffDelay *string) *BrokerBuilder {
	if backoffDelay == nil || *backoffDelay == "" {
		return b
	}
	if b.broker.Spec.Delivery == nil {
		b.broker.Spec.Delivery = &v1.DeliverySpec{}
	}
	b.broker.Spec.Delivery.BackoffDelay = backoffDelay
	return b
}

// RetryAfterMax for the broker builder
func (b *BrokerBuilder) RetryAfterMax(max *string) *BrokerBuilder {
	if max == nil || *max == "" {
		return b
	}
	if b.broker.Spec.Delivery == nil {
		b.broker.Spec.Delivery = &v1.DeliverySpec{}
	}
	b.broker.Spec.Delivery.RetryAfterMax = max
	return b

}

// Config for the broker builder
func (b *BrokerBuilder) Config(config *duckv1.KReference) *BrokerBuilder {
	b.broker.Spec.Config = config
	return b

}

// Build to return an instance of broker object
func (b *BrokerBuilder) Build() *eventingv1.Broker {
	return b.broker
}
