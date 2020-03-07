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

package v1alpha2

import (
	"testing"

	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"

	"knative.dev/client/pkg/util/mock"
)

// MockKnSinkBindingClient is a combine of test object and recorder
type MockKnSinkBindingClient struct {
	t         *testing.T
	recorder  *EventingRecorder
	namespace string
}

// NewMockKnSinkBindingClient returns a new mock instance which you need to record for
func NewMockKnSinkBindingClient(t *testing.T, ns ...string) *MockKnSinkBindingClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnSinkBindingClient{
		t:        t,
		recorder: &EventingRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnSinkBindingClient = &MockKnSinkBindingClient{}

// EventingRecorder is recorder for eventing objects
type EventingRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnSinkBindingClient) Recorder() *EventingRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnSinkBindingClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// CreateSinkBinding records a call for CreateSinkBinding with the expected error
func (sr *EventingRecorder) CreateSinkBinding(binding interface{}, err error) {
	sr.r.Add("CreateSinkBinding", []interface{}{binding}, []interface{}{err})
}

// CreateSinkBinding performs a previously recorded action
func (c *MockKnSinkBindingClient) CreateSinkBinding(binding *v1alpha2.SinkBinding) error {
	call := c.recorder.r.VerifyCall("CreateSinkBinding", binding)
	return mock.ErrorOrNil(call.Result[0])
}

// GetSinkBinding records a call for GetSinkBinding with the expected object or error. Either binding or err should be nil
func (sr *EventingRecorder) GetSinkBinding(name interface{}, binding *v1alpha2.SinkBinding, err error) {
	sr.r.Add("GetSinkBinding", []interface{}{name}, []interface{}{binding, err})
}

// GetSinkBinding performs a previously recorded action
func (c *MockKnSinkBindingClient) GetSinkBinding(name string) (*v1alpha2.SinkBinding, error) {
	call := c.recorder.r.VerifyCall("GetSinkBinding", name)
	return call.Result[0].(*v1alpha2.SinkBinding), mock.ErrorOrNil(call.Result[1])
}

// DeleteSinkBinding records a call for DeleteSinkBinding with the expected error (nil if none)
func (sr *EventingRecorder) DeleteSinkBinding(name interface{}, err error) {
	sr.r.Add("DeleteSinkBinding", []interface{}{name}, []interface{}{err})
}

// DeleteSinkBinding performs a previously recorded action, failing if non has been registered
func (c *MockKnSinkBindingClient) DeleteSinkBinding(name string) error {
	call := c.recorder.r.VerifyCall("DeleteSinkBinding", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListSinkBindings records a call for ListSinkBindings with the expected result and error (nil if none)
func (sr *EventingRecorder) ListSinkBindings(bindingList *v1alpha2.SinkBindingList, err error) {
	sr.r.Add("ListSinkBindings", nil, []interface{}{bindingList, err})
}

// ListSinkBindings performs a previously recorded action
func (c *MockKnSinkBindingClient) ListSinkBindings() (*v1alpha2.SinkBindingList, error) {
	call := c.recorder.r.VerifyCall("ListSinkBindings")
	return call.Result[0].(*v1alpha2.SinkBindingList), mock.ErrorOrNil(call.Result[1])
}

// UpdateSinkBinding records a call for ListSinkBindings with the expected result and error (nil if none)
func (sr *EventingRecorder) UpdateSinkBinding(binding interface{}, err error) {
	sr.r.Add("UpdateSinkBinding", []interface{}{binding}, []interface{}{err})
}

// UpdateSinkBinding performs a previously recorded action
func (c *MockKnSinkBindingClient) UpdateSinkBinding(binding *v1alpha2.SinkBinding) error {
	call := c.recorder.r.VerifyCall("UpdateSinkBinding")
	return mock.ErrorOrNil(call.Result[0])
}

// Validate validates whether every recorded action has been called
func (sr *EventingRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
