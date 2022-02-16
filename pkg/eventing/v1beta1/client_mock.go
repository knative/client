// Copyright © 2022 The Knative Authors
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
	"context"
	"testing"

	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"

	"knative.dev/client/pkg/util/mock"
)

// MockKnEventingV1beta1Client is a combine of test object and recorder
type MockKnEventingV1beta1Client struct {
	t        *testing.T
	recorder *EventingV1beta1Recorder
}

// NewMockKnEventingV1beta1Client returns a new mock instance which you need to record for
func NewMockKnEventingV1beta1Client(t *testing.T, ns ...string) *MockKnEventingV1beta1Client {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnEventingV1beta1Client{
		t:        t,
		recorder: &EventingV1beta1Recorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnEventingV1Beta1Client = &MockKnEventingV1beta1Client{}

// EventingV1beta1Recorder is recorder for eventingv1beta1 objects
type EventingV1beta1Recorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnEventingV1beta1Client) Recorder() *EventingV1beta1Recorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnEventingV1beta1Client) Namespace() string {
	return c.recorder.r.Namespace()
}

// ListEventtypes records a call for ListEventtypes with the expected result and error (nil if none)
func (sr *EventingV1beta1Recorder) ListEventtypes(eventtypeList *eventingv1beta1.EventTypeList, err error) {
	sr.r.Add("ListEventtypes", nil, []interface{}{eventtypeList, err})
}

func (c *MockKnEventingV1beta1Client) ListEventtypes(ctx context.Context) (*eventingv1beta1.EventTypeList, error) {
	call := c.recorder.r.VerifyCall("ListEventtypes")
	return call.Result[0].(*eventingv1beta1.EventTypeList), mock.ErrorOrNil(call.Result[1])
}

// GetEventtypes records a call for GetEventtype with the expected object or error. Either eventtype or err should be nil
func (c *MockKnEventingV1beta1Client) GetEventtype(ctx context.Context, name string) (*eventingv1beta1.EventType, error) {
	call := c.recorder.r.VerifyCall("GetEventtype", name)
	return call.Result[0].(*eventingv1beta1.EventType), mock.ErrorOrNil(call.Result[1])
}

// CreateEventtype records a call for CreateEventtype with the expected error
func (sr *EventingV1beta1Recorder) CreateEventtype(eventtype interface{}, err error) {
	sr.r.Add("CreateEventtype", []interface{}{eventtype}, []interface{}{err})
}

func (c *MockKnEventingV1beta1Client) CreateEventtype(ctx context.Context, eventtype *eventingv1beta1.EventType) error {
	call := c.recorder.r.VerifyCall("CreateEventtype", eventtype)
	return mock.ErrorOrNil(call.Result[0])
}

// DeleteEventtype records a call for DeleteEventtype with the expected error
func (sr *EventingV1beta1Recorder) DeleteEventtype(name interface{}, err error) {
	sr.r.Add("DeleteEventtype", []interface{}{name}, []interface{}{err})
}

func (c *MockKnEventingV1beta1Client) DeleteEventtype(ctx context.Context, name string) error {
	call := c.recorder.r.VerifyCall("DeleteEventtype", name)
	return mock.ErrorOrNil(call.Result[0])
}

// Validate validates whether every recorded action has been called
func (sr *EventingV1beta1Recorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
