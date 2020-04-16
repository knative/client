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

// MockKnAPIServerSourceClient for mocking the client
type MockKnAPIServerSourceClient struct {
	t         *testing.T
	recorder  *APIServerSourcesRecorder
	namespace string
}

// NewMockKnAPIServerSourceClient returns a new mock instance which you need to record for
func NewMockKnAPIServerSourceClient(t *testing.T, ns ...string) *MockKnAPIServerSourceClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnAPIServerSourceClient{
		t:        t,
		recorder: &APIServerSourcesRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnAPIServerSourcesClient = &MockKnAPIServerSourceClient{}

// APIServerSourcesRecorder for recording actions on source
type APIServerSourcesRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnAPIServerSourceClient) Recorder() *APIServerSourcesRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnAPIServerSourceClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// GetAPIServerSource records a call for GetApiServerSource with the expected object or error. Either apiServerSource or err should be nil
func (sr *APIServerSourcesRecorder) GetAPIServerSource(name interface{}, apiServerSource *v1alpha2.ApiServerSource, err error) {
	sr.r.Add("GetApiServerSource", []interface{}{name}, []interface{}{apiServerSource, err})
}

// GetAPIServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnAPIServerSourceClient) GetAPIServerSource(name string) (*v1alpha2.ApiServerSource, error) {
	call := c.recorder.r.VerifyCall("GetApiServerSource", name)
	return call.Result[0].(*v1alpha2.ApiServerSource), mock.ErrorOrNil(call.Result[1])
}

// CreateAPIServerSource records a call for CreateApiServerSource with the expected error
func (sr *APIServerSourcesRecorder) CreateAPIServerSource(apiServerSource interface{}, err error) {
	sr.r.Add("CreateApiServerSource", []interface{}{apiServerSource}, []interface{}{err})
}

// CreateAPIServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnAPIServerSourceClient) CreateAPIServerSource(apiServerSource *v1alpha2.ApiServerSource) error {
	call := c.recorder.r.VerifyCall("CreateApiServerSource", apiServerSource)
	return mock.ErrorOrNil(call.Result[0])
}

// UpdateAPIServerSource records a call for UpdateAPIServerSource with the expected error (nil if none)
func (sr *APIServerSourcesRecorder) UpdateAPIServerSource(apiServerSource interface{}, err error) {
	sr.r.Add("UpdateAPIServerSource", []interface{}{apiServerSource}, []interface{}{err})
}

// UpdateAPIServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnAPIServerSourceClient) UpdateAPIServerSource(apiServerSource *v1alpha2.ApiServerSource) error {
	call := c.recorder.r.VerifyCall("UpdateAPIServerSource", apiServerSource)
	return mock.ErrorOrNil(call.Result[0])
}

// DeleteAPIServerSource records a call for DeleteAPIServerSource with the expected error (nil if none)
func (sr *APIServerSourcesRecorder) DeleteAPIServerSource(name interface{}, err error) {
	sr.r.Add("DeleteAPIServerSource", []interface{}{name}, []interface{}{err})
}

// DeleteAPIServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnAPIServerSourceClient) DeleteAPIServerSource(name string) error {
	call := c.recorder.r.VerifyCall("DeleteAPIServerSource", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListAPIServerSource records a call for ListAPIServerSource with the expected error (nil if none)
func (sr *APIServerSourcesRecorder) ListAPIServerSource(apiJobSourceList *v1alpha2.ApiServerSourceList, err error) {
	sr.r.Add("ListAPIServerSource", []interface{}{}, []interface{}{apiJobSourceList, err})
}

// ListAPIServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnAPIServerSourceClient) ListAPIServerSource() (*v1alpha2.ApiServerSourceList, error) {
	call := c.recorder.r.VerifyCall("ListAPIServerSource")
	return call.Result[0].(*v1alpha2.ApiServerSourceList), mock.ErrorOrNil(call.Result[1])
}

// Validate validates whether every recorded action has been called
func (sr *APIServerSourcesRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
