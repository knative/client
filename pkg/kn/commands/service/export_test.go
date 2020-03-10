// Copyright © 2020 The Knative Authors
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
	"encoding/json"
	"fmt"
	"testing"

	"gotest.tools/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servinglib "knative.dev/client/pkg/serving"
	knclient "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util/mock"
	"knative.dev/pkg/ptr"
	apiserving "knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/yaml"
)

func TestServiceExport(t *testing.T) {
	var svcs []*servingv1.Service
	typeMeta := metav1.TypeMeta{
		Kind:       "service",
		APIVersion: "serving.knative.dev/v1",
	}

	// case 1 - plain svc
	plainService := getService("foo")
	svcs = append(svcs, plainService)

	// case 2 - svc with env variables
	envSvc := getService("foo")
	envVars := []v1.EnvVar{
		{Name: "a", Value: "mouse"},
		{Name: "b", Value: "cookie"},
		{Name: "empty", Value: ""},
	}
	template := &envSvc.Spec.Template
	template.Spec.Containers[0].Env = envVars
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	svcs = append(svcs, envSvc)

	//case 3 - svc with labels
	labelService := getService("foo")
	expected := map[string]string{
		"a":     "mouse",
		"b":     "cookie",
		"empty": "",
	}
	labelService.Labels = expected
	labelService.Spec.Template.Annotations = map[string]string{
		servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz",
	}
	template = &labelService.Spec.Template
	template.ObjectMeta.Labels = expected
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	svcs = append(svcs, labelService)

	//case 4 - config map
	CMservice := getService("foo")
	template = &CMservice.Spec.Template
	template.Spec.Containers[0].EnvFrom = []v1.EnvFromSource{
		{
			ConfigMapRef: &v1.ConfigMapEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: "config-map-name",
				},
			},
		},
	}
	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}
	svcs = append(svcs, CMservice)

	//case 5 - volume mount and secrets
	Volservice := getService("foo")
	template = &Volservice.Spec.Template
	template.Spec.Volumes = []v1.Volume{
		{
			Name: "volume-name",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "secret-name",
				},
			},
		},
	}

	template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
		{
			Name:      "volume-name",
			MountPath: "/mount/path",
			ReadOnly:  true,
		},
	}
	svcs = append(svcs, Volservice)

	template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
	template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	for _, svc := range svcs {
		svc.TypeMeta = typeMeta
		callServiceExportTest(t, svc)
	}

}

func callServiceExportTest(t *testing.T, expectedService *servingv1.Service) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.GetService(expectedService.ObjectMeta.Name, expectedService, nil)

	output, err := executeServiceCommand(client, "export", expectedService.ObjectMeta.Name, "-o", "yaml")

	assert.NilError(t, err)

	expectedService.ObjectMeta.Namespace = ""

	expSvcYaml, err := yaml.Marshal(expectedService)

	assert.NilError(t, err)

	assert.Equal(t, string(expSvcYaml), output)

	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceExportwithMultipleRevisions(t *testing.T) {
	//case 1 = 2 revisions with traffic split
	trafficSplitService := createServiceTwoRevsionsWithTraffic("foo", true)

	multiRevs := createTestRevisionList("rev", "foo")

	callServiceExportHistoryTest(t, trafficSplitService, multiRevs)

	//case 2 - same revisions no traffic split
	noTrafficSplitService := createServiceTwoRevsionsWithTraffic("foo", false)

	callServiceExportHistoryTest(t, noTrafficSplitService, multiRevs)
}

func callServiceExportHistoryTest(t *testing.T, expectedService *servingv1.Service, revs *servingv1.RevisionList) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	// Recording:
	r := client.Recorder()

	r.GetService(expectedService.ObjectMeta.Name, expectedService, nil)

	r.ListRevisions(mock.Any(), revs, nil)

	output, err := executeServiceCommand(client, "export", expectedService.ObjectMeta.Name, "--with-revisions", "-o", "json")

	assert.NilError(t, err)

	actSvcList := servingv1.ServiceList{}

	json.Unmarshal([]byte(output), &actSvcList)

	for i, actSvc := range actSvcList.Items {
		var checkTraffic bool
		if i == (len(actSvcList.Items) - 1) {
			checkTraffic = true
		}
		validateServiceWithRevisionHistory(t, expectedService, revs, actSvc, checkTraffic)
	}

	// Validate that all recorded API methods have been called
	r.Validate()
}

