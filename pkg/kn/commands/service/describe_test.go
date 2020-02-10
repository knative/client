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

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/serving/pkg/apis/autoscaling"
	api_serving "knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	client_serving "knative.dev/client/pkg/serving"
	knclient "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/pkg/ptr"
)

const (
	imageDigest = "sha256:1234567890123456789012345678901234567890123456789012345678901234"
)

func TestServiceDescribeBasic(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	// Prepare service
	expectedService := createTestServiceWithServiceAccount("foo", []string{"rev1"}, "default-sa", goodConditions())

	// Get service & revision
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	r.GetRevision("rev1", &rev1, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "123456"))
	assert.Assert(t, util.ContainsAll(output, "Annotations:", "anno1=aval1, anno2=aval2, anno3="))
	assert.Assert(t, cmp.Regexp(`(?m)\s*Annotations:.*\.\.\.$`, output))
	assert.Assert(t, util.ContainsAll(output, "Labels:", "label1=lval1, label2=lval2\n"))
	assert.Assert(t, util.ContainsAll(output, "[1]"))
	assert.Assert(t, cmp.Regexp("Service Account: \\s+default-sa", output))

	assert.Equal(t, strings.Count(output, "rev1"), 1)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeSad(t *testing.T) {
	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	expectedService := createTestService("foo", []string{"rev1"}, goodConditions())
	expectedService.Status.Conditions[0].Status = v1.ConditionFalse
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	r.GetRevision("rev1", &rev1, nil)

	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)
	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "!!", "Ready"))

	r.Validate()
}

