// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e
// +build !eventing

package e2e

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"

	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/ptr"
	"sigs.k8s.io/yaml"

	"knative.dev/client/lib/test"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type expectedServiceOption func(*servingv1.Service)

func TestServiceExportImportApply(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("create service with byo revision")
	serviceCreateWithOptions(r, "hello", "--revision-name", "rev1")

	t.Log("export service and compare")
	serviceExport(r, "hello", getSvc(withName("hello"), withRevisionName("hello-rev1"), withAnnotations()), "-o", "json")

	t.Log("update service - add env variable")
	serviceUpdateWithOptions(r, "hello", "--env", "key1=val1", "--revision-name", "rev2", "--no-lock-to-digest")
	serviceExport(r, "hello", getSvc(withName("hello"), withRevisionName("hello-rev2"), withEnv("key1", "val1")), "-o", "json")
	serviceExportWithRevisions(r, "hello", getSvcListWithOneRevision(), "--with-revisions", "-o", "yaml")

	t.Log("update service with tag and split traffic")
	serviceUpdateWithOptions(r, "hello", "--tag", "hello-rev1=candidate", "--traffic", "candidate=2%,@latest=98%")
	serviceExportWithRevisions(r, "hello", getSvcListWithTags(), "--with-revisions", "-o", "yaml")

	t.Log("update service - untag, add env variable and traffic split")
	serviceUpdateWithOptions(r, "hello", "--untag", "candidate")
	serviceUpdateWithOptions(r, "hello", "--env", "key2=val2", "--revision-name", "rev3", "--traffic", "hello-rev1=30,hello-rev2=30,hello-rev3=40")
	serviceExportWithRevisions(r, "hello", getSvcListWOTags(), "--with-revisions", "-o", "yaml")
}

// Private methods

func serviceExport(r *test.KnRunResultCollector, serviceName string, expService servingv1.Service, options ...string) {
	command := []string{"service", "export", serviceName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	validateExportedService(r.T(), r.KnTest(), out.Stdout, expService)
}

func serviceExportWithRevisions(r *test.KnRunResultCollector, serviceName string, expServiceList servingv1.ServiceList, options ...string) {
	command := []string{"service", "export", serviceName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	validateExportedServiceList(r.T(), r.KnTest(), out.Stdout, expServiceList)
}

// Private functions

func validateExportedService(t *testing.T, it *test.KnTest, out string, expService servingv1.Service) {
	actSvcJSON := servingv1.Service{}
	err := json.Unmarshal([]byte(out), &actSvcJSON)
	assert.NilError(t, err)
	assert.DeepEqual(t, &expService, &actSvcJSON)
}

func validateExportedServiceList(t *testing.T, it *test.KnTest, out string, expServiceList servingv1.ServiceList) {
	actYaml := servingv1.ServiceList{}
	err := yaml.Unmarshal([]byte(out), &actYaml)
	assert.NilError(t, err)
	assert.DeepEqual(t, &expServiceList, &actYaml)
}

func getSvc(options ...expectedServiceOption) servingv1.Service {
	svc := servingv1.Service{
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						ContainerConcurrency: ptr.Int64(int64(0)),
						TimeoutSeconds:       ptr.Int64(int64(300)),
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:      "user-container",
									Image:     test.KnDefaultTestImage,
									Resources: corev1.ResourceRequirements{},
									ReadinessProbe: &corev1.Probe{
										SuccessThreshold: int32(1),
										Handler: corev1.Handler{
											TCPSocket: &corev1.TCPSocketAction{
												Port: intstr.FromInt(0),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "",
		},
	}
	for _, fn := range options {
		fn(&svc)
	}
	return svc
}

func getSvcListWOTags() servingv1.ServiceList {
	return servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []servingv1.Service{
			getSvc(
				withName("hello"),
				withRevisionName("hello-rev1"),
			),
			getSvc(
				withName("hello"),
				withRevisionName("hello-rev2"),
				withEnv("key1", "val1"),
			),
			getSvc(
				withName("hello"),
				withRevisionName("hello-rev3"),
				withEnv("key1", "val1"), withEnv("key2", "val2"),
				withTrafficSplit([]string{"hello-rev1", "hello-rev2", "hello-rev3"}, []int{30, 30, 40}, []string{"", "", ""}),
			),
		},
	}
}

func getSvcListWithTags() servingv1.ServiceList {
	return servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []servingv1.Service{
			getSvc(
				withName("hello"),
				withRevisionName("hello-rev1"),
			),
			getSvc(
				withName("hello"),
				withRevisionName("hello-rev2"),
				withEnv("key1", "val1"),
				withTrafficSplit([]string{"latest", "hello-rev1"}, []int{98, 2}, []string{"", "candidate"}),
			),
		},
	}
}

func getSvcListWithOneRevision() servingv1.ServiceList {
	return servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: []servingv1.Service{
			getSvc(
				withName("hello"),
				withRevisionName("hello-rev2"),
				withEnv("key1", "val1"),
				withTrafficSplit([]string{"latest"}, []int{100}, []string{""}),
			),
		},
	}
}

func withRevisionName(revName string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.ObjectMeta.Name = revName
	}
}

func withAnnotations() expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.ObjectMeta.Annotations = map[string]string{
			"client.knative.dev/user-image": "gcr.io/knative-samples/helloworld-go",
		}
	}
}

func withName(name string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.ObjectMeta.Name = name
	}
}

func withEnv(key string, val string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		env := []corev1.EnvVar{
			{
				Name:  key,
				Value: val,
			},
		}
		currentEnv := svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Env
		if len(currentEnv) > 0 {
			svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Env = append(currentEnv, env...)
		} else {
			svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0].Env = env
		}

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
