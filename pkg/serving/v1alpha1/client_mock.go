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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

type Recorder struct {
	t *testing.T

	// List of recorded calls in order
	recordedCalls map[string][]apiMethodCall
}

type MockKnClient struct {
	t        *testing.T
	recorder Recorder
}

// Recorded method call
type apiMethodCall struct {
	args   []interface{}
	result []interface{}
}

// NewMockKnClient returns a new mock instance which you need to record for
func NewMockKnClient(t *testing.T) *MockKnClient {
	return &MockKnClient{
		t: t,
		recorder: Recorder{
			t:             t,
			recordedCalls: make(map[string][]apiMethodCall),
		},
	}
}

// Get the record to start for the recorder
func (c *MockKnClient) Recorder() *Recorder {
	return &c.recorder
}

// any() can be used in recording to not check for the argument
func Any() func(t *testing.T, a interface{}) {
	return func(t *testing.T, a interface{}) {}
}

// Get Service
func (r *Recorder) GetService(name interface{}, service *v1alpha1.Service, err error) {
	r.add("GetService", apiMethodCall{[]interface{}{name}, []interface{}{service, err}})
}

func (c *MockKnClient) GetService(name string) (*v1alpha1.Service, error) {
	call := c.getCall("GetService")
	c.verifyArgs(call, name)
	return call.result[0].(*v1alpha1.Service), errorOrNil(call.result[1])
}

// List services
func (r *Recorder) ListServices(opts interface{}, serviceList *v1alpha1.ServiceList, err error) {
	r.add("ListServices", apiMethodCall{[]interface{}{opts}, []interface{}{serviceList, err}})
}

func (c *MockKnClient) ListServices(opts ...ListConfig) (*v1alpha1.ServiceList, error) {
	call := c.getCall("ListServices")
	c.verifyArgs(call, opts)
	return call.result[0].(*v1alpha1.ServiceList), errorOrNil(call.result[1])
}

// Create a new service
func (r *Recorder) CreateService(service interface{}, err error) {
	r.add("CreateService", apiMethodCall{[]interface{}{service}, []interface{}{err}})
}

func (c *MockKnClient) CreateService(service *v1alpha1.Service) error {
	call := c.getCall("CreateService")
	c.verifyArgs(call, service)
	return errorOrNil(call.result[0])
}

// Update the given service
func (r *Recorder) UpdateService(service interface{}, err error) {
	r.add("UpdateService", apiMethodCall{[]interface{}{service}, []interface{}{err}})
}

func (c *MockKnClient) UpdateService(service *v1alpha1.Service) error {
	call := c.getCall("UpdateService")
	c.verifyArgs(call, service)
	return errorOrNil(call.result[0])
}

// Delete a service by name
func (r *Recorder) DeleteService(name interface{}, err error) {
	r.add("DeleteService", apiMethodCall{[]interface{}{name}, []interface{}{err}})
}

func (c *MockKnClient) DeleteService(name string) error {
	call := c.getCall("DeleteService")
	c.verifyArgs(call, name)
	return errorOrNil(call.result[0])
}

// Wait for a service to become ready, but not longer than provided timeout
func (r *Recorder) WaitForService(name interface{}, timeout interface{}, err error) {
	r.add("WaitForService", apiMethodCall{[]interface{}{name}, []interface{}{err}})
}

func (c *MockKnClient) WaitForService(name string, timeout time.Duration) error {
	call := c.getCall("WaitForService")
	c.verifyArgs(call, name)
	return errorOrNil(call.result[0])
}

// Get a revision by name
func (r *Recorder) GetRevision(name interface{}, revision *v1alpha1.Revision, err error) {
	r.add("GetRevision", apiMethodCall{[]interface{}{name}, []interface{}{revision, err}})
}

func (c *MockKnClient) GetRevision(name string) (*v1alpha1.Revision, error) {
	call := c.getCall("GetRevision")
	c.verifyArgs(call, name)
	return call.result[0].(*v1alpha1.Revision), errorOrNil(call.result[1])
}

// List revisions
func (r *Recorder) ListRevisions(opts interface{}, revisionList *v1alpha1.RevisionList, err error) {
	r.add("ListRevisions", apiMethodCall{[]interface{}{opts}, []interface{}{revisionList, err}})
}

