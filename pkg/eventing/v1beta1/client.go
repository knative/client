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

package v1beta1

import (
	"time"

	apis_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
	client_v1beta1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1beta1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kn_errors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"
)

// KnEventingClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnEventingClient interface {
	// Namespace in which this client is operating for
	Namespace() string
	// CreateTrigger is used to create an instance of trigger
	CreateTrigger(trigger *v1beta1.Trigger) error
	// DeleteTrigger is used to delete an instance of trigger
	DeleteTrigger(name string) error
	// GetTrigger is used to get an instance of trigger
	GetTrigger(name string) (*v1beta1.Trigger, error)
	// ListTrigger returns list of trigger CRDs
	ListTriggers() (*v1beta1.TriggerList, error)
	// UpdateTrigger is used to update an instance of trigger
	UpdateTrigger(trigger *v1beta1.Trigger) error
	// CreateBroker is used to create an instance of broker
	CreateBroker(broker *v1beta1.Broker) error
	// GetBroker is used to get an instance of broker
	GetBroker(name string) (*v1beta1.Broker, error)
	// DeleteBroker is used to delete an instance of broker
	DeleteBroker(name string, timeout time.Duration) error
	// ListBroker returns list of broker CRDs
	ListBrokers() (*v1beta1.BrokerList, error)
}

// KnEventingClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type knEventingClient struct {
	client    client_v1beta1.EventingV1beta1Interface
	namespace string
}

// NewKnEventingClient is to invoke Eventing Sources Client API to create object
func NewKnEventingClient(client client_v1beta1.EventingV1beta1Interface, namespace string) KnEventingClient {
	return &knEventingClient{
		client:    client,
		namespace: namespace,
	}
}

