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

package v1

import (
	"testing"
	"time"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/util/mock"
	"knative.dev/client/pkg/wait"
)

type MockKnServingClient struct {
	t        *testing.T
	recorder *ServingRecorder
}

// NewMockKnServiceClient returns a new mock instance which you need to record for
func NewMockKnServiceClient(t *testing.T, ns ...string) *MockKnServingClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnServingClient{
		t:        t,
		recorder: &ServingRecorder{mock.NewRecorder(t, namespace)},
	}
}

// recorder for service
type ServingRecorder struct {
	r *mock.Recorder
}

// Get the record to start for the recorder
func (c *MockKnServingClient) Recorder() *ServingRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnServingClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// Get Service
func (sr *ServingRecorder) GetService(name interface{}, service *servingv1.Service, err error) {
	sr.r.Add("GetService", []interface{}{name}, []interface{}{service, err})
}

func (c *MockKnServingClient) GetService(name string) (*servingv1.Service, error) {
	call := c.recorder.r.VerifyCall("GetService", name)
	return call.Result[0].(*servingv1.Service), mock.ErrorOrNil(call.Result[1])
}

// List services
func (sr *ServingRecorder) ListServices(opts interface{}, serviceList *servingv1.ServiceList, err error) {
	sr.r.Add("ListServices", []interface{}{opts}, []interface{}{serviceList, err})
}

func (c *MockKnServingClient) ListServices(opts ...ListConfig) (*servingv1.ServiceList, error) {
	call := c.recorder.r.VerifyCall("ListServices", opts)
	return call.Result[0].(*servingv1.ServiceList), mock.ErrorOrNil(call.Result[1])
}

// Create a new service
func (sr *ServingRecorder) CreateService(service interface{}, err error) {
	sr.r.Add("CreateService", []interface{}{service}, []interface{}{err})
}

func (c *MockKnServingClient) CreateService(service *servingv1.Service) error {
	call := c.recorder.r.VerifyCall("CreateService", service)
	return mock.ErrorOrNil(call.Result[0])
}

// Update the given service
func (sr *ServingRecorder) UpdateService(service interface{}, err error) {
	sr.r.Add("UpdateService", []interface{}{service}, []interface{}{err})
}

func (c *MockKnServingClient) UpdateService(service *servingv1.Service) error {
	call := c.recorder.r.VerifyCall("UpdateService", service)
	return mock.ErrorOrNil(call.Result[0])
}

// Delegate to shared retry method
func (c *MockKnServingClient) UpdateServiceWithRetry(name string, updateFunc ServiceUpdateFunc, maxRetry int) error {
	return updateServiceWithRetry(c, name, updateFunc, maxRetry)
}

// Delete a service by name
func (sr *ServingRecorder) DeleteService(name, timeout interface{}, err error) {
	sr.r.Add("DeleteService", []interface{}{name, timeout}, []interface{}{err})
}

func (c *MockKnServingClient) DeleteService(name string, timeout time.Duration) error {
	call := c.recorder.r.VerifyCall("DeleteService", name, timeout)
	return mock.ErrorOrNil(call.Result[0])
}

// Wait for a service to become ready, but not longer than provided timeout
func (sr *ServingRecorder) WaitForService(name interface{}, timeout interface{}, callback interface{}, err error, duration time.Duration) {
	sr.r.Add("WaitForService", []interface{}{name, timeout, callback}, []interface{}{err, duration})
}

func (c *MockKnServingClient) WaitForService(name string, timeout time.Duration, msgCallback wait.MessageCallback) (error, time.Duration) {
	call := c.recorder.r.VerifyCall("WaitForService", name, timeout, msgCallback)
	return mock.ErrorOrNil(call.Result[0]), call.Result[1].(time.Duration)
}

// Get a revision by name
func (sr *ServingRecorder) GetRevision(name interface{}, revision *servingv1.Revision, err error) {
	sr.r.Add("GetRevision", []interface{}{name}, []interface{}{revision, err})
}

func (c *MockKnServingClient) GetRevision(name string) (*servingv1.Revision, error) {
	call := c.recorder.r.VerifyCall("GetRevision", name)
	return call.Result[0].(*servingv1.Revision), mock.ErrorOrNil(call.Result[1])
}

// List revisions
func (sr *ServingRecorder) ListRevisions(opts interface{}, revisionList *servingv1.RevisionList, err error) {
	sr.r.Add("ListRevisions", []interface{}{opts}, []interface{}{revisionList, err})
}