func (c *MockKnClient) ListRevisions(opts ...ListConfig) (*v1alpha1.RevisionList, error) {
	call := c.getCall("ListRevisions")
	c.verifyArgs(call, opts)
	return call.result[0].(*v1alpha1.RevisionList), errorOrNil(call.result[1])
}

// Delete a revision
func (r *Recorder) DeleteRevision(name interface{}, err error) {
	r.add("DeleteRevision", apiMethodCall{[]interface{}{name}, []interface{}{err}})
}

func (c *MockKnClient) DeleteRevision(name string) error {
	call := c.getCall("DeleteRevision")
	c.verifyArgs(call, name)
	return errorOrNil(call.result[0])

}

// Get a route by its unique name
func (r *Recorder) GetRoute(name interface{}, route *v1alpha1.Route, err error) {
	r.add("GetRoute", apiMethodCall{[]interface{}{name}, []interface{}{route, err}})
}

func (c *MockKnClient) GetRoute(name string) (*v1alpha1.Route, error) {
	call := c.getCall("GetRoute")
	c.verifyArgs(call, name)
	return call.result[0].(*v1alpha1.Route), errorOrNil(call.result[1])

}

// List routes
func (r *Recorder) ListRoutes(opts interface{}, routeList *v1alpha1.RouteList, err error) {
	r.add("ListRoutes", apiMethodCall{[]interface{}{opts}, []interface{}{routeList, err}})
}

func (c *MockKnClient) ListRoutes(opts ...ListConfig) (*v1alpha1.RouteList, error) {
	call := c.getCall("ListRoutes")
	c.verifyArgs(call, opts)
	return call.result[0].(*v1alpha1.RouteList), errorOrNil(call.result[1])
}

// GetConfiguration records a call to GetConfiguration with possible return values
func (r *Recorder) GetConfiguration(name string, config *v1alpha1.Configuration, err error) {
	r.add("GetConfiguration", apiMethodCall{[]interface{}{name}, []interface{}{config, err}})

}

// GetConfiguration returns a configuration looked up by name
func (c *MockKnClient) GetConfiguration(name string) (*v1alpha1.Configuration, error) {
	call := c.getCall("GetConfiguration")
	c.verifyArgs(call, name)
	return call.result[0].(*v1alpha1.Configuration), errorOrNil(call.result[1])
}

// Check that every recorded method has been called
func (r *Recorder) Validate() {
	for k, v := range r.recordedCalls {
		if len(v) > 0 {
			r.t.Errorf("Recorded method \"%s\" not been called", k)
		}
	}
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

// Add a recorded api call the list of calls
func (r *Recorder) add(name string, call apiMethodCall) {
	calls, ok := r.recordedCalls[name]
	if !ok {
		calls = make([]apiMethodCall, 0)
		r.recordedCalls[name] = calls
	}
	r.recordedCalls[name] = append(calls, call)
}

// Get the next recorded call
func (r *Recorder) shift(name string) (*apiMethodCall, error) {
	calls := r.recordedCalls[name]
	if len(calls) == 0 {
		return nil, fmt.Errorf("no call to '%s' recorded", name)
	}
	call, calls := calls[0], calls[1:]
	r.recordedCalls[name] = calls
	return &call, nil
}

// Get call and verify that it exist
func (c *MockKnClient) getCall(name string) *apiMethodCall {
	call, err := c.recorder.shift(name)
	assert.NilError(c.t, err, "invalid mock setup, missing recording step")
	return call
}

// Verify given arguments against recorded arguments
func (c *MockKnClient) verifyArgs(call *apiMethodCall, args ...interface{}) {
	callArgs := call.args
	for i, arg := range args {
		assert.Assert(c.t, len(callArgs) > i, "Internal: Invalid recording: Expected %d args, got %d", len(callArgs), len(args))
		fn := reflect.ValueOf(call.args[i])
		fnType := fn.Type()
		if fnType.Kind() == reflect.Func && fnType.NumIn() == 2 {
			fn.Call([]reflect.Value{reflect.ValueOf(c.t), reflect.ValueOf(arg)})
		} else {
			assert.DeepEqual(c.t, call.args[i], arg)
		}
	}
}

func errorOrNil(err interface{}) error {
	if err == nil {
		return nil
	}
	return err.(error)
}
