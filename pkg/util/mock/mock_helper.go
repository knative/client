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

package mock

import (
	"fmt"
	"reflect"
	"testing"

	"gotest.tools/assert"
)

// Recorded method call
type ApiMethodCall struct {
	Args   []interface{}
	Result []interface{}
}

// Recorder for recording mock call
type Recorder struct {

	// Test object used for asserting
	t *testing.T

	// List of recorded calls in order
	recordedCalls map[string][]ApiMethodCall

	// Namespace for client
	namespace string
}

// Add a recorded api call the list of calls
func (r *Recorder) Add(name string, args []interface{}, result []interface{}) {
	call := ApiMethodCall{args, result}
	calls, ok := r.recordedCalls[name]
	if !ok {
		calls = make([]ApiMethodCall, 0)
		r.recordedCalls[name] = calls
	}
	r.recordedCalls[name] = append(calls, call)
}

// Get the next recorded call
func (r *Recorder) Shift(name string) (*ApiMethodCall, error) {
	calls := r.recordedCalls[name]
	if len(calls) == 0 {
		return nil, fmt.Errorf("no call to '%s' recorded", name)
	}
	call, calls := calls[0], calls[1:]
	r.recordedCalls[name] = calls
	return &call, nil
}

func NewRecorder(t *testing.T, namespace string) *Recorder {
	return &Recorder{
		t:             t,
		recordedCalls: make(map[string][]ApiMethodCall),
		namespace:     namespace,
	}
}

func (r *Recorder) Namespace() string {
	return r.namespace
}

// Check if every method has been called
func (r *Recorder) CheckThatAllRecordedMethodsHaveBeenCalled() {
	for k, v := range r.recordedCalls {
		if len(v) > 0 {
			r.t.Errorf("Recorded method \"%s\" not been called", k)
		}
	}
}

// Verify given arguments against recorded arguments
func (r *Recorder) VerifyCall(name string, args ...interface{}) *ApiMethodCall {
	call := r.getCall(name)
	callArgs := call.Args
	for i, arg := range args {
		assert.Assert(r.t, len(callArgs) > i, "Internal: Invalid recording: Expected %d args, got %d", len(callArgs), len(args))
		fn := reflect.ValueOf(callArgs[i])
		fnType := fn.Type()
		if fnType.Kind() == reflect.Func {
			if fnType.NumIn() == 2 &&
				// It's an assertion function which takes a Testing as first parameter
				fnType.In(0).AssignableTo(reflect.TypeOf(r.t)) {
				fn.Call([]reflect.Value{reflect.ValueOf(r.t), reflect.ValueOf(arg)})
			} else {
				assert.Assert(r.t, fnType.AssignableTo(reflect.TypeOf(arg)))
			}
		} else {
			assert.DeepEqual(r.t, callArgs[i], arg)
		}
	}
	return call
}

// Get call and verify that it exist
func (r *Recorder) getCall(name string) *ApiMethodCall {
	call, err := r.Shift(name)
	assert.NilError(r.t, err, "invalid mock setup, missing recording step")
	return call
}

// =====================================================================
// Helper methods

// mock.Any() can be used in recording to not check for the argument
func Any() func(t *testing.T, a interface{}) {
	return func(t *testing.T, a interface{}) {}
}

// Helper method to cast to an error if given
func ErrorOrNil(err interface{}) error {
	if err == nil {
		return nil
	}
	return err.(error)
}
