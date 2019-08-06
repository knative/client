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

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	api_serving "github.com/knative/serving/pkg/apis/serving"

	knclient "github.com/knative/client/pkg/serving/v1alpha1"
)

func TestServiceDescribeBasic(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	expectedService := createnTestService("foo")

	// Get service
	r.GetService("foo", &expectedService, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeWithDetails(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	expectedService := createnTestService("foo")
	r.GetService("foo", &expectedService, nil)

	revList := v1alpha1.RevisionList{}

	// Return the list of all revisions
	r.ListRevisions(HasServiceLabelSelector("foo"), &revList, nil)

	// Fetch the revision
	// r.GetRevision()

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo", "--details")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestEmptyServiceDescribe(t *testing.T) {
	client := knclient.NewMockKnClient(t)
	_, err := executeServiceCommand(client, "service", "describe")
	assert.ErrorContains(t, err, "no", "service", "provided")
}

func validateServiceOutput(t *testing.T, service string, output string) {
	assert.Assert(t, cmp.Regexp("Name:\\s+"+service, output))
	assert.Assert(t, cmp.Regexp("Namespace:\\s+default", output))
	assert.Assert(t, cmp.Regexp("Address:\\s+http://"+service+".default.svc.cluster.local", output))
	assert.Assert(t, cmp.Regexp("URL:\\s+"+service+".default.example.com", output))
	assert.Assert(t, cmp.Regexp("Age:", output))
}

func HasServiceLabelSelector(service string) func(t *testing.T, a interface{}) {
	return func(t *testing.T, a interface{}) {
		lc := a.([]knclient.ListConfig)
		listConfigCollector := knclient.ListConfigCollector{
			Labels: make(labels.Set),
			Fields: make(fields.Set),
		}
		lc[0](&listConfigCollector)
		assert.Equal(t, listConfigCollector.Labels[api_serving.ServiceLabelKey], service)
	}
}

func createnTestService(service string) v1alpha1.Service {
	expectedService := v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service,
			Namespace: "default",
		},
		Status: v1alpha1.ServiceStatus{
			RouteStatusFields: v1alpha1.RouteStatusFields{
				DeprecatedDomain: service + ".default.example.com",
				Address:          &duckv1alpha1.Addressable{Hostname: service + ".default.svc.cluster.local"},
			},
		},
	}
	return expectedService
}