//CreateTrigger is used to create an instance of trigger
func (c *knEventingClient) CreateTrigger(trigger *v1beta1.Trigger) error {
	_, err := c.client.Triggers(c.namespace).Create(trigger)
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

//DeleteTrigger is used to delete an instance of trigger
func (c *knEventingClient) DeleteTrigger(name string) error {
	err := c.client.Triggers(c.namespace).Delete(name, &apis_v1.DeleteOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

//GetTrigger is used to get an instance of trigger
func (c *knEventingClient) GetTrigger(name string) (*v1beta1.Trigger, error) {
	trigger, err := c.client.Triggers(c.namespace).Get(name, apis_v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	return trigger, nil
}

func (c *knEventingClient) ListTriggers() (*v1beta1.TriggerList, error) {
	triggerList, err := c.client.Triggers(c.namespace).List(apis_v1.ListOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	triggerListNew := triggerList.DeepCopy()
	err = updateEventingGVK(triggerListNew)
	if err != nil {
		return nil, err
	}

	triggerListNew.Items = make([]v1beta1.Trigger, len(triggerList.Items))
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

//CreateTrigger is used to create an instance of trigger
func (c *knEventingClient) UpdateTrigger(trigger *v1beta1.Trigger) error {
	_, err := c.client.Triggers(c.namespace).Update(trigger)
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// Return the client's namespace
func (c *knEventingClient) Namespace() string {
	return c.namespace
}

// update with the v1beta1 group + version
func updateEventingGVK(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, v1beta1.SchemeGroupVersion, scheme.Scheme)
}

// TriggerBuilder is for building the trigger
type TriggerBuilder struct {
	trigger *v1beta1.Trigger
}

// NewTriggerBuilder for building trigger object
func NewTriggerBuilder(name string) *TriggerBuilder {
	return &TriggerBuilder{trigger: &v1beta1.Trigger{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
	}}
}

// NewTriggerBuilderFromExisting for building the object from existing Trigger object
func NewTriggerBuilderFromExisting(trigger *v1beta1.Trigger) *TriggerBuilder {
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
		meta_v1.SetMetaDataAnnotation(&b.trigger.ObjectMeta, v1beta1.DeprecatedInjectionAnnotation, "enabled")
	} else {
		if meta_v1.HasAnnotation(b.trigger.ObjectMeta, v1beta1.DeprecatedInjectionAnnotation) {
			delete(b.trigger.ObjectMeta.Annotations, v1beta1.DeprecatedInjectionAnnotation)
		}
	}
	return b
}

func (b *TriggerBuilder) Filters(filters map[string]string) *TriggerBuilder {
	if len(filters) == 0 {
		b.trigger.Spec.Filter = &v1beta1.TriggerFilter{}
		return b
	}
	filter := b.trigger.Spec.Filter
	if filter == nil {
		filter = &v1beta1.TriggerFilter{}
		b.trigger.Spec.Filter = filter
	}
	filter.Attributes = v1beta1.TriggerFilterAttributes{}
	for k, v := range filters {
		filter.Attributes[k] = v
	}
	return b
}

// Build to return an instance of trigger object
func (b *TriggerBuilder) Build() *v1beta1.Trigger {
	return b.trigger
}

// CreateBroker is used to create an instance of broker
func (c *knEventingClient) CreateBroker(broker *v1beta1.Broker) error {
	_, err := c.client.Brokers(c.namespace).Create(broker)
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// GetBroker is used to get an instance of broker
func (c *knEventingClient) GetBroker(name string) (*v1beta1.Broker, error) {
	trigger, err := c.client.Brokers(c.namespace).Get(name, apis_v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	return trigger, nil
}

// WatchBroker is used to create watcher object
func (c *knEventingClient) WatchBroker(name string, timeout time.Duration) (watch.Interface, error) {
	return wait.NewWatcher(c.client.Brokers(c.namespace).Watch,
		c.client.RESTClient(), c.namespace, "brokers", name, timeout)
}

// DeleteBroker is used to delete an instance of broker and wait for completion until given timeout
// For `timeout == 0` delete is performed async without any wait
func (c *knEventingClient) DeleteBroker(name string, timeout time.Duration) error {
	if timeout == 0 {
		return c.deleteBroker(name, apis_v1.DeletePropagationBackground)
	}
	waitC := make(chan error)
	go func() {
		waitForEvent := wait.NewWaitForEvent("broker", c.WatchBroker, func(evt *watch.Event) bool { return evt.Type == watch.Deleted })
		err, _ := waitForEvent.Wait(name, wait.Options{Timeout: &timeout}, wait.NoopMessageCallback())
		waitC <- err
	}()
	err := c.deleteBroker(name, apis_v1.DeletePropagationForeground)
	if err != nil {
		return err
	}
	return <-waitC
}

// deleteBroker is used to delete an instance of broker
func (c *knEventingClient) deleteBroker(name string, propagationPolicy apis_v1.DeletionPropagation) error {
	err := c.client.Brokers(c.namespace).Delete(name, &apis_v1.DeleteOptions{PropagationPolicy: &propagationPolicy})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// ListBrokers is used to retrieve the list of broker instances
func (c *knEventingClient) ListBrokers() (*v1beta1.BrokerList, error) {
	brokerList, err := c.client.Brokers(c.namespace).List(apis_v1.ListOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	brokerListNew := brokerList.DeepCopy()
	err = updateEventingGVK(brokerListNew)
	if err != nil {
		return nil, err
	}

	brokerListNew.Items = make([]v1beta1.Broker, len(brokerList.Items))
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

// BrokerBuilder is for building the broker
type BrokerBuilder struct {
	broker *v1beta1.Broker
}

// NewBrokerBuilder for building broker object
func NewBrokerBuilder(name string) *BrokerBuilder {
	return &BrokerBuilder{broker: &v1beta1.Broker{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
	}}
}

// Namespace for broker builder
func (b *BrokerBuilder) Namespace(ns string) *BrokerBuilder {
	b.broker.Namespace = ns
	return b
}

// Build to return an instance of broker object
func (b *BrokerBuilder) Build() *v1beta1.Broker {
	return b.broker
}
