// Copyright © 2020 The Knative Authors
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

	"knative.dev/eventing/pkg/apis/messaging/v1beta1"

	"knative.dev/client/pkg/util/mock"
)

type MockKnChannelsClient struct {
	t         *testing.T
	recorder  *ChannelsRecorder
	namespace string
}

// NewMockKnChannelsClient returns a new mock instance which you need to record for
func NewMockKnChannelsClient(t *testing.T, ns ...string) *MockKnChannelsClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnChannelsClient{
		t:        t,
		recorder: &ChannelsRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnChannelsClient = &MockKnChannelsClient{}

// recorder for service
type ChannelsRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnChannelsClient) Recorder() *ChannelsRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnChannelsClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// CreateChannel records a call for CreateChannel with the expected error
func (sr *ChannelsRecorder) CreateChannel(channels interface{}, err error) {
	sr.r.Add("CreateChannel", []interface{}{channels}, []interface{}{err})
}

// CreateChannel performs a previously recorded action, failing if non has been registered
func (c *MockKnChannelsClient) CreateChannel(channels *v1beta1.Channel) error {
	call := c.recorder.r.VerifyCall("CreateChannel", channels)
	return mock.ErrorOrNil(call.Result[0])
}

// GetChannel records a call for GetChannel with the expected object or error. Either channels or err should be nil
func (sr *ChannelsRecorder) GetChannel(name interface{}, channels *v1beta1.Channel, err error) {
	sr.r.Add("GetChannel", []interface{}{name}, []interface{}{channels, err})
}

// GetChannel performs a previously recorded action, failing if non has been registered
func (c *MockKnChannelsClient) GetChannel(name string) (*v1beta1.Channel, error) {
	call := c.recorder.r.VerifyCall("GetChannel", name)
	return call.Result[0].(*v1beta1.Channel), mock.ErrorOrNil(call.Result[1])
}

// DeleteChannel records a call for DeleteChannel with the expected error (nil if none)
func (sr *ChannelsRecorder) DeleteChannel(name interface{}, err error) {
	sr.r.Add("DeleteChannel", []interface{}{name}, []interface{}{err})
}

// DeleteChannel performs a previously recorded action, failing if non has been registered
func (c *MockKnChannelsClient) DeleteChannel(name string) error {
	call := c.recorder.r.VerifyCall("DeleteChannel", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListChannel records a call for ListChannel with the expected error (nil if none)
func (sr *ChannelsRecorder) ListChannel(channelsList *v1beta1.ChannelList, err error) {
	sr.r.Add("ListChannel", []interface{}{}, []interface{}{channelsList, err})
}

// ListChannel performs a previously recorded action, failing if non has been registered
func (c *MockKnChannelsClient) ListChannel() (*v1beta1.ChannelList, error) {
	call := c.recorder.r.VerifyCall("ListChannel")
	return call.Result[0].(*v1beta1.ChannelList), mock.ErrorOrNil(call.Result[1])
}

// Validates validates whether every recorded action has been called
func (sr *ChannelsRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
