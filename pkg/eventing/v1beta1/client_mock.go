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
	"testing"
	"time"

	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"

	"knative.dev/client/pkg/util/mock"
)

// MockKnEventingClient is a combine of test object and recorder
type MockKnEventingClient struct {
	t         *testing.T
	recorder  *EventingRecorder
	namespace string
}

// NewMockKnEventingClient returns a new mock instance which you need to record for
func NewMockKnEventingClient(t *testing.T, ns ...string) *MockKnEventingClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnEventingClient{
		t:        t,
		recorder: &EventingRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnEventingClient = &MockKnEventingClient{}

// EventingRecorder is recorder for eventing objects
type EventingRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnEventingClient) Recorder() *EventingRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnEventingClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// CreateTrigger records a call for CreatePingSource with the expected error
func (sr *EventingRecorder) CreateTrigger(trigger interface{}, err error) {
	sr.r.Add("CreateTrigger", []interface{}{trigger}, []interface{}{err})
}

// CreateTrigger performs a previously recorded action
func (c *MockKnEventingClient) CreateTrigger(trigger *v1beta1.Trigger) error {
	call := c.recorder.r.VerifyCall("CreateTrigger", trigger)
	return mock.ErrorOrNil(call.Result[0])
}

// GetTrigger records a call for GetTrigger with the expected object or error. Either trigger or err should be nil
func (sr *EventingRecorder) GetTrigger(name interface{}, trigger *v1beta1.Trigger, err error) {
	sr.r.Add("GetTrigger", []interface{}{name}, []interface{}{trigger, err})
}

// GetTrigger performs a previously recorded action
func (c *MockKnEventingClient) GetTrigger(name string) (*v1beta1.Trigger, error) {
	call := c.recorder.r.VerifyCall("GetTrigger", name)
	return call.Result[0].(*v1beta1.Trigger), mock.ErrorOrNil(call.Result[1])
}

// DeleteTrigger records a call for DeleteTrigger with the expected error (nil if none)
func (sr *EventingRecorder) DeleteTrigger(name interface{}, err error) {
	sr.r.Add("DeleteTrigger", []interface{}{name}, []interface{}{err})
}

// DeleteTrigger performs a previously recorded action, failing if non has been registered
func (c *MockKnEventingClient) DeleteTrigger(name string) error {
	call := c.recorder.r.VerifyCall("DeleteTrigger", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListTriggers records a call for ListTriggers with the expected result and error (nil if none)
func (sr *EventingRecorder) ListTriggers(triggerList *v1beta1.TriggerList, err error) {
	sr.r.Add("ListTriggers", nil, []interface{}{triggerList, err})
}

// ListTriggers performs a previously recorded action
func (c *MockKnEventingClient) ListTriggers() (*v1beta1.TriggerList, error) {
	call := c.recorder.r.VerifyCall("ListTriggers")
	return call.Result[0].(*v1beta1.TriggerList), mock.ErrorOrNil(call.Result[1])
}

// UpdateTrigger records a call for ListTriggers with the expected result and error (nil if none)
func (sr *EventingRecorder) UpdateTrigger(trigger interface{}, err error) {
	sr.r.Add("UpdateTrigger", []interface{}{trigger}, []interface{}{err})
}

// UpdateTrigger performs a previously recorded action
func (c *MockKnEventingClient) UpdateTrigger(trigger *v1beta1.Trigger) error {
	call := c.recorder.r.VerifyCall("UpdateTrigger")
	return mock.ErrorOrNil(call.Result[0])
}

// CreateBroker records a call for CreateBroker with the expected error
func (sr *EventingRecorder) CreateBroker(broker interface{}, err error) {
	sr.r.Add("CreateBroker", []interface{}{broker}, []interface{}{err})
}

// CreateBroker performs a previously recorded action
func (c *MockKnEventingClient) CreateBroker(broker *v1beta1.Broker) error {
	call := c.recorder.r.VerifyCall("CreateBroker", broker)
	return mock.ErrorOrNil(call.Result[0])
}

// GetBroker records a call for GetBroker with the expected object or error. Either trigger or err should be nil
func (sr *EventingRecorder) GetBroker(name interface{}, broker *v1beta1.Broker, err error) {
	sr.r.Add("GetBroker", []interface{}{name}, []interface{}{broker, err})
}

// GetBroker performs a previously recorded action
func (c *MockKnEventingClient) GetBroker(name string) (*v1beta1.Broker, error) {
	call := c.recorder.r.VerifyCall("GetBroker", name)
	return call.Result[0].(*v1beta1.Broker), mock.ErrorOrNil(call.Result[1])
}

// DeleteBroker records a call for DeleteBroker with the expected error (nil if none)
func (sr *EventingRecorder) DeleteBroker(name, timeout interface{}, err error) {
	sr.r.Add("DeleteBroker", []interface{}{name, timeout}, []interface{}{err})
}

// DeleteBroker performs a previously recorded action, failing if non has been registered
func (c *MockKnEventingClient) DeleteBroker(name string, timeout time.Duration) error {
	call := c.recorder.r.VerifyCall("DeleteBroker", name, timeout)
	return mock.ErrorOrNil(call.Result[0])
}

// ListBrokers records a call for ListBrokers with the expected result and error (nil if none)
func (sr *EventingRecorder) ListBrokers(brokerList *v1beta1.BrokerList, err error) {
	sr.r.Add("ListBrokers", nil, []interface{}{brokerList, err})
}

// ListBrokers performs a previously recorded action
func (c *MockKnEventingClient) ListBrokers() (*v1beta1.BrokerList, error) {
	call := c.recorder.r.VerifyCall("ListBrokers")
	return call.Result[0].(*v1beta1.BrokerList), mock.ErrorOrNil(call.Result[1])
}

// Validate validates whether every recorded action has been called
func (sr *EventingRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
