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

package v1beta1

import (
	"testing"

	"knative.dev/eventing/pkg/apis/messaging/v1beta1"

	"knative.dev/client/pkg/util/mock"
)

type MockKnSubscriptionsClient struct {
	t         *testing.T
	recorder  *SubscriptionsRecorder
	namespace string
}

// NewMockKnSubscriptionsClient returns a new mock instance which you need to record for
func NewMockKnSubscriptionsClient(t *testing.T, ns ...string) *MockKnSubscriptionsClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnSubscriptionsClient{
		t:        t,
		recorder: &SubscriptionsRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnSubscriptionsClient = &MockKnSubscriptionsClient{}

// recorder for service
type SubscriptionsRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnSubscriptionsClient) Recorder() *SubscriptionsRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnSubscriptionsClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// CreateSubscription records a call for CreateSubscription with the expected error
func (sr *SubscriptionsRecorder) CreateSubscription(subscription interface{}, err error) {
	sr.r.Add("CreateSubscription", []interface{}{subscription}, []interface{}{err})
}

// CreateSubscription performs a previously recorded action, failing if non has been registered
func (c *MockKnSubscriptionsClient) CreateSubscription(subscription *v1beta1.Subscription) error {
	call := c.recorder.r.VerifyCall("CreateSubscription", subscription)
	return mock.ErrorOrNil(call.Result[0])
}

// GetSubscription records a call for GetSubscription with the expected object or error. Either subscriptions or err should be nil
func (sr *SubscriptionsRecorder) GetSubscription(name interface{}, subscription *v1beta1.Subscription, err error) {
	sr.r.Add("GetSubscription", []interface{}{name}, []interface{}{subscription, err})
}

// GetSubscription performs a previously recorded action, failing if non has been registered
func (c *MockKnSubscriptionsClient) GetSubscription(name string) (*v1beta1.Subscription, error) {
	call := c.recorder.r.VerifyCall("GetSubscription", name)
	return call.Result[0].(*v1beta1.Subscription), mock.ErrorOrNil(call.Result[1])
}

// DeleteSubscription records a call for DeleteSubscription with the expected error (nil if none)
func (sr *SubscriptionsRecorder) DeleteSubscription(name interface{}, err error) {
	sr.r.Add("DeleteSubscription", []interface{}{name}, []interface{}{err})
}

// DeleteSubscription performs a previously recorded action, failing if non has been registered
func (c *MockKnSubscriptionsClient) DeleteSubscription(name string) error {
	call := c.recorder.r.VerifyCall("DeleteSubscription", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListSubscription records a call for ListSubscription with the expected error (nil if none)
func (sr *SubscriptionsRecorder) ListSubscription(subscriptionsList *v1beta1.SubscriptionList, err error) {
	sr.r.Add("ListSubscription", []interface{}{}, []interface{}{subscriptionsList, err})
}

// ListSubscription performs a previously recorded action, failing if non has been registered
func (c *MockKnSubscriptionsClient) ListSubscription() (*v1beta1.SubscriptionList, error) {
	call := c.recorder.r.VerifyCall("ListSubscription")
	return call.Result[0].(*v1beta1.SubscriptionList), mock.ErrorOrNil(call.Result[1])
}

// UpdateSubscription records a call for CreateSubscription with the expected error
func (sr *SubscriptionsRecorder) UpdateSubscription(subscription interface{}, err error) {
	sr.r.Add("UpdateSubscription", []interface{}{subscription}, []interface{}{err})
}

// UpdateSubscription performs a previously recorded action, failing if non has been registered
func (c *MockKnSubscriptionsClient) UpdateSubscription(subscription *v1beta1.Subscription) error {
	call := c.recorder.r.VerifyCall("UpdateSubscription", subscription)
	return mock.ErrorOrNil(call.Result[0])
}

// Validates validates whether every recorded action has been called
func (sr *SubscriptionsRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}
