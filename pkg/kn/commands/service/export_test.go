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
	"strconv"
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

type expectedServiceOption func(*servingv1.Service)
type expectedRevisionOption func(*servingv1.Revision)

func TestServiceExport(t *testing.T) {

	svcs := []*servingv1.Service{
		getServiceWithOptions(getService("foo"), withContainer()),
		getServiceWithOptions(getService("foo"), withContainer(), withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}})),
		getServiceWithOptions(getService("foo"), withContainer(), withLabels(map[string]string{"a": "mouse", "b": "cookie", "empty": ""})),
		getServiceWithOptions(getService("foo"), withContainer(), withEnvFrom([]string{"cm-name"})),
		getServiceWithOptions(getService("foo"), withContainer(), withVolumeandSecrets("volName", "secretName")),
	}

	for _, svc := range svcs {
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

	actSvc := servingv1.Service{}
	err = yaml.Unmarshal([]byte(output), &actSvc)
	assert.NilError(t, err)
	stripExpectedSvcVariables(expectedService)
	assert.DeepEqual(t, expectedService, &actSvc)
	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceExportwithMultipleRevisions(t *testing.T) {
	//case 1 - 2 revisions with traffic split
	expSvc1 := getServiceWithOptions(getService("foo"), withContainer(), withServiceRevisionName("foo-rev-1"))
	stripExpectedSvcVariables(expSvc1)
	expSvc2 := getServiceWithOptions(getService("foo"), withContainer(), withTrafficSplit([]string{"foo-rev-1", "foo-rev-2"}, []int{50, 50}, []string{"latest", "current"}), withServiceRevisionName("foo-rev-2"))
	stripExpectedSvcVariables(expSvc2)
	latestSvc := getServiceWithOptions(getService("foo"), withContainer(), withTrafficSplit([]string{"foo-rev-1", "foo-rev-2"}, []int{50, 50}, []string{"latest", "current"}))

	expSvcList := servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []servingv1.Service{*expSvc1, *expSvc2},
	}

	multiRevs := getRevisionList("rev", "foo")

	callServiceExportHistoryTest(t, latestSvc, multiRevs, &expSvcList)

	// case 2 - same revisions no traffic split
	expSvc2 = getServiceWithOptions(getService("foo"), withContainer(), withServiceRevisionName("foo-rev-2"))
	stripExpectedSvcVariables(expSvc2)
	expSvcList = servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []servingv1.Service{*expSvc2},
	}
	latestSvc = getServiceWithOptions(getService("foo"), withContainer(), withTrafficSplit([]string{"foo-rev-2"}, []int{100}, []string{"latest"}))
	callServiceExportHistoryTest(t, latestSvc, multiRevs, &expSvcList)
}

func callServiceExportHistoryTest(t *testing.T, latestSvc *servingv1.Service, revs *servingv1.RevisionList, expSvcList *servingv1.ServiceList) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)
	// Recording:
	r := client.Recorder()

	r.GetService(latestSvc.ObjectMeta.Name, latestSvc, nil)
	r.ListRevisions(mock.Any(), revs, nil)

	output, err := executeServiceCommand(client, "export", latestSvc.ObjectMeta.Name, "--with-revisions", "-o", "json")
	assert.NilError(t, err)

	actSvcList := servingv1.ServiceList{}
	err = json.Unmarshal([]byte(output), &actSvcList)
	assert.NilError(t, err)
	assert.DeepEqual(t, expSvcList, &actSvcList)
	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceExportError(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	expectedService := getService("foo")

	_, err := executeServiceCommand(client, "export", expectedService.ObjectMeta.Name)

	assert.Error(t, err, "'kn service export' requires output format")
}

func getRevisionList(revision string, service string) *servingv1.RevisionList {
	rev1 := getRevisionWithOptions(
		service,
		withRevisionGeneration("1"),
		withRevisionName(fmt.Sprintf("%s-%s-%d", service, revision, 1)),
	)

	rev2 := getRevisionWithOptions(
		service,
		withRevisionGeneration("2"),
		withRevisionName(fmt.Sprintf("%s-%s-%d", service, revision, 2)),
	)

	return &servingv1.RevisionList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []servingv1.Revision{rev1, rev2},
	}
}

func stripExpectedSvcVariables(expectedsvc *servingv1.Service) {
	expectedsvc.ObjectMeta.Namespace = ""
	expectedsvc.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{}
	expectedsvc.Status = servingv1.ServiceStatus{}
	expectedsvc.ObjectMeta.Annotations = nil
	expectedsvc.ObjectMeta.CreationTimestamp = metav1.Time{}
}

func getRevisionWithOptions(service string, options ...expectedRevisionOption) servingv1.Revision {
	rev := servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Labels: map[string]string{
				apiserving.ServiceLabelKey: service,
			},
		},
		Spec: servingv1.RevisionSpec{
			PodSpec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image: "gcr.io/foo/bar:baz",
					},
				},
			},
		},
	}
	for _, fn := range options {
		fn(&rev)
	}
	return rev
}

func getServiceWithOptions(svc *servingv1.Service, options ...expectedServiceOption) *servingv1.Service {
	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "service",
		APIVersion: "serving.knative.dev/v1",
	}

	for _, fn := range options {
		fn(svc)
	}

	return svc
}

func withLabels(labels map[string]string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.ObjectMeta.Labels = labels
	}
}

func withEnvFrom(cmNames []string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		var list []v1.EnvFromSource
		for _, cmName := range cmNames {
			list = append(list, v1.EnvFromSource{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: cmName,
					},
				},
			})
		}
		svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].EnvFrom = list
	}
}

func withEnv(env []v1.EnvVar) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Env = env
	}
}

func withContainer() expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.Spec.Containers[0].Image = "gcr.io/foo/bar:baz"
		svc.Spec.ConfigurationSpec.Template.Annotations = map[string]string{servinglib.UserImageAnnotationKey: "gcr.io/foo/bar:baz"}

	}
}

func withVolumeandSecrets(volName string, secretName string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		template := &svc.Spec.Template
		template.Spec.Volumes = []v1.Volume{
			{
				Name: volName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			},
		}

		template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
			{
				Name:      volName,
				MountPath: "/mount/path",
				ReadOnly:  true,
			},
		}
	}
}

func withRevisionGeneration(gen string) expectedRevisionOption {
	return func(rev *servingv1.Revision) {
		i, _ := strconv.Atoi(gen)
		rev.ObjectMeta.Generation = int64(i)
		rev.ObjectMeta.Labels[apiserving.ConfigurationGenerationLabelKey] = gen
	}
}

func withRevisionName(name string) expectedRevisionOption {
	return func(rev *servingv1.Revision) {
		rev.ObjectMeta.Name = name
	}
}

func withServiceRevisionName(name string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.ObjectMeta.Name = name
	}
}

func withTrafficSplit(revisions []string, percentages []int, tags []string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		var trafficTargets []servingv1.TrafficTarget
		for i, rev := range revisions {
			trafficTargets = append(trafficTargets, servingv1.TrafficTarget{
				Percent: ptr.Int64(int64(percentages[i])),
			})
			if tags[i] != "" {
				trafficTargets[i].Tag = tags[i]
			}
			if rev == "latest" {
				trafficTargets[i].LatestRevision = ptr.Bool(true)
			} else {
				trafficTargets[i].RevisionName = rev
				trafficTargets[i].LatestRevision = ptr.Bool(false)
			}
		}
		svc.Spec.RouteSpec = servingv1.RouteSpec{
			Traffic: trafficTargets,
		}
	}
}
