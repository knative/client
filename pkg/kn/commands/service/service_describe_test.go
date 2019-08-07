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
	api_serving "github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knclient "github.com/knative/client/pkg/serving/v1alpha1"
	"github.com/knative/client/pkg/util"
)

func TestServiceDescribeBasic(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	expectedService := createTestService("foo")

	// Get service
	r.GetService("foo", &expectedService, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeVerbose(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	expectedService := createTestService("foo")
	r.GetService("foo", &expectedService, nil)

	revList := v1alpha1.RevisionList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RevisionList",
			APIVersion: "knative.dev/v1alpha1",
		},
		Items: []v1alpha1.Revision{
			createTestRevision("rev1", 1),
			createTestRevision("rev2", 2),
		},
	}

	// Return the list of all revisions
	r.ListRevisions(knclient.HasLabelSelector(api_serving.ServiceLabelKey, "foo"), &revList, nil)

	// Fetch the revision
	// r.GetRevision()

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo", "--verbose")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeWithWrongArguments(t *testing.T) {
	client := knclient.NewMockKnClient(t)
	_, err := executeServiceCommand(client, "describe")
	assert.ErrorContains(t, err, "no", "service", "provided")

	_, err = executeServiceCommand(client, "describe", "foo", "bar")
	assert.ErrorContains(t, err, "more than one", "service", "provided")
}

func TestServiceDescribeMachineReadable(t *testing.T) {
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	expectedService := createTestService("foo")
	r.GetService("foo", &expectedService, nil)

	output, err := executeServiceCommand(client, "describe", "foo", "-o", "yaml")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(output, "kind: Service", "spec:", "status:", "metadata:"))

	r.Validate()
}

func validateServiceOutput(t *testing.T, service string, output string) {
	assert.Assert(t, cmp.Regexp("Name:\\s+"+service, output))
	assert.Assert(t, cmp.Regexp("Namespace:\\s+default", output))
	assert.Assert(t, cmp.Regexp("Address:\\s+http://"+service+".default.svc.cluster.local", output))
	assert.Assert(t, cmp.Regexp("URL:\\s+"+service+".default.example.com", output))
	assert.Assert(t, cmp.Regexp("Age:", output))
}

func createTestService(service string) v1alpha1.Service {
	return v1alpha1.Service{
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
}

func createTestRevision(revision string, gen int64) v1alpha1.Revision {
	return v1alpha1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       revision,
			Namespace:  "default",
			Generation: gen,
		},
	}

}
