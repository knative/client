// Copyright © 2022 The Knative Authors
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

package v1beta2

import (
	"context"

	apis_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kn_errors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/util"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	eventingv1beta2 "knative.dev/eventing/pkg/apis/eventing/v1beta2"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
	beta1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1beta2"
	"knative.dev/pkg/apis"
	v1 "knative.dev/pkg/apis/duck/v1"
)

// KnEventingV1Beta2Client to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnEventingV1Beta2Client interface {
	// Namespace in which this client is operating for
	Namespace() string
	// ListEventtypes is used to list eventtypes
	ListEventtypes(ctx context.Context) (*eventingv1beta2.EventTypeList, error)
	// GetEventtype is used to describe an eventtype
	GetEventtype(ctx context.Context, name string) (*eventingv1beta2.EventType, error)
	// CreateEventtype is used to create an eventtype
	CreateEventtype(ctx context.Context, eventtype *eventingv1beta2.EventType) error
	// DeleteEventtype is used to delete an eventtype
	DeleteEventtype(ctx context.Context, name string) error
}

// KnEventingV1Beta2Client is a client for eventing v1beta2 resources
type knEventingV1Beta1Client struct {
	client    beta1.EventingV1beta2Interface
	namespace string
}

// NewKnEventingV1Beta2Client is to invoke Eventing Types Client API to create object
func NewKnEventingV1Beta2Client(client beta1.EventingV1beta2Interface, namespace string) KnEventingV1Beta2Client {
	return &knEventingV1Beta1Client{
		client:    client,
		namespace: namespace,
	}
}

func updateEventingBeta1GVK(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, eventingv1beta2.SchemeGroupVersion, scheme.Scheme)
}

func (c *knEventingV1Beta1Client) Namespace() string {
	return c.namespace
}

func (c *knEventingV1Beta1Client) ListEventtypes(ctx context.Context) (*eventingv1beta2.EventTypeList, error) {
	eventTypeList, err := c.client.EventTypes(c.namespace).List(ctx, apis_v1.ListOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	listNew := eventTypeList.DeepCopy()
	err = updateEventingBeta1GVK(listNew)
	if err != nil {
		return nil, err
	}

	listNew.Items = make([]eventingv1beta2.EventType, len(eventTypeList.Items))
	for idx, eventType := range eventTypeList.Items {
		clone := eventType.DeepCopy()
		err := updateEventingBeta1GVK(clone)
		if err != nil {
			return nil, err
		}
		listNew.Items[idx] = *clone
	}
	return listNew, nil
}

func (c *knEventingV1Beta1Client) GetEventtype(ctx context.Context, name string) (*eventingv1beta2.EventType, error) {
	eventType, err := c.client.EventTypes(c.namespace).Get(ctx, name, apis_v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	err = updateEventingBeta1GVK(eventType)
	if err != nil {
		return nil, err
	}
	return eventType, nil
}

func (c *knEventingV1Beta1Client) DeleteEventtype(ctx context.Context, name string) error {
	err := c.client.EventTypes(c.namespace).Delete(ctx, name, apis_v1.DeleteOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

func (c *knEventingV1Beta1Client) CreateEventtype(ctx context.Context, eventtype *eventingv1beta2.EventType) error {
	_, err := c.client.EventTypes(c.namespace).Create(ctx, eventtype, apis_v1.CreateOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}
	return nil
}

// EventtypeBuilder is for building the eventtype
type EventtypeBuilder struct {
	eventtype *eventingv1beta2.EventType
}

// NewEventtypeBuilder for building eventtype object
func NewEventtypeBuilder(name string) *EventtypeBuilder {
	return &EventtypeBuilder{eventtype: &eventingv1beta2.EventType{
		ObjectMeta: apis_v1.ObjectMeta{
			Name: name,
		},
	}}
}

// WithGvk add the GVK coordinates for read tests
func (e *EventtypeBuilder) WithGvk() *EventtypeBuilder {
	_ = updateEventingBeta1GVK(e.eventtype)
	return e
}

// Namespace for eventtype builder
func (e *EventtypeBuilder) Namespace(ns string) *EventtypeBuilder {
	e.eventtype.Namespace = ns
	return e
}

// Type for eventtype builder
func (e *EventtypeBuilder) Type(ceType string) *EventtypeBuilder {
	e.eventtype.Spec.Type = ceType
	return e
}

// Source for eventtype builder
func (e *EventtypeBuilder) Source(source *apis.URL) *EventtypeBuilder {
	e.eventtype.Spec.Source = source
	return e
}

// Broker for eventtype builder
func (e *EventtypeBuilder) Broker(broker string) *EventtypeBuilder {
	e.eventtype.Spec.Reference = &v1.KReference{
		APIVersion: eventingv1.SchemeGroupVersion.String(),
		Kind:       "Broker",
		Name:       broker,
	}
	return e
}

// Reference for eventtype builder
func (e *EventtypeBuilder) Reference(ref *v1.KReference) *EventtypeBuilder {
	e.eventtype.Spec.Reference = ref
	return e
}

// Build to return an instance of eventtype object
func (e *EventtypeBuilder) Build() *eventingv1beta2.EventType {
	return e.eventtype
}
