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

package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	knclient "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util/mock"
	"knative.dev/client/pkg/wait"

	"knative.dev/client/pkg/util"
)

func TestServiceApplyCreateMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	//service := createServiceWithImage("foo", "gcr.io/foo/bar:baz")

	r := setupServiceApplyRecorder(client, "foo", nil, apierrors.NewNotFound(servingv1.Resource("service"), "foo"), true)

	// Testing:
	output, err := executeServiceCommand(client, "apply", "foo", "--image", "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "http://foo.example.com", "Ready"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceApplyUpdateMock(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	service := createServiceWithImage("foo", "gcr.io/foo/bar:baz")

	r := setupServiceApplyRecorder(client, "foo", service, nil, true)

	// Testing:
	output, err := executeServiceCommand(client, "apply", "foo", "--image", "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "applied", "foo", "http://foo.example.com", "Ready"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceApplyUpdateUnchanged(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	service := createServiceWithImage("foo", "gcr.io/foo/bar:baz")

	r := setupServiceApplyRecorder(client, "foo", service, nil, false)

	// Testing:
	output, err := executeServiceCommand(client, "apply", "foo", "--image", "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "No changes", "apply", "foo", "http://foo.example.com"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceApplyWithGetError(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	errThrown := errors.New("boom!")
	r := setupServiceApplyRecorder(client, "foo", nil, errThrown, true)

	_, err := executeServiceCommand(client, "apply", "foo", "--image", "gcr.io/foo/bar:baz")
	assert.Equal(t, err, errThrown)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func setupServiceApplyRecorder(client *knclient.MockKnServingClient, name string, service *servingv1.Service, err error, hasChanged bool) *knclient.ServingRecorder {
	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	r.GetService(name, service, err)
	// Error test
	if err != nil && !apierrors.IsNotFound(err) {
		return r
	}

	// Create service (don't validate given service --> "Any()" arg is allowed)
	r.ApplyService(func(t *testing.T, a interface{}) {
		svc := a.(*servingv1.Service)
		assert.Equal(t, svc.Name, name)
		setUrl(svc, fmt.Sprintf("http://%s.example.com", name))
	}, hasChanged, nil)

	// Fetch service for URL
	r.GetService(name, getServiceWithUrl(name, fmt.Sprintf("http://%s.example.com", name)), nil)

	if !hasChanged {
		return r
	}
	// Wait for service to become ready
	r.WaitForService(name, mock.Any(), wait.NoopMessageCallback(), nil, time.Second)

	return r
}

func createServiceWithImage(name string, image string) *servingv1.Service {
	return &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: image,
								},
							},
						},
					},
				},
			},
		},
	}
}
