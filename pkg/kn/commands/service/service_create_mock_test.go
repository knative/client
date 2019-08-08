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
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"

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

func getServiceWithUrl(name string, urlName string) *v1alpha1.Service {
	service := v1alpha1.Service{}
	url, _ := apis.ParseURL(urlName)
	service.Status.URL = url
	service.Name = name
	return &service
}