func (c *MockKnServingClient) ListRevisions(opts ...ListConfig) (*servingv1.RevisionList, error) {
	call := c.recorder.r.VerifyCall("ListRevisions", opts)
	return call.Result[0].(*servingv1.RevisionList), mock.ErrorOrNil(call.Result[1])
}

// Delete a revision
func (sr *ServingRecorder) DeleteRevision(name, timeout interface{}, err error) {
	sr.r.Add("DeleteRevision", []interface{}{name, timeout}, []interface{}{err})
}

func (c *MockKnServingClient) DeleteRevision(name string, timeout time.Duration) error {
	call := c.recorder.r.VerifyCall("DeleteRevision", name, timeout)
	return mock.ErrorOrNil(call.Result[0])
}

// Get a route by its unique name
func (sr *ServingRecorder) GetRoute(name interface{}, route *servingv1.Route, err error) {
	sr.r.Add("GetRoute", []interface{}{name}, []interface{}{route, err})
}

func (c *MockKnServingClient) GetRoute(name string) (*servingv1.Route, error) {
	call := c.recorder.r.VerifyCall("GetRoute", name)
	return call.Result[0].(*servingv1.Route), mock.ErrorOrNil(call.Result[1])

}

// List routes
func (sr *ServingRecorder) ListRoutes(opts interface{}, routeList *servingv1.RouteList, err error) {
	sr.r.Add("ListRoutes", []interface{}{opts}, []interface{}{routeList, err})
}

func (c *MockKnServingClient) ListRoutes(opts ...ListConfig) (*servingv1.RouteList, error) {
	call := c.recorder.r.VerifyCall("ListRoutes", opts)
	return call.Result[0].(*servingv1.RouteList), mock.ErrorOrNil(call.Result[1])
}

// GetConfiguration records a call to GetConfiguration with possible return values
func (sr *ServingRecorder) GetConfiguration(name string, config *servingv1.Configuration, err error) {
	sr.r.Add("GetConfiguration", []interface{}{name}, []interface{}{config, err})

}

// Check for the base revision
func (c *MockKnServingClient) GetBaseRevision(service *servingv1.Service) (*servingv1.Revision, error) {
	return getBaseRevision(c, service)
}

// GetConfiguration returns a configuration looked up by name
func (c *MockKnServingClient) GetConfiguration(name string) (*servingv1.Configuration, error) {
	call := c.recorder.r.VerifyCall("GetConfiguration", name)
	return call.Result[0].(*servingv1.Configuration), mock.ErrorOrNil(call.Result[1])
}

// Check that every recorded method has been called
func (sr *ServingRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}

// HasLabelSelector returns a comparable which can be used for asserting that list methods are called
// with the appropriate label selector
func HasLabelSelector(keyAndValues ...string) func(t *testing.T, a interface{}) {
	return func(t *testing.T, a interface{}) {
		lc := a.([]ListConfig)
		listConfigCollector := listConfigCollector{
			Labels: make(labels.Set),
			Fields: make(fields.Set),
		}
		lc[0](&listConfigCollector)
		for i := 0; i < len(keyAndValues); i += 2 {
			assert.Equal(t, listConfigCollector.Labels[keyAndValues[i]], keyAndValues[i+1])
		}
	}
}

// HasFieldSelector returns a comparable which can be used for asserting that list methods are called
// with the appropriate field selectors
func HasFieldSelector(keyAndValues ...string) func(t *testing.T, a interface{}) {
	return func(t *testing.T, a interface{}) {
		lc := a.([]ListConfig)
		listConfigCollector := listConfigCollector{
			Labels: make(labels.Set),
			Fields: make(fields.Set),
		}
		lc[0](&listConfigCollector)
		for i := 0; i < len(keyAndValues); i += 2 {
			assert.Equal(t, listConfigCollector.Fields[keyAndValues[i]], keyAndValues[i+1])
		}
	}
}

// HasSelector returns a comparable which can be used for asserting that list methods are called
// with the appropriate label and field selectors
func HasSelector(labelKeysAndValues []string, fieldKeysAndValue []string) func(t *testing.T, a interface{}) {
	return func(t *testing.T, a interface{}) {
		HasLabelSelector(labelKeysAndValues...)(t, a)
		HasFieldSelector(fieldKeysAndValue...)(t, a)
	}
}
