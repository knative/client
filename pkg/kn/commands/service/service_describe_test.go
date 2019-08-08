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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	"github.com/knative/serving/pkg/apis/autoscaling"
	api_serving "github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knclient "github.com/knative/client/pkg/serving/v1alpha1"
	"github.com/knative/client/pkg/util"
)

const (
	imageDigest = "sha256:1234567890123456789012345678901234567890123456789012345678901234"
)

func TestServiceDescribeBasic(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()
	// Prepare service
	expectedService := createTestService("foo", []string{"rev1"}, goodConditions())

	// Get service & revision
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	r.GetRevision("rev1", &rev1, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "Env:", "label1=lval1, label2=lval2\n"))
	assert.Assert(t, util.ContainsAll(output, "Annotations:", "anno1=aval1, anno2=aval2, anno3="))
	assert.Assert(t, cmp.Regexp(`(?m)\s*Annotations:.*\.\.\.$`, output))
	assert.Assert(t, util.ContainsAll(output, "Labels:", "label1=lval1, label2=lval2\n"))
	assert.Assert(t, util.ContainsAll(output, "[1]"))
	// no digest added (added only for details)
	assert.Assert(t, !strings.Contains(output, "(123456789012)"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeScaling(t *testing.T) {

	for _, data := range []struct {
		minScale, maxScale, limit, target string
		scaleOut                          string
	}{
		{"", "", "", "", ""},
		{"", "10", "", "", "0 ... 10"},
		{"10", "", "", "", "10 ... ∞"},
		{"5", "20", "10", "", "5 ... 20"},
		{"", "", "20", "30", ""},
	} {
		// New mock client
		client := knclient.NewMockKnClient(t)

		// Recording:
		r := client.Recorder()

		// Prepare service
		expectedService := createTestService("foo", []string{"rev1"}, goodConditions())

		// Get service & revision
		r.GetService("foo", &expectedService, nil)
		rev1 := createTestRevision("rev1", 1)
		addScaling(&rev1, data.minScale, data.maxScale, data.target, data.limit)
		r.GetRevision("rev1", &rev1, nil)

		revList := v1alpha1.RevisionList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RevisionList",
				APIVersion: "knative.dev/v1alpha1",
			},
			Items: []v1alpha1.Revision{
				rev1,
			},
		}

		// Return the list of all revisions
		r.ListRevisions(knclient.HasLabelSelector(api_serving.ServiceLabelKey, "foo"), &revList, nil)

		// Testing:
		output, err := executeServiceCommand(client, "describe", "foo", "--verbose")
		assert.NilError(t, err)

		validateServiceOutput(t, "foo", output)

		if data.limit != "" || data.target != "" {
			assert.Assert(t, util.ContainsAll(output, "Concurrency:"))
		} else {
			assert.Assert(t, !strings.Contains(output, "Concurrency:"))
		}

		validateOutputLine(t, output, "Scale", data.scaleOut)
		validateOutputLine(t, output, "Limit", data.limit)
		validateOutputLine(t, output, "Target", data.target)

		// Validate that all recorded API methods have been called
		r.Validate()
	}
}

func validateOutputLine(t *testing.T, output string, label string, value string) {
	if value != "" {
		assert.Assert(t, cmp.Regexp(fmt.Sprintf("%s:\\s*%s", label, value), output))
	} else {
		assert.Assert(t, !strings.Contains(output, label+":"))
	}
}

func TestServiceDescribeResources(t *testing.T) {

	for _, data := range []struct {
		reqMem, limitMem, reqCPU, limitCPU string
		memoryOut, cpuOut                  string
	}{
		{"", "", "", "", "", ""},
		{"10Mi", "100Mi", "100m", "1", "10Mi ... 100Mi", "100m ... 1"},
		{"", "100Mi", "", "1", "100Mi", "1"},
		{"10Mi", "", "100m", "", "10Mi", "100m"},
	} {
		// New mock client
		client := knclient.NewMockKnClient(t)

		// Recording:
		r := client.Recorder()

		// Prepare service
		expectedService := createTestService("foo", []string{"rev1"}, goodConditions())

		// Get service & revision
		r.GetService("foo", &expectedService, nil)
		rev1 := createTestRevision("rev1", 1)
		addResourceLimits(&rev1.Spec.Containers[0].Resources, data.reqMem, data.limitMem, data.reqCPU, data.limitCPU)
		r.GetRevision("rev1", &rev1, nil)

		revList := v1alpha1.RevisionList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RevisionList",
				APIVersion: "knative.dev/v1alpha1",
			},
			Items: []v1alpha1.Revision{
				rev1,
			},
		}

		// Return the list of all revisions
		r.ListRevisions(knclient.HasLabelSelector(api_serving.ServiceLabelKey, "foo"), &revList, nil)

		// Testing:
		output, err := executeServiceCommand(client, "describe", "foo", "--verbose")
		assert.NilError(t, err)

		validateServiceOutput(t, "foo", output)

		validateOutputLine(t, output, "Memory", data.memoryOut)
		validateOutputLine(t, output, "CPU", data.cpuOut)

		// Validate that all recorded API methods have been called
		r.Validate()
	}
}

