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

package errors

import (
	"errors"
	"fmt"
	"testing"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"gotest.tools/v3/assert"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockErrType struct{}

func (err mockErrType) Error() string {
	return "mock error message"
}
func (err mockErrType) Status() metav1.Status {
	return metav1.Status{}
}

func TestKnErrorsStatusErrors(t *testing.T) {
	cases := []struct {
		Name        string
		Schema      schema.GroupResource
		StatusError func(schema.GroupResource) *api_errors.StatusError
		ExpectedMsg string
		Validate    func(t *testing.T, err error, msg string)
	}{
		{
			Name: "Should get a missing serving api error",
			Schema: schema.GroupResource{
				Group:    "serving.knative.dev",
				Resource: "service",
			},
			StatusError: func(resource schema.GroupResource) *api_errors.StatusError {
				statusError := api_errors.NewNotFound(resource, "serv")
				statusError.Status().Details.Causes = []v1.StatusCause{
					{
						Type:    "UnexpectedServerResponse",
						Message: "404 page not found",
					},
				}
				return statusError
			},
			ExpectedMsg: "no or newer Knative Serving API found on the backend, please verify the installation or update the 'kn' client",
			Validate: func(t *testing.T, err error, msg string) {
				assert.Error(t, err, msg)
			},
		},
		{
			Name: "Should get the default not found error",
			Schema: schema.GroupResource{
				Group:    "serving.knative.dev",
				Resource: "service",
			},
			StatusError: func(resource schema.GroupResource) *api_errors.StatusError {
				return api_errors.NewNotFound(resource, "serv")
			},
			ExpectedMsg: "service.serving.knative.dev \"serv\" not found",
			Validate: func(t *testing.T, err error, msg string) {
				assert.Error(t, err, msg)
			},
		},
		{
			Name: "Should return the original error",
			Schema: schema.GroupResource{
				Group:    "serving.knative.dev",
				Resource: "service",
			},
			StatusError: func(resource schema.GroupResource) *api_errors.StatusError {
				return api_errors.NewAlreadyExists(resource, "serv")
			},
			ExpectedMsg: "service.serving.knative.dev \"serv\" already exists",
			Validate: func(t *testing.T, err error, msg string) {
				assert.Error(t, err, msg)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			statusError := tc.StatusError(tc.Schema)
			err := GetError(statusError)
			tc.Validate(t, err, tc.ExpectedMsg)
		})
	}
}

func TestKnErrors(t *testing.T) {
	cases := []struct {
		Name        string
		Error       error
		ExpectedMsg string
	}{
		{
			Name:        "no kubeconfig provided",
			Error:       errors.New("invalid configuration: no configuration has been provided"),
			ExpectedMsg: "no kubeconfig has been provided, please use a valid configuration to connect to the cluster",
		},
		{
			Name:        "i/o timeout",
			Error:       errors.New("Get https://api.example.com:27435/apis/foo/bar: dial tcp 192.168.1.1:27435: i/o timeout"),
			ExpectedMsg: "error connecting to the cluster, please verify connection at: 192.168.1.1:27435: i/o timeout",
		},
		{
			Name:        "no route to host",
			Error:       errors.New("Get https://192.168.39.141:8443/apis/foo/bar: dial tcp 192.168.39.141:8443: connect: no route to host"),
			ExpectedMsg: "error connecting to the cluster, please verify connection at: 192.168.39.141:8443: connect: no route to host",
		},
		{
			Name:        "no route to host without dial tcp string",
			Error:       errors.New("no route to host 192.168.1.1"),
			ExpectedMsg: "error connecting to the cluster: no route to host 192.168.1.1",
		},
		{
			Name:        "foo error which cant be converted to APIStatus",
			Error:       errors.New("foo error"),
			ExpectedMsg: "foo error",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			err := GetError(tc.Error)
			assert.Error(t, err, tc.ExpectedMsg)
		})
	}
}

func TestIsForbiddenError(t *testing.T) {
	cases := []struct {
		Name      string
		Error     error
		Forbidden bool
	}{
		{
			Name:      "forbidden error",
			Error:     api_errors.NewForbidden(schema.GroupResource{Group: "apiextensions.k8s.io", Resource: "CustomResourceDefinition"}, "", nil),
			Forbidden: true,
		},
		{
			Name:      "non forbidden error",
			Error:     errors.New("panic"),
			Forbidden: false,
		},
	}
	for _, tc := range cases {
		err := tc.Error
		forbidden := tc.Forbidden
		t.Run(tc.Name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			assert.Equal(t, IsForbiddenError(GetError(err)), forbidden)
		})
	}
}

func TestNilError(t *testing.T) {
	assert.NilError(t, GetError(nil), nil)
}

func TestIsInternalError(t *testing.T) {
	cases := []struct {
		Name     string
		Error    error
		Internal bool
	}{
		{
			Name:     "internal error with connection refused",
			Error:    api_errors.NewInternalError(fmt.Errorf("failed calling webhook \"webhook.serving.knative.dev\": Post \"https://webhook.knative-serving.svc:443/defaulting?timeout=10s\": dial tcp 10.96.27.233:443: connect: connection refused")),
			Internal: true,
		},
		{
			Name:     "internal error with context deadline exceeded",
			Error:    api_errors.NewInternalError(fmt.Errorf("failed calling webhook \"webhook.serving.knative.dev\": Post https://webhook.knative-serving.svc:443/defaulting?timeout=10s: context deadline exceeded")),
			Internal: true,
		},
		{
			Name:     "internal error with i/o timeout",
			Error:    api_errors.NewInternalError(fmt.Errorf("failed calling webhook \"webhook.serving.knative.dev\": Post https://webhook.knative-serving.svc:443/defaulting?timeout=10s: i/o timeout")),
			Internal: true,
		},
		{
			Name:     "not internal error",
			Error:    mockErrType{},
			Internal: false,
		},
	}

	for _, tc := range cases {
		err := tc.Error
		internal := tc.Internal
		t.Run(tc.Name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			assert.Equal(t, api_errors.IsInternalError(GetError(err)), internal)
		})
	}
}

func TestStatusError(t *testing.T) {
	cases := []struct {
		Name      string
		Error     error
		ErrorType func(error) bool
	}{
		{
			Name:      "Timeout error",
			Error:     api_errors.NewTimeoutError("failed processing request: i/o timeout", 10),
			ErrorType: api_errors.IsTimeout,
		},
		{
			Name:      "Conflict error",
			Error:     api_errors.NewConflict(servingv1.Resource("service"), "tempService", fmt.Errorf("failure: i/o timeout")),
			ErrorType: api_errors.IsConflict,
		},
		{
			Name:  "i/o timeout",
			Error: errors.New("Get https://api.example.com:27435/apis/foo/bar: dial tcp 192.168.1.1:27435: i/o timeout"),
			ErrorType: func(err error) bool {
				var kne *KNError
				return errors.As(err, &kne)
			},
		},
	}
	for _, tc := range cases {
		itc := tc
		t.Run(tc.Name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			assert.Assert(t, itc.ErrorType(GetError(itc.Error)))
		})
	}
}