func TestServiceDescribeLatest(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	expectedService := createTestService("foo", []string{"rev1"}, goodConditions())
	expectedService.Status.Traffic[0].LatestRevision = ptr.Bool(true)

	// Get service & revision
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	r.GetRevision("rev1", &rev1, nil)

	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)
	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "@latest (rev1)"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeLatestNotInTraffic(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	// Prepare service
	expectedService := createTestService("foo", []string{"rev1", "rev2"}, goodConditions())
	expectedService.Status.Traffic = expectedService.Status.Traffic[:1]
	expectedService.Status.Traffic[0].LatestRevision = ptr.Bool(false)
	expectedService.Status.Traffic[0].Percent = ptr.Int64(int64(100))

	// Get service & revision
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	rev2 := createTestRevision("rev2", 2)
	r.GetRevision("rev1", &rev1, nil)
	r.GetRevision("rev2", &rev2, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "rev2 (current @latest)"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeEachNamedOnce(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	// Prepare service
	expectedService := createTestService("foo", []string{"rev1", "rev2"}, goodConditions())
	expectedService.Status.Traffic = expectedService.Status.Traffic[:1]
	expectedService.Status.Traffic[0].LatestRevision = ptr.Bool(false)
	expectedService.Status.Traffic[0].Percent = ptr.Int64(int64(100))

	// Get service & revision
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	rev2 := createTestRevision("rev2", 2)
	r.GetRevision("rev1", &rev1, nil)
	r.GetRevision("rev2", &rev2, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "rev1", "rev2"))
	assert.Equal(t, strings.Count(output, "rev2"), 1)
	assert.Equal(t, strings.Count(output, "rev1"), 1)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeLatestAndCurrentBothHaveTrafficEntries(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	// Prepare service
	expectedService := createTestService("foo", []string{"rev1", "rev1"}, goodConditions())
	expectedService.Status.Traffic[0].LatestRevision = ptr.Bool(true)
	expectedService.Status.Traffic[0].Tag = "latest"
	expectedService.Status.Traffic[1].Tag = "current"

	// Get service & revision
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	r.GetRevision("rev1", &rev1, nil)
	r.GetRevision("rev1", &rev1, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "@latest (rev1) #latest", "rev1 (current @latest) #current", "50%"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeLatestCreatedIsBroken(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()
	// Prepare service
	expectedService := createTestService("foo", []string{"rev1"}, goodConditions())
	expectedService.Status.Traffic[0].LatestRevision = ptr.Bool(true)
	expectedService.Status.LatestCreatedRevisionName = "rev2"

	// Get service & revision
	r.GetService("foo", &expectedService, nil)
	rev1 := createTestRevision("rev1", 1)
	rev2 := createTestRevision("rev2", 2)
	rev2.Status.Conditions[0].Status = v1.ConditionFalse
	r.GetRevision("rev1", &rev1, nil)
	r.GetRevision("rev2", &rev2, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)
	assert.Assert(t, util.ContainsAll(output, "!", "rev2", "100%", "@latest (rev1)"))

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
		client := knclient.NewMockKnServiceClient(t)

		// Recording:
		r := client.Recorder()

		// Prepare service
		expectedService := createTestService("foo", []string{"rev1"}, goodConditions())

		// Get service & revision
		r.GetService("foo", &expectedService, nil)
		rev1 := createTestRevision("rev1", 1)
		addScaling(&rev1, data.minScale, data.maxScale, data.target, data.limit)
		r.GetRevision("rev1", &rev1, nil)

		revList := servingv1.RevisionList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RevisionList",
				APIVersion: "serving.knative.dev/v1",
			},
			Items: []servingv1.Revision{
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
		assert.Assert(t, cmp.Regexp("Cluster:\\s+https://foo.default.svc.cluster.local", output))

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
		client := knclient.NewMockKnServiceClient(t)

		// Recording:
		r := client.Recorder()

		// Prepare service
		expectedService := createTestService("foo", []string{"rev1"}, goodConditions())

		// Get service & revision
		r.GetService("foo", &expectedService, nil)
		rev1 := createTestRevision("rev1", 1)
		addResourceLimits(&rev1.Spec.Containers[0].Resources, data.reqMem, data.limitMem, data.reqCPU, data.limitCPU)
		r.GetRevision("rev1", &rev1, nil)

		revList := servingv1.RevisionList{
			TypeMeta: metav1.TypeMeta{
				Kind:       "RevisionList",
				APIVersion: "serving.knative.dev/v1",
			},
			Items: []servingv1.Revision{
				rev1,
			},
		}

		// Return the list of all revisions
		r.ListRevisions(knclient.HasLabelSelector(api_serving.ServiceLabelKey, "foo"), &revList, nil)

		// Testing:
		output, err := executeServiceCommand(client, "describe", "foo", "--verbose")
		assert.NilError(t, err)

		validateServiceOutput(t, "foo", output)

		assert.Assert(t, cmp.Regexp("Cluster:\\s+https://foo.default.svc.cluster.local", output))

		validateOutputLine(t, output, "Memory", data.memoryOut)
		validateOutputLine(t, output, "CPU", data.cpuOut)

		// Validate that all recorded API methods have been called
		r.Validate()
	}
}

func TestServiceDescribeUserImageVsImage(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	// Prepare service
	expectedService := createTestService("foo", []string{"rev1", "rev2", "rev3", "rev4"}, goodConditions())
	r.GetService("foo", &expectedService, nil)

	rev1 := createTestRevision("rev1", 1)
	rev2 := createTestRevision("rev2", 2)
	rev3 := createTestRevision("rev3", 3)
	rev4 := createTestRevision("rev4", 4)

	// Different combinations of image annotations and not.
	rev1.Spec.Containers[0].Image = "gcr.io/test/image:latest"
	rev1.Annotations[client_serving.UserImageAnnotationKey] = "gcr.io/test/image:latest"
	rev2.Spec.Containers[0].Image = "gcr.io/test/image@" + imageDigest
	rev2.Annotations[client_serving.UserImageAnnotationKey] = "gcr.io/test/image:latest"
	// rev3 is as if we changed the image but didn't change the annotation
	rev3.Annotations[client_serving.UserImageAnnotationKey] = "gcr.io/test/image:latest"
	rev3.Spec.Containers[0].Image = "gcr.io/a/b"
	// rev4 is without the annotation at all and no hash
	rev4.Status.ImageDigest = ""
	rev4.Spec.Containers[0].Image = "gcr.io/x/y"

	// Fetch the revisions
	r.GetRevision("rev1", &rev1, nil)
	r.GetRevision("rev2", &rev2, nil)
	r.GetRevision("rev3", &rev3, nil)
	r.GetRevision("rev4", &rev4, nil)

	// Testing:
	output, err := executeServiceCommand(client, "describe", "foo")
	assert.NilError(t, err)

	validateServiceOutput(t, "foo", output)

	assert.Assert(t, util.ContainsAll(output, "Image", "Name",
		"gcr.io/test/image:latest (pinned to 123456)", "gcr.io/a/b (at 123456)", "gcr.io/x/y"))
	assert.Assert(t, util.ContainsAll(output, "[1]", "[2]"))

	// Validate that all recorded API methods have been called
	r.Validate()

}

func TestServiceDescribeVerbose(t *testing.T) {

	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	// Prepare service
	expectedService := createTestService("foo", []string{"rev1", "rev2"}, goodConditions())
	r.GetService("foo", &expectedService, nil)

	rev1 := createTestRevision("rev1", 1)
	rev2 := createTestRevision("rev2", 2)

	revList := servingv1.RevisionList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RevisionList",
			APIVersion: "serving.knative.dev/v1",
		},
		Items: []servingv1.Revision{
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

	assert.Assert(t, cmp.Regexp("Cluster:\\s+https://foo.default.svc.cluster.local", output))
	assert.Assert(t, util.ContainsAll(output, "Image", "Name", "gcr.io/test/image (at 123456)", "50%", "(0s)"))
	assert.Assert(t, util.ContainsAll(output, "Env:", "env1=eval1\n", "env2=eval2\n"))
	assert.Assert(t, util.ContainsAll(output, "EnvFrom:", "cm:test1\n", "cm:test2\n"))
	assert.Assert(t, util.ContainsAll(output, "Annotations:", "anno1=aval1\n", "anno2=aval2\n"))
	assert.Assert(t, util.ContainsAll(output, "Labels:", "label1=lval1\n", "label2=lval2\n"))
	assert.Assert(t, util.ContainsAll(output, "[1]", "[2]"))

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceDescribeWithWrongArguments(t *testing.T) {
	client := knclient.NewMockKnServiceClient(t)
	_, err := executeServiceCommand(client, "describe")
	assert.ErrorContains(t, err, "requires", "name", "service", "single", "argument")

	_, err = executeServiceCommand(client, "describe", "foo", "bar")
	assert.ErrorContains(t, err, "requires", "name", "service", "single", "argument")
}

func TestServiceDescribeMachineReadable(t *testing.T) {
	client := knclient.NewMockKnServiceClient(t)

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
	assert.Assert(t, cmp.Regexp("URL:\\s+https://"+service+".default.example.com", output))

	assert.Assert(t, util.ContainsAll(output, "Age:", "Revisions:", "Conditions:", "Labels:", "Annotations:"))
	assert.Assert(t, util.ContainsAll(output, "Ready", "RoutesReady", "OK", "TYPE", "AGE", "REASON"))
}

func createTestService(name string, revisionNames []string, conditions duckv1.Conditions) servingv1.Service {

	labelMap := make(map[string]string)
	labelMap["label1"] = "lval1"
	labelMap["label2"] = "lval2"
	annoMap := make(map[string]string)
	annoMap["anno1"] = "aval1"
	annoMap["anno2"] = "aval2"
	annoMap["anno3"] = "very_long_value_which_should_be_truncated_in_normal_output_if_we_make_it_even_longer"

	serviceUrl, _ := apis.ParseURL(fmt.Sprintf("https://%s.default.svc.cluster.local", name))
	addressUrl, _ := apis.ParseURL(fmt.Sprintf("https://%s.default.example.com", name))
	service := servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			Labels:            labelMap,
			Annotations:       annoMap,
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-30 * time.Second)},
		},
		Status: servingv1.ServiceStatus{
			RouteStatusFields: servingv1.RouteStatusFields{
				URL:     addressUrl,
				Address: &duckv1.Addressable{URL: serviceUrl},
			},
			Status: duckv1.Status{
				Conditions: conditions,
			},
		},
	}
	service.Status.LatestCreatedRevisionName = revisionNames[len(revisionNames)-1]
	service.Status.LatestReadyRevisionName = revisionNames[len(revisionNames)-1]

	if len(revisionNames) > 0 {
		trafficTargets := make([]servingv1.TrafficTarget, 0)
		for _, rname := range revisionNames {
			url, _ := apis.ParseURL(fmt.Sprintf("https://%s", rname))
			target := servingv1.TrafficTarget{
				RevisionName:      rname,
				ConfigurationName: name,
				Percent:           ptr.Int64(int64(100 / len(revisionNames))),
				URL:               url,
			}
			trafficTargets = append(trafficTargets, target)
		}
		service.Status.Traffic = trafficTargets
	}

	return service
}

