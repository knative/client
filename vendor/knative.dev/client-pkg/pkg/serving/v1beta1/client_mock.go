// Copyright © 2021 The Knative Authors
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

	"knative.dev/client-pkg/pkg/util/mock"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
)

// MockKnServingClient client mock
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

// ServingRecorder recorder for service
type ServingRecorder struct {
	r *mock.Recorder
}

// Recorder returns the record instance
func (c *MockKnServingClient) Recorder() *ServingRecorder {
	return c.recorder
}

// Validate checks that every recorded method has been called
func (sr *ServingRecorder) Validate() {
	sr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}

// Namespace of this client
func (c *MockKnServingClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// GetDomainMapping mock function recorder
func (sr *ServingRecorder) GetDomainMapping(name interface{}, domainMapping *servingv1beta1.DomainMapping, err error) {
	sr.r.Add("GetDomainMapping", []interface{}{name}, []interface{}{domainMapping, err})
}

// GetDomainMapping mock function
func (c *MockKnServingClient) GetDomainMapping(ctx context.Context, name string) (*servingv1beta1.DomainMapping, error) {
	call := c.recorder.r.VerifyCall("GetDomainMapping", name)
	return call.Result[0].(*servingv1beta1.DomainMapping), mock.ErrorOrNil(call.Result[1])
}

// CreateDomainMapping recorder function
func (sr *ServingRecorder) CreateDomainMapping(domainMapping interface{}, err error) {
	sr.r.Add("CreateDomainMapping", []interface{}{domainMapping}, []interface{}{err})
}

// CreateDomainMapping mock function
func (c *MockKnServingClient) CreateDomainMapping(ctx context.Context, domainMapping *servingv1beta1.DomainMapping) error {
	call := c.recorder.r.VerifyCall("CreateDomainMapping", domainMapping)
	return mock.ErrorOrNil(call.Result[0])
}

// UpdateDomainMapping recorder function
func (sr *ServingRecorder) UpdateDomainMapping(domainMapping interface{}, err error) {
	sr.r.Add("UpdateDomainMapping", []interface{}{domainMapping}, []interface{}{err})
}

// UpdateDomainMapping mock function
func (c *MockKnServingClient) UpdateDomainMapping(ctx context.Context, domainMapping *servingv1beta1.DomainMapping) error {
	call := c.recorder.r.VerifyCall("UpdateDomainMapping", domainMapping)
	return mock.ErrorOrNil(call.Result[0])
}

func (cl *MockKnServingClient) UpdateDomainMappingWithRetry(ctx context.Context, name string, updateFunc DomainUpdateFunc, nrRetries int) error {
	return updateDomainMappingWithRetry(ctx, cl, name, updateFunc, nrRetries)
}

// DeleteDomainMapping recorder function
func (sr *ServingRecorder) DeleteDomainMapping(name string, err error) {
	sr.r.Add("DeleteDomainMapping", []interface{}{name}, []interface{}{err})
}

// DeleteDomainMapping mock function
func (c *MockKnServingClient) DeleteDomainMapping(ctx context.Context, name string) error {
	call := c.recorder.r.VerifyCall("DeleteDomainMapping", name)
	return mock.ErrorOrNil(call.Result[0])
}

// ListDomainMappings recorder function
func (sr *ServingRecorder) ListDomainMappings(domainMappingList *servingv1beta1.DomainMappingList, err error) {
	sr.r.Add("ListDomainMappings", nil, []interface{}{domainMappingList, err})
}

// ListDomainMappings mock function
func (c *MockKnServingClient) ListDomainMappings(ctx context.Context) (*servingv1beta1.DomainMappingList, error) {
	call := c.recorder.r.VerifyCall("ListDomainMappings")
	return call.Result[0].(*servingv1beta1.DomainMappingList), mock.ErrorOrNil(call.Result[1])
}
