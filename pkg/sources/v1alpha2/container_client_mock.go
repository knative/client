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

package v1alpha2

import (
	"testing"

	"knative.dev/client/pkg/util/mock"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
)

// MockKnContainerSourceClient is a combine of test object and recorder
type MockKnContainerSourceClient struct {
	t         *testing.T
	recorder  *ConainterSourceRecorder
	namespace string
}

// NewMockKnContainerSourceClient returns a new mock instance which you need to record for
func NewMockKnContainerSourceClient(t *testing.T, ns ...string) *MockKnContainerSourceClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnContainerSourceClient{
		t:         t,
		recorder:  &ConainterSourceRecorder{mock.NewRecorder(t, namespace)},
		namespace: namespace,
	}
}

// Ensure that the interface is implemented
var _ KnContainerSourcesClient = &MockKnContainerSourceClient{}

// ConainterSourceRecorder is recorder for eventing objects
type ConainterSourceRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnContainerSourceClient) Recorder() *ConainterSourceRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnContainerSourceClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// CreateContainerSource records a call for CreateContainerSource with the expected error
func (sr *ConainterSourceRecorder) CreateContainerSource(binding interface{}, err error) {
	sr.r.Add("CreateContainerSource", []interface{}{binding}, []interface{}{err})
}

// CreateContainerSource performs a previously recorded action
func (c *MockKnContainerSourceClient) CreateContainerSource(binding *v1alpha2.ContainerSource) error {
	call := c.recorder.r.VerifyCall("CreateContainerSource", binding)
	return mock.ErrorOrNil(call.Result[0])
}

// GetContainerSource records a call for GetContainerSource with the expected object or error. Either binding or err should be nil
func (sr *ConainterSourceRecorder) GetContainerSource(name interface{}, binding *v1alpha2.ContainerSource, err error) {
	sr.r.Add("GetContainerSource", []interface{}{name}, []interface{}{binding, err})
}

// GetContainerSource performs a previously recorded action
func (c *MockKnContainerSourceClient) GetContainerSource(name string) (*v1alpha2.ContainerSource, error) {
	call := c.recorder.r.VerifyCall("GetContainerSource", name)
	return call.Result[0].(*v1alpha2.ContainerSource), mock.ErrorOrNil(call.Result[1])
}

// DeleteContainerSource records a call for DeleteContainerSource with the expected error (nil if none)
func (sr *ConainterSourceRecorder) DeleteContainerSource(name interface{}, err error) {
	sr.r.Add("DeleteContainerSource", []interface{}{name}, []interface{}{err})
}

// DeleteContainerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnContainerSourceClient) DeleteContainerSource(name string) error {
	call := c.recorder.r.VerifyCall("DeleteContainerSource", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListContainerSources records a call for ListContainerSources with the expected result and error (nil if none)
func (sr *ConainterSourceRecorder) ListContainerSources(bindingList *v1alpha2.ContainerSourceList, err error) {
	sr.r.Add("ListContainerSources", nil, []interface{}{bindingList, err})
}

// ListContainerSources performs a previously recorded action
func (c *MockKnContainerSourceClient) ListContainerSources() (*v1alpha2.ContainerSourceList, error) {
	call := c.recorder.r.VerifyCall("ListContainerSources")
	return call.Result[0].(*v1alpha2.ContainerSourceList), mock.ErrorOrNil(call.Result[1])
}

// UpdateContainerSource records a call for ListContainerSources with the expected result and error (nil if none)
func (sr *ConainterSourceRecorder) UpdateContainerSource(binding interface{}, err error) {
	sr.r.Add("UpdateContainerSource", []interface{}{binding}, []interface{}{err})
}

// UpdateContainerSource performs a previously recorded action
func (c *MockKnContainerSourceClient) UpdateContainerSource(binding *v1alpha2.ContainerSource) error {
	call := c.recorder.r.VerifyCall("UpdateContainerSource")
	return mock.ErrorOrNil(call.Result[0])
}

// Validate validates whether every recorded action has been called
func (sr *ConainterSourceRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