func createTestServiceWithServiceAccount(name string, revisionNames []string, serviceAccountName string, conditions duckv1.Conditions) servingv1.Service {
	service := createTestService(name, revisionNames, conditions)

	if serviceAccountName != "" {
		template := servingv1.RevisionTemplateSpec{
			Spec: servingv1.RevisionSpec{
				PodSpec: v1.PodSpec{
					ServiceAccountName: serviceAccountName,
				},
			},
		}
		service.Spec.Template = template
	}

	return service
}

func addScaling(revision *servingv1.Revision, minScale, maxScale, concurrencyTarget, concurrencyLimit string) {
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
	if concurrencyLimit != "" {
		l, _ := strconv.ParseInt(concurrencyLimit, 10, 64)
		revision.Spec.ContainerConcurrency = ptr.Int64(l)
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

func createTestRevision(revision string, gen int64) servingv1.Revision {
	labels := make(map[string]string)
	labels[api_serving.ConfigurationGenerationLabelKey] = fmt.Sprintf("%d", gen)

	return servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              revision,
			Namespace:         "default",
			Generation:        1,
			Labels:            labels,
			Annotations:       make(map[string]string),
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Spec: servingv1.RevisionSpec{
			PodSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image: "gcr.io/test/image",
						Env: []v1.EnvVar{
							{Name: "env1", Value: "eval1"},
							{Name: "env2", Value: "eval2"},
						},
						EnvFrom: []v1.EnvFromSource{
							{ConfigMapRef: &v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: "test1"}}},
							{ConfigMapRef: &v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: "test2"}}},
						},
						Ports: []v1.ContainerPort{
							{ContainerPort: 8080},
						},
					},
				},
			},
		},
		Status: servingv1.RevisionStatus{
			ImageDigest: "gcr.io/test/image@" + imageDigest,
			Status: duckv1.Status{
				Conditions: goodConditions(),
			},
		},
	}
}

func goodConditions() duckv1.Conditions {
	ret := make(duckv1.Conditions, 0)
	ret = append(ret, apis.Condition{
		Type:   apis.ConditionReady,
		Status: v1.ConditionTrue,
		LastTransitionTime: apis.VolatileTime{
			Inner: metav1.Time{Time: time.Now()},
		},
	})
	ret = append(ret, apis.Condition{
		Type:   servingv1.ServiceConditionRoutesReady,
		Status: v1.ConditionTrue,
		LastTransitionTime: apis.VolatileTime{
			Inner: metav1.Time{Time: time.Now()},
		},
	})
	return ret
}
