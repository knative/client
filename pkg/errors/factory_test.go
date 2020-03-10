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
	"testing"

	"gotest.tools/assert"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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
			ExpectedMsg: "no Knative serving API found on the backend, please verify the installation",
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
