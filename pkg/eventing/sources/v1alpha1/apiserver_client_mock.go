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
	"testing"

	"knative.dev/eventing/pkg/apis/sources/v1alpha1"

	"knative.dev/client/pkg/util/mock"
)

type MockKnApiServerSourceClient struct {
	t         *testing.T
	recorder  *ApiServerSourcesRecorder
	namespace string
}

// NewMockKnApiServerSourceClient returns a new mock instance which you need to record for
func NewMockKnApiServerSourceClient(t *testing.T, ns ...string) *MockKnApiServerSourceClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnApiServerSourceClient{
		t:        t,
		recorder: &ApiServerSourcesRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnApiServerSourcesClient = &MockKnApiServerSourceClient{}

// recorder for service
type ApiServerSourcesRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnApiServerSourceClient) Recorder() *ApiServerSourcesRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnApiServerSourceClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// GetApiServerSource records a call for GetApiServerSource with the expected object or error. Either apiServerSource or err should be nil
func (sr *ApiServerSourcesRecorder) GetApiServerSource(name interface{}, apiServerSource *v1alpha1.ApiServerSource, err error) {
	sr.r.Add("GetApiServerSource", []interface{}{name}, []interface{}{apiServerSource, err})
}

// GetApiServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnApiServerSourceClient) GetApiServerSource(name string) (*v1alpha1.ApiServerSource, error) {
	call := c.recorder.r.VerifyCall("GetApiServerSource", name)
	return call.Result[0].(*v1alpha1.ApiServerSource), mock.ErrorOrNil(call.Result[1])
}

// CreateApiServerSource records a call for CreateApiServerSource with the expected error
func (sr *ApiServerSourcesRecorder) CreateApiServerSource(apiServerSource interface{}, err error) {
	sr.r.Add("CreateApiServerSource", []interface{}{apiServerSource}, []interface{}{err})
}

// CreateApiServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnApiServerSourceClient) CreateApiServerSource(apiServerSource *v1alpha1.ApiServerSource) error {
	call := c.recorder.r.VerifyCall("CreateApiServerSource", apiServerSource)
	return mock.ErrorOrNil(call.Result[0])
}

// UpdateApiServerSource records a call for UpdateApiServerSource with the expected error (nil if none)
func (sr *ApiServerSourcesRecorder) UpdateApiServerSource(apiServerSource interface{}, err error) {
	sr.r.Add("UpdateApiServerSource", []interface{}{apiServerSource}, []interface{}{err})
}

// UpdateApiServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnApiServerSourceClient) UpdateApiServerSource(apiServerSource *v1alpha1.ApiServerSource) error {
	call := c.recorder.r.VerifyCall("UpdateApiServerSource", apiServerSource)
	return mock.ErrorOrNil(call.Result[0])
}

// UpdateApiServerSource records a call for DeleteApiServerSource with the expected error (nil if none)
func (sr *ApiServerSourcesRecorder) DeleteApiServerSource(name interface{}, err error) {
	sr.r.Add("DeleteApiServerSource", []interface{}{name}, []interface{}{err})
}

// DeleteApiServerSource performs a previously recorded action, failing if non has been registered
func (c *MockKnApiServerSourceClient) DeleteApiServerSource(name string) error {
	call := c.recorder.r.VerifyCall("DeleteApiServerSource", name)
	return mock.ErrorOrNil(call.Result[0])
}

// Validates validates whether every recorded action has been called
func (sr *ApiServerSourcesRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