func validateServiceWithRevisionHistory(t *testing.T, expectedsvc *servingv1.Service, expectedRevList *servingv1.RevisionList, actualSvc servingv1.Service, checkTraffic bool) {
	var expectedRev servingv1.Revision
	var routeSpec servingv1.RouteSpec
	for _, rev := range expectedRevList.Items {
		if actualSvc.Spec.ConfigurationSpec.Template.ObjectMeta.Name == rev.ObjectMeta.Name {
			expectedRev = rev
			break
		}
	}
	expectedsvc.Spec.ConfigurationSpec.Template.ObjectMeta.Name = expectedRev.ObjectMeta.Name
	expectedsvc.Spec.Template.Spec = expectedRev.Spec

	stripExpectedSvcVariables(expectedsvc)

	if !checkTraffic {
		routeSpec = expectedsvc.Spec.RouteSpec
		expectedsvc.Spec.RouteSpec = servingv1.RouteSpec{}
	}
	assert.DeepEqual(t, expectedsvc, &actualSvc)

	expectedsvc.Spec.RouteSpec = routeSpec
}

func TestServiceExportError(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	expectedService := getService("foo")

	_, err := executeServiceCommand(client, "export", expectedService.ObjectMeta.Name)

	assert.Error(t, err, "'kn service export' requires output format")
}

func createTestRevisionList(revision string, service string) *servingv1.RevisionList {
	labels1 := make(map[string]string)
	labels1[apiserving.ConfigurationGenerationLabelKey] = "1"
	labels1[apiserving.ServiceLabelKey] = service

	labels2 := make(map[string]string)
	labels2[apiserving.ConfigurationGenerationLabelKey] = "2"
	labels2[apiserving.ServiceLabelKey] = service

	rev1 := servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       fmt.Sprintf("%s-%s-%d", service, revision, 1),
			Namespace:  "default",
			Generation: int64(1),
			Labels:     labels1,
		},
		Spec: servingv1.RevisionSpec{
			PodSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image: "gcr.io/test/image:v1",
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
	}

	rev2 := rev1

	rev2.Spec.PodSpec.Containers[0].Image = "gcr.io/test/image:v2"
	rev2.ObjectMeta.Labels = labels2
	rev2.ObjectMeta.Generation = int64(2)
	rev2.ObjectMeta.Name = fmt.Sprintf("%s-%s-%d", service, revision, 2)

	typeMeta := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "List",
	}

	return &servingv1.RevisionList{
		TypeMeta: typeMeta,
		Items:    []servingv1.Revision{rev1, rev2},
	}
}

func createServiceTwoRevsionsWithTraffic(svc string, trafficSplit bool) *servingv1.Service {
	expectedService := createTestService(svc, []string{svc + "-rev-1", svc + "-rev-2"}, goodConditions())
	expectedService.Status.Traffic[0].LatestRevision = ptr.Bool(true)
	expectedService.Status.Traffic[0].Tag = "latest"
	expectedService.Status.Traffic[1].Tag = "current"

	if trafficSplit {
		trafficList := []servingv1.TrafficTarget{
			{
				RevisionName: "foo-rev-1",
				Percent:      ptr.Int64(int64(50)),
			}, {
				RevisionName: "foo-rev-2",
				Percent:      ptr.Int64(int64(50)),
			}}
		expectedService.Spec.RouteSpec = servingv1.RouteSpec{Traffic: trafficList}
	} else {
		trafficList := []servingv1.TrafficTarget{
			{
				RevisionName: "foo-rev-2",
				Percent:      ptr.Int64(int64(50)),
			}}
		expectedService.Spec.RouteSpec = servingv1.RouteSpec{Traffic: trafficList}
	}

	return &expectedService
}

func stripExpectedSvcVariables(expectedsvc *servingv1.Service) {
	expectedsvc.ObjectMeta.Namespace = ""
	expectedsvc.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{}
	expectedsvc.Status = servingv1.ServiceStatus{}
	expectedsvc.ObjectMeta.Annotations = nil
	expectedsvc.ObjectMeta.CreationTimestamp = metav1.Time{}
}
