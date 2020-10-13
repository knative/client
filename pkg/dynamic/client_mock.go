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

package dynamic

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"knative.dev/client/pkg/util/mock"
)

// MockKnDynamicClient is a combine of test object and recorder
type MockKnDynamicClient struct {
	t        *testing.T
	recorder *ClientRecorder
}

// NewMockKnDynamicClient returns a new mock instance which you need to record for
func NewMockKnDynamicClient(t *testing.T, ns ...string) *MockKnDynamicClient {
	namespace := "default"
	if len(ns) > 0 {
		namespace = ns[0]
	}
	return &MockKnDynamicClient{
		t:        t,
		recorder: &ClientRecorder{mock.NewRecorder(t, namespace)},
	}
}

// Ensure that the interface is implemented
var _ KnDynamicClient = &MockKnDynamicClient{}

// ClientRecorder is recorder for eventing objects
type ClientRecorder struct {
	r *mock.Recorder
}

// Recorder returns the recorder for registering API calls
func (c *MockKnDynamicClient) Recorder() *ClientRecorder {
	return c.recorder
}

// Namespace of this client
func (c *MockKnDynamicClient) Namespace() string {
	return c.recorder.r.Namespace()
}

// ListCRDs returns list of installed CRDs in the cluster and filters based on the given options
func (dr *ClientRecorder) ListCRDs(options interface{}, ulist *unstructured.UnstructuredList, err error) {
	dr.r.Add("ListCRDs", []interface{}{options}, []interface{}{ulist, err})
}

// ListCRDs returns list of installed CRDs in the cluster and filters based on the given options
func (c *MockKnDynamicClient) ListCRDs(options metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	call := c.recorder.r.VerifyCall("ListCRDs", options)
	return call.Result[0].(*unstructured.UnstructuredList), mock.ErrorOrNil(call.Result[1])
}

// ListSourcesTypes returns installed knative eventing sources CRDs
func (dr *ClientRecorder) ListSourcesTypes(ulist *unstructured.UnstructuredList, err error) {
	dr.r.Add("ListSourcesTypes", []interface{}{}, []interface{}{ulist, err})
}

// ListSourcesTypes returns installed knative eventing sources CRDs
func (c *MockKnDynamicClient) ListSourcesTypes() (*unstructured.UnstructuredList, error) {
	call := c.recorder.r.VerifyCall("ListSourcesTypes")
	return call.Result[0].(*unstructured.UnstructuredList), mock.ErrorOrNil(call.Result[1])
}

// ListChannelsTypes returns installed knative messaging CRDs
func (dr *ClientRecorder) ListChannelsTypes(ulist *unstructured.UnstructuredList, err error) {
	dr.r.Add("ListChannelsTypes", []interface{}{}, []interface{}{ulist, err})
}

// ListChannelsTypes returns installed knative messaging CRDs
func (c *MockKnDynamicClient) ListChannelsTypes() (*unstructured.UnstructuredList, error) {
	call := c.recorder.r.VerifyCall("ListChannelsTypes")
	return call.Result[0].(*unstructured.UnstructuredList), mock.ErrorOrNil(call.Result[1])
}

// ListSources returns list of available sources objects
func (dr *ClientRecorder) ListSources(types interface{}, ulist *unstructured.UnstructuredList, err error) {
	dr.r.Add("ListSources", []interface{}{types}, []interface{}{ulist, err})
}

// ListSources returns list of available sources objects
func (c *MockKnDynamicClient) ListSources(types ...WithType) (*unstructured.UnstructuredList, error) {
	call := c.recorder.r.VerifyCall("ListSources")
	return call.Result[0].(*unstructured.UnstructuredList), mock.ErrorOrNil(call.Result[1])
}

// RawClient creates a client
func (dr *ClientRecorder) RawClient(dynamicInterface dynamic.Interface) {
	dr.r.Add("RawClient", []interface{}{}, []interface{}{dynamicInterface})
}

// RawClient creates a client
func (c *MockKnDynamicClient) RawClient() (dynamicInterface dynamic.Interface) {
	call := c.recorder.r.VerifyCall("RawClient")
	return call.Result[0].(dynamic.Interface)
}

// ListSourcesUsingGVKs returns list of available source objects using given list of GVKs
func (dr *ClientRecorder) ListSourcesUsingGVKs(gvks interface{}, types interface{}, ulist *unstructured.UnstructuredList, err error) {
	dr.r.Add("ListSourcesUsingGVKs", []interface{}{gvks, types}, []interface{}{ulist, err})
}

// ListSourcesUsingGVKs returns list of available source objects using given list of GVKs
func (c *MockKnDynamicClient) ListSourcesUsingGVKs(gvks *[]schema.GroupVersionKind, types ...WithType) (*unstructured.UnstructuredList, error) {
	call := c.recorder.r.VerifyCall("ListSourcesUsingGVKs")
	return call.Result[0].(*unstructured.UnstructuredList), mock.ErrorOrNil(call.Result[1])
}

// Validate validates whether every recorded action has been called
func (dr *ClientRecorder) Validate() {
	dr.r.CheckThatAllRecordedMethodsHaveBeenCalled()
}

// ListChannelsUsingGVKs returns list of available channel objects using given list of GVKs
func (dr *ClientRecorder) ListChannelsUsingGVKs(gvks interface{}, types interface{}, ulist *unstructured.UnstructuredList, err error) {
	dr.r.Add("ListChannelsUsingGVKs", []interface{}{gvks, types}, []interface{}{ulist, err})
}

// ListChannelsUsingGVKs returns list of available channel objects using given list of GVKs
func (c *MockKnDynamicClient) ListChannelsUsingGVKs(gvks *[]schema.GroupVersionKind, types ...WithType) (*unstructured.UnstructuredList, error) {
	call := c.recorder.r.VerifyCall("ListChannelsUsingGVKs")
	return call.Result[0].(*unstructured.UnstructuredList), mock.ErrorOrNil(call.Result[1])
}
