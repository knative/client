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

	kn_errors "knative.dev/client/pkg/errors"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	client_v1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
)

// KnEventingClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnEventingClient interface {
	// Namespace in which this client is operating for
	Namespace() string
	// CreateTrigger is used to create an instance of trigger
	CreateTrigger(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error)
	// DeleteTrigger is used to delete an instance of trigger
	DeleteTrigger(name string) error
	// GetTrigger is used to get an instance of trigger
	GetTrigger(name string) (*v1alpha1.Trigger, error)
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
func (c *knEventingClient) CreateTrigger(trigger *v1alpha1.Trigger) (*v1alpha1.Trigger, error) {
	trigger, err := c.client.Triggers(c.namespace).Create(trigger)
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	return trigger, nil
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

// Return the client's namespace
func (c *knEventingClient) Namespace() string {
	return c.namespace
}
