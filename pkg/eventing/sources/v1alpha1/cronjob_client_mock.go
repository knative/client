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

type MockKnCronJobSourceClient struct {
	t         *testing.T
	recorder  *CronJobSourcesRecorder
	namespace string
}

// NewMockKnCronJobSourceClient returns a new mock instance which you need to record for
func NewMockKnCronJobSourceClient(t *testing.T, ns ...string) *MockKnCronJobSourceClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnCronJobSourceClient{
		t:        t,
		recorder: &CronJobSourcesRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnCronJobSourcesClient = &MockKnCronJobSourceClient{}

// recorder for service
type CronJobSourcesRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnCronJobSourceClient) Recorder() *CronJobSourcesRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnCronJobSourceClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// CreateCronJobSource records a call for CreateCronJobSource with the expected error
func (sr *CronJobSourcesRecorder) CreateCronJobSource(cronjobSource interface{}, err error) {
	sr.r.Add("CreateCronJobSource", []interface{}{cronjobSource}, []interface{}{err})
}

// CreateCronJobSource performs a previously recorded action, failing if non has been registered
func (c *MockKnCronJobSourceClient) CreateCronJobSource(cronjobSource *v1alpha1.CronJobSource) error {
	call := c.recorder.r.VerifyCall("CreateCronJobSource", cronjobSource)
	return mock.ErrorOrNil(call.Result[0])
}

// GetCronJobSource records a call for GetCronJobSource with the expected object or error. Either cronjobsource or err should be nil
func (sr *CronJobSourcesRecorder) GetCronJobSource(name interface{}, cronjobSource *v1alpha1.CronJobSource, err error) {
	sr.r.Add("GetCronJobSource", []interface{}{name}, []interface{}{cronjobSource, err})
}

// GetCronJobSource performs a previously recorded action, failing if non has been registered
func (c *MockKnCronJobSourceClient) GetCronJobSource(name string) (*v1alpha1.CronJobSource, error) {
	call := c.recorder.r.VerifyCall("GetCronJobSource", name)
	return call.Result[0].(*v1alpha1.CronJobSource), mock.ErrorOrNil(call.Result[1])
}

// UpdateCronJobSource records a call for UpdateCronJobSource with the expected error (nil if none)
func (sr *CronJobSourcesRecorder) UpdateCronJobSource(cronjobSource interface{}, err error) {
	sr.r.Add("UpdateCronJobSource", []interface{}{cronjobSource}, []interface{}{err})
}

// UpdateCronJobSource performs a previously recorded action, failing if non has been registered
func (c *MockKnCronJobSourceClient) UpdateCronJobSource(cronjobSource *v1alpha1.CronJobSource) error {
	call := c.recorder.r.VerifyCall("UpdateCronJobSource", cronjobSource)
	return mock.ErrorOrNil(call.Result[0])
}

// UpdateCronJobSource records a call for DeleteCronJobSource with the expected error (nil if none)
func (sr *CronJobSourcesRecorder) DeleteCronJobSource(name interface{}, err error) {
	sr.r.Add("DeleteCronJobSource", []interface{}{name}, []interface{}{err})
}

// DeleteCronJobSource performs a previously recorded action, failing if non has been registered
func (c *MockKnCronJobSourceClient) DeleteCronJobSource(name string) error {
	call := c.recorder.r.VerifyCall("DeleteCronJobSource", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListCronJobSource records a call for ListCronJobSource with the expected error (nil if none)
func (sr *CronJobSourcesRecorder) ListCronJobSource(cronJobSourceList *v1alpha1.CronJobSourceList, err error) {
	sr.r.Add("ListCronJobSource", []interface{}{}, []interface{}{cronJobSourceList, err})
}

// ListCronJobSource performs a previously recorded action, failing if non has been registered
func (c *MockKnCronJobSourceClient) ListCronJobSource() (*v1alpha1.CronJobSourceList, error) {
	call := c.recorder.r.VerifyCall("ListCronJobSource")
	return call.Result[0].(*v1alpha1.CronJobSourceList), mock.ErrorOrNil(call.Result[1])
}

// Validates validates whether every recorded action has been called
func (sr *CronJobSourcesRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
