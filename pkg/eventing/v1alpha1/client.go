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

package v1alpha1

import (
	apis_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
	client_v1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kn_errors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/util"
)

// KnEventingClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnEventingClient interface {
	// Namespace in which this client is operating for
	Namespace() string
	// CreateTrigger is used to create an instance of trigger
	CreateTrigger(trigger *v1alpha1.Trigger) error
	// DeleteTrigger is used to delete an instance of trigger
	DeleteTrigger(name string) error
	// GetTrigger is used to get an instance of trigger
	GetTrigger(name string) (*v1alpha1.Trigger, error)
	// ListTrigger returns list of trigger CRDs
	ListTriggers() (*v1alpha1.TriggerList, error)
	// UpdateTrigger is used to update an instance of trigger
	UpdateTrigger(trigger *v1alpha1.Trigger) error
}

// KnEventingClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type knEventingClient struct {
	client    client_v1alpha1.EventingV1alpha1Interface
	namespace string
}

// NewKnEventingClient is to invoke Eventing Sources Client API to create object
func NewKnEventingClient(client client_v1alpha1.EventingV1alpha1Interface, namespace string) KnEventingClient {
	return &knEventingClient{
		client:    client,
		namespace: namespace,
	}
}

//CreateTrigger is used to create an instance of trigger
func (c *knEventingClient) CreateTrigger(trigger *v1alpha1.Trigger) error {
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
func (c *knEventingClient) GetTrigger(name string) (*v1alpha1.Trigger, error) {
	trigger, err := c.client.Triggers(c.namespace).Get(name, apis_v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	return trigger, nil
}

func (c *knEventingClient) ListTriggers() (*v1alpha1.TriggerList, error) {
	triggerList, err := c.client.Triggers(c.namespace).List(apis_v1.ListOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	triggerListNew := triggerList.DeepCopy()
	err = updateTriggerGvk(triggerListNew)
	if err != nil {
		return nil, err
	}

	triggerListNew.Items = make([]v1alpha1.Trigger, len(triggerList.Items))
	for idx, trigger := range triggerList.Items {
		triggerClone := trigger.DeepCopy()
		err := updateTriggerGvk(triggerClone)
		if err != nil {
			return nil, err
		}
		triggerListNew.Items[idx] = *triggerClone
	}
	return triggerListNew, nil
}

//CreateTrigger is used to create an instance of trigger
func (c *knEventingClient) UpdateTrigger(trigger *v1alpha1.Trigger) error {
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

// update with the v1alpha1 group + version
func updateTriggerGvk(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, v1alpha1.SchemeGroupVersion, scheme.Scheme)
}

// TriggerBuilder is for building the trigger
type TriggerBuilder struct {
	trigger *v1alpha1.Trigger
}

// NewTriggerBuilder for building trigger object
func NewTriggerBuilder(name string) *TriggerBuilder {
	return &TriggerBuilder{trigger: &v1alpha1.Trigger{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
	}}
}

// NewTriggerBuilderFromExisting for building the object from existing Trigger object
func NewTriggerBuilderFromExisting(trigger *v1alpha1.Trigger) *TriggerBuilder {
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
		meta_v1.SetMetaDataAnnotation(&b.trigger.ObjectMeta, v1alpha1.InjectionAnnotation, "enabled")
	} else {
		if meta_v1.HasAnnotation(b.trigger.ObjectMeta, v1alpha1.InjectionAnnotation) {
			delete(b.trigger.ObjectMeta.Annotations, v1alpha1.InjectionAnnotation)
		}
	}
	return b
}

func (b *TriggerBuilder) Filters(filters map[string]string) *TriggerBuilder {
	if len(filters) == 0 {
		b.trigger.Spec.Filter = &v1alpha1.TriggerFilter{}
		return b
	}
	filter := b.trigger.Spec.Filter
	if filter == nil {
		filter = &v1alpha1.TriggerFilter{}
		b.trigger.Spec.Filter = filter
	}
	filter.Attributes = &v1alpha1.TriggerFilterAttributes{}
	for k, v := range filters {
		(*filter.Attributes)[k] = v
	}
	return b
}

// Build to return an instance of trigger object
func (b *TriggerBuilder) Build() *v1alpha1.Trigger {
	return b.trigger
}
