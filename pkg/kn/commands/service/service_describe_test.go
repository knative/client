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

	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	api_serving "github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	v1 "k8s.io/api/core/v1"
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
	expectedService := createTestService("foo", nil, goodConditions())

	// Get service
	r.GetService("foo", &expectedService, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func goodConditions() duckv1beta1.Conditions {
	ret := make(duckv1beta1.Conditions, 0)
	ret = append(ret, apis.Condition{
		Type:   apis.ConditionReady,
		Status: v1.ConditionTrue,
	})
	ret = append(ret, apis.Condition{
		Type:   apis.ConditionSucceeded,
		Status: v1.ConditionTrue,
	})
	return ret
}

func TestServiceDescribeVerbose(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Check for existing service --> no
	expectedService := createTestService("foo", []string{"rev1", "rev2"}, goodConditions())
	r.GetService("foo", &expectedService, nil)

	rev1 := createTestRevision("rev1", 1)
	rev2 := createTestRevision("rev2", 2)

	revList := v1alpha1.RevisionList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RevisionList",
			APIVersion: "knative.dev/v1alpha1",
		},
		Items: []v1alpha1.Revision{
			rev1, rev2,
		},
	}

	// Return the list of all revisions
	r.ListRevisions(knclient.HasLabelSelector(api_serving.ServiceLabelKey, "foo"), &revList, nil)

	// Fetch the revisions
	r.GetRevision("rev1", &rev1, nil)
	r.GetRevision("rev2", &rev2, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo", "--verbose")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "[1]", "[2]"))
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
	expectedService := createTestService("foo", []string{"rev1", "rev2"}, goodConditions())
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

func createTestService(name string, revisionNames []string, conditions duckv1beta1.Conditions) v1alpha1.Service {

	labelMap := make(map[string]string)
	labelMap["label1"] = "lval1"
	labelMap["label2"] = "lval2"
	annoMap := make(map[string]string)
	annoMap["anno1"] = "aval1"
	annoMap["anno2"] = "aval2"

	service := v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Labels:      labelMap,
			Annotations: annoMap,
		},
		Status: v1alpha1.ServiceStatus{
			RouteStatusFields: v1alpha1.RouteStatusFields{
				DeprecatedDomain: name + ".default.example.com",
				Address:          &duckv1alpha1.Addressable{Hostname: name + ".default.svc.cluster.local"},
			},
			Status: duckv1beta1.Status{
				Conditions: conditions,
			},
		},
	}
	if len(revisionNames) > 0 {
		trafficTargets := make([]v1alpha1.TrafficTarget, 0)
		for _, rname := range revisionNames {
			url, _ := apis.ParseURL(fmt.Sprintf("https://%s", rname))
			target := v1alpha1.TrafficTarget{
				TrafficTarget: v1beta1.TrafficTarget{
					RevisionName:      rname,
					ConfigurationName: name,
					Percent:           100 / len(revisionNames),
					URL:               url,
				},
			}
			trafficTargets = append(trafficTargets, target)
		}
		service.Status.Traffic = trafficTargets
	}
	return service
}

func createTestRevision(revision string, gen int64) v1alpha1.Revision {
	labels := make(map[string]string)
	labels[api_serving.ConfigurationGenerationLabelKey] = fmt.Sprintf("%d", gen)
	return v1alpha1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       revision,
			Namespace:  "default",
			Generation: 1,
			Labels:     labels,
		},
		Spec: v1alpha1.RevisionSpec{
			RevisionSpec: v1beta1.RevisionSpec{
				PodSpec: v1beta1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "gcr.io/test/image",
						},
					},
				},
			},
		},
	}
}
