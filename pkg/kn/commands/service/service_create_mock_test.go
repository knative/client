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
	"testing"

	"github.com/knative/pkg/apis"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"

	servinglib "github.com/knative/client/pkg/serving"
	knclient "github.com/knative/client/pkg/serving/v1alpha1"

	"github.com/knative/client/pkg/util"
)

func TestServiceCreateImageMock(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))
	// Create service (don't validate given service --> "Any()" arg is allowed)
	r.CreateService(knclient.Any(), nil)
	// Wait for service to become ready
	r.WaitForService("foo", knclient.Any(), nil)
	// Get for showing the URL
	r.GetService("foo", getServiceWithUrl("foo", "http://foo.example.com"), nil)

	// Testing:
	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "http://foo.example.com", "Waiting"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceCreateLabel(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	r := client.Recorder()
	r.GetService("foo", nil, errors.NewNotFound(v1alpha1.Resource("service"), "foo"))

	service := getService("foo")
	expected := map[string]string{
		"a":     "mouse",
		"b":     "cookie",
		"empty": "",
	}
	service.ObjectMeta.Labels = expected
	template, err := servinglib.RevisionTemplateOfService(service)
	assert.NilError(t, err)
	template.ObjectMeta.Labels = expected
	template.Spec.DeprecatedContainer.Image = "gcr.io/foo/bar:baz"
	r.CreateService(service, nil)

	output, err := executeServiceCommand(client, "create", "foo", "--image", "gcr.io/foo/bar:baz", "-l", "a=mouse", "--label", "b=cookie", "--label=empty", "--async")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "created", "foo", "default"))

	r.Validate()
}

func getService(name string) *v1alpha1.Service {
	service := &v1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{
			DeprecatedRunLatest: &v1alpha1.RunLatestType{
				Configuration: v1alpha1.ConfigurationSpec{
					DeprecatedRevisionTemplate: &v1alpha1.RevisionTemplateSpec{
						Spec: v1alpha1.RevisionSpec{
							DeprecatedContainer: &corev1.Container{
								Resources: corev1.ResourceRequirements{
									Limits:   corev1.ResourceList{},
									Requests: corev1.ResourceList{},
								},
							},
						},
					},
				},
			},
		},
	}
	return service
}

func getServiceWithUrl(name string, urlName string) *v1alpha1.Service {
	service := v1alpha1.Service{}
	url, _ := apis.ParseURL(urlName)
	service.Status.URL = url
	service.Name = name
	return &service
}