func TestServiceDescribeVerbose(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnClient(t)

	// Recording:
	r := client.Recorder()

	// Prepare service
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

	assert.Assert(t, util.ContainsAll(output, "Image", "Name", "(123456789012)", "50%", "gcr.io/test/image", "(0s)"))
	assert.Assert(t, util.ContainsAll(output, "Env:", "label1=lval1\n", "label2=lval2\n"))
	assert.Assert(t, util.ContainsAll(output, "Annotations:", "anno1=aval1\n", "anno2=aval2\n"))
	assert.Assert(t, util.ContainsAll(output, "Labels:", "label1=lval1\n", "label2=lval2\n"))
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

	// Prepare service
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

	assert.Assert(t, util.ContainsAll(output, "Age:", "Revisions:", "Conditions:", "Labels:", "Annotations:", "Port:", "8080"))
	assert.Assert(t, util.ContainsAll(output, "Ready", "RoutesReady", "OK", "TYPE", "AGE", "REASON"))
}

func createTestService(name string, revisionNames []string, conditions duckv1beta1.Conditions) v1alpha1.Service {

	labelMap := make(map[string]string)
	labelMap["label1"] = "lval1"
	labelMap["label2"] = "lval2"
	annoMap := make(map[string]string)
	annoMap["anno1"] = "aval1"
	annoMap["anno2"] = "aval2"
	annoMap["anno3"] = "very_long_value_which_should_be_truncated_in_normal_output_if_we_make_it_even_longer"

	service := v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			Labels:            labelMap,
			Annotations:       annoMap,
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-30 * time.Second)},
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

func addScaling(revision *v1alpha1.Revision, minScale, maxScale, concurrencyTarget, concurrenyLimit string) {
	annos := make(map[string]string)
	if minScale != "" {
		annos[autoscaling.MinScaleAnnotationKey] = minScale
	}
	if maxScale != "" {
		annos[autoscaling.MaxScaleAnnotationKey] = maxScale
	}
	if concurrencyTarget != "" {
		annos[autoscaling.TargetAnnotationKey] = concurrencyTarget
	}
	revision.Annotations = annos
	if concurrenyLimit != "" {
		l, _ := strconv.Atoi(concurrenyLimit)
		revision.Spec.ContainerConcurrency = v1beta1.RevisionContainerConcurrencyType(l)
	}
}

func addResourceLimits(resources *v1.ResourceRequirements, reqMem, limitMem, reqCPU, limitCPU string) {
	(*resources).Requests = getResourceListQuantity(reqMem, reqCPU)
	(*resources).Limits = getResourceListQuantity(limitMem, limitCPU)
}

func getResourceListQuantity(mem string, cpu string) v1.ResourceList {
	list := v1.ResourceList{}
	if mem != "" {
		q, _ := resource.ParseQuantity(mem)
		list[v1.ResourceMemory] = q
	}
	if cpu != "" {
		q, _ := resource.ParseQuantity(cpu)
		list[v1.ResourceCPU] = q
	}
	return list
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
			Name:              revision,
			Namespace:         "default",
			Generation:        1,
			Labels:            labels,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: v1alpha1.RevisionSpec{
			RevisionSpec: v1beta1.RevisionSpec{
				PodSpec: v1beta1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "gcr.io/test/image",
							Env: []v1.EnvVar{
								{Name: "env1", Value: "eval1"},
								{Name: "env2", Value: "eval2"},
							},
							Ports: []v1.ContainerPort{
								{ContainerPort: 8080},
							},
						},
					},
				},
			},
		},
		Status: v1alpha1.RevisionStatus{
			ImageDigest: imageDigest,
		},
	}
}

func goodConditions() duckv1beta1.Conditions {
	ret := make(duckv1beta1.Conditions, 0)
	ret = append(ret, apis.Condition{
		Type:   apis.ConditionReady,
		Status: v1.ConditionTrue,
		LastTransitionTime: apis.VolatileTime{
			Inner: metav1.Time{Time: time.Now()},
		},
	})
	ret = append(ret, apis.Condition{
		Type:   v1alpha1.ServiceConditionRoutesReady,
		Status: v1.ConditionTrue,
		LastTransitionTime: apis.VolatileTime{
			Inner: metav1.Time{Time: time.Now()},
		},
	})
	return ret
}
