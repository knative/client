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
	"strings"
	"testing"

	"gotest.tools/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knclient "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util/mock"
	"knative.dev/pkg/ptr"
	apiserving "knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/yaml"
)

type expectedServiceOption func(*servingv1.Service)
type expectedRevisionOption func(*servingv1.Revision)
type expectedServiceListOption func(*servingv1.ServiceList)
type expectedRevisionListOption func(*servingv1.RevisionList)
type podSpecOption func(*v1.PodSpec)

func TestServiceExportError(t *testing.T) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)

	expectedService := getService("foo")

	_, err := executeServiceCommand(client, "export", expectedService.ObjectMeta.Name)

	assert.Error(t, err, "'kn service export' requires output format")
}

func TestServiceExport(t *testing.T) {

	svcs := []*servingv1.Service{
		getServiceWithOptions(getService("foo"), WithServicePodSpecOption(withContainer())),
		getServiceWithOptions(getService("foo"), WithServicePodSpecOption(withContainer(), withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}}))),
		getServiceWithOptions(getService("foo"), withConfigurationLabels(map[string]string{"a": "mouse"}), withConfigurationAnnotations(map[string]string{"a": "mouse"}), WithServicePodSpecOption(withContainer())),
		getServiceWithOptions(getService("foo"), withLabels(map[string]string{"a": "mouse"}), withAnnotations(map[string]string{"a": "mouse"}), WithServicePodSpecOption(withContainer())),
		getServiceWithOptions(getService("foo"), WithServicePodSpecOption(withContainer(), withVolumeandSecrets("secretName"))),
	}

	for _, svc := range svcs {
		exportServiceTest(t, svc)
	}
}

func exportServiceTest(t *testing.T, expectedService *servingv1.Service) {
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

	stripUnwantedFields(expectedService)
	assert.DeepEqual(t, expectedService, &actSvc)
	// Validate that all recorded API methods have been called
	r.Validate()
}

func TestServiceExportwithMultipleRevisions(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		latestSvc            *servingv1.Service
		expectedSvcList      *servingv1.ServiceList
		revisionList         *servingv1.RevisionList
		expectedRevisionList *servingv1.RevisionList
	}{{
		name: "test 2 revisions with traffic split",
		latestSvc: getServiceWithOptions(
			getService("foo"),
			withTrafficSplit([]string{"foo-rev-1", "foo-rev-2"}, []int{50, 50}, []string{"latest", "current"}),
			withServiceRevisionName("foo-rev-2"),
			WithServicePodSpecOption(withContainer()),
		),
		expectedSvcList: getServiceListWithOptions(
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				WithServicePodSpecOption(withContainer()),
				withServiceRevisionName("foo-rev-1"),
			),
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				WithServicePodSpecOption(withContainer()),
				withServiceRevisionName("foo-rev-2"),
				withTrafficSplit([]string{"foo-rev-1", "foo-rev-2"}, []int{50, 50}, []string{"latest", "current"}),
			),
		),
		revisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("1"),
				withRevisionName("foo-rev-1"),
				WithRevisionPodSpecOption(withContainer()),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("2"),
				withRevisionName("foo-rev-2"),
				WithRevisionPodSpecOption(withContainer()),
			),
		),
		expectedRevisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionName("foo-rev-1"),
				withRevisionGeneration("1"),
				WithRevisionPodSpecOption(withContainer()),
			),
		),
	}, {
		name: "test 2 revisions no traffic split",
		latestSvc: getServiceWithOptions(
			getService("foo"),
			withTrafficSplit([]string{"foo-rev-2"}, []int{100}, []string{"latest"}),
			withServiceRevisionName("foo-rev-2"),
			WithServicePodSpecOption(withContainer()),
		),
		expectedSvcList: getServiceListWithOptions(
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				WithServicePodSpecOption(withContainer()),
				withServiceRevisionName("foo-rev-2"),
				withTrafficSplit([]string{"foo-rev-2"}, []int{100}, []string{"latest"}),
			),
		),
		revisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("1"),
				withRevisionName("foo-rev-1"),
				WithRevisionPodSpecOption(withContainer()),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("2"),
				withRevisionName("foo-rev-2"),
				WithRevisionPodSpecOption(withContainer()),
			),
		),
	}, {
		name: "test 3 active revisions with traffic split",
		latestSvc: getServiceWithOptions(
			getService("foo"),
			withTrafficSplit([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []string{"", "", "latest"}),
			withServiceRevisionName("foo-rev-3"),
			WithServicePodSpecOption(
				withContainer(),
				withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}}),
			),
		),
		expectedSvcList: getServiceListWithOptions(
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				WithServicePodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "cat"}}),
				),
				withServiceRevisionName("foo-rev-1"),
			),
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				WithServicePodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "dog"}}),
				),
				withServiceRevisionName("foo-rev-2"),
			),
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				WithServicePodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}}),
				),
				withServiceRevisionName("foo-rev-3"),
				withTrafficSplit([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []string{"", "", "latest"}),
			),
		),
		revisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("1"),
				withRevisionName("foo-rev-1"),
				WithRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "cat"}}),
				),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("2"),
				withRevisionName("foo-rev-2"),
				WithRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "dog"}}),
				),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("3"),
				withRevisionName("foo-rev-3"),
				WithRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}}),
				),
			),
		),
		expectedRevisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionName("foo-rev-1"),
				withRevisionGeneration("1"),
				WithRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "cat"}}),
				),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionName("foo-rev-2"),
				withRevisionGeneration("2"),
				WithRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "dog"}}),
				),
			),
		),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			exportWithRevisionsforKubernetesTest(t, tc.latestSvc, tc.revisionList, tc.expectedSvcList)
			exportWithRevisionsTest(t, tc.latestSvc, tc.revisionList, tc.expectedRevisionList)
		})
	}
}

func exportWithRevisionsforKubernetesTest(t *testing.T, latestSvc *servingv1.Service, revs *servingv1.RevisionList, expSvcList *servingv1.ServiceList) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)
	// Recording:
	r := client.Recorder()

	r.GetService(latestSvc.ObjectMeta.Name, latestSvc, nil)
	r.ListRevisions(mock.Any(), revs, nil)

	output, err := executeServiceCommand(client, "export", latestSvc.ObjectMeta.Name, "--with-revisions", "--kubernetes-resources", "-o", "json")
	assert.NilError(t, err)

	actSvcList := servingv1.ServiceList{}
	err = json.Unmarshal([]byte(output), &actSvcList)
	assert.NilError(t, err)
	assert.DeepEqual(t, expSvcList, &actSvcList)
	// Validate that all recorded API methods have been called
	r.Validate()
}

func exportWithRevisionsTest(t *testing.T, latestSvc *servingv1.Service, revs *servingv1.RevisionList, expRevList *servingv1.RevisionList) {
	// New mock client
	client := knclient.NewMockKnServiceClient(t)
	// Recording:
	r := client.Recorder()

	r.GetService(latestSvc.ObjectMeta.Name, latestSvc, nil)
	r.ListRevisions(mock.Any(), revs, nil)

	output, err := executeServiceCommand(client, "export", latestSvc.ObjectMeta.Name, "--with-revisions", "-o", "json")
	assert.NilError(t, err)

	stripUnwantedFields(latestSvc)
	expOut := strings.Builder{}
	expSvcJSON, err := json.MarshalIndent(latestSvc, "", "    ")
	assert.NilError(t, err)
	expOut.Write(expSvcJSON)
	expOut.WriteString("\n")

	if expRevList != nil {
		expRevsJSON, err := json.MarshalIndent(expRevList, "", "    ")
		assert.NilError(t, err)
		expOut.Write(expRevsJSON)
		expOut.WriteString("\n")
	}

	assert.Equal(t, expOut.String(), output)
	// Validate that all recorded API methods have been called
	r.Validate()
}

func stripUnwantedFields(svc *servingv1.Service) {
	svc.ObjectMeta.Namespace = ""
	svc.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{}
	svc.Status = servingv1.ServiceStatus{}
	svc.ObjectMeta.CreationTimestamp = metav1.Time{}
}

func getServiceListWithOptions(options ...expectedServiceListOption) *servingv1.ServiceList {
	list := &servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
	}

	for _, fn := range options {
		fn(list)
	}

	return list
}

func withServices(svc *servingv1.Service, options ...expectedServiceOption) expectedServiceListOption {
	return func(list *servingv1.ServiceList) {
		list.Items = append(list.Items, *(getServiceWithOptions(svc, options...)))
	}
}

func getRevisionListWithOptions(options ...expectedRevisionListOption) *servingv1.RevisionList {
	list := &servingv1.RevisionList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
	}

	for _, fn := range options {
		fn(list)
	}

	return list
}

func withRevisions(options ...expectedRevisionOption) expectedRevisionListOption {
	return func(list *servingv1.RevisionList) {
		list.Items = append(list.Items, getRevisionWithOptions(options...))
	}
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
		svc.ObjectMeta.Labels = labels
	}
}
func withConfigurationLabels(labels map[string]string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.ObjectMeta.Labels = labels
	}
}
func withAnnotations(Annotations map[string]string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.ObjectMeta.Annotations = Annotations
	}
}
func withConfigurationAnnotations(Annotations map[string]string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.ObjectMeta.Annotations = Annotations
	}
}
func withServiceRevisionName(name string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.ConfigurationSpec.Template.ObjectMeta.Name = name
	}
}
func withUnwantedFieldsStripped() expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.ObjectMeta.Namespace = ""
		svc.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{}
		svc.Status = servingv1.ServiceStatus{}
		svc.ObjectMeta.CreationTimestamp = metav1.Time{}
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
func WithServicePodSpecOption(options ...podSpecOption) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.Template.Spec.PodSpec = getPodSpecWithOptions(options...)
	}
}

func getRevisionWithOptions(options ...expectedRevisionOption) servingv1.Revision {
	rev := servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
	}
	for _, fn := range options {
		fn(&rev)
	}
	return rev
}
func withRevisionGeneration(gen string) expectedRevisionOption {
	return func(rev *servingv1.Revision) {
		rev.ObjectMeta.Labels[apiserving.ConfigurationGenerationLabelKey] = gen
	}
}
func withRevisionName(name string) expectedRevisionOption {
	return func(rev *servingv1.Revision) {
		rev.ObjectMeta.Name = name
	}
}
func withRevisionLabels(labels map[string]string) expectedRevisionOption {
	return func(rev *servingv1.Revision) {
		rev.ObjectMeta.Labels = labels
	}
}
func withRevisionAnnotations(Annotations map[string]string) expectedRevisionOption {
	return func(rev *servingv1.Revision) {
		rev.ObjectMeta.Annotations = Annotations
	}
}
func WithRevisionPodSpecOption(options ...podSpecOption) expectedRevisionOption {
	return func(rev *servingv1.Revision) {
		rev.Spec.PodSpec = getPodSpecWithOptions(options...)
	}
}

func getPodSpecWithOptions(options ...podSpecOption) v1.PodSpec {
	spec := v1.PodSpec{}
	for _, fn := range options {
		fn(&spec)
	}
	return spec
}

func withEnv(env []v1.EnvVar) podSpecOption {
	return func(spec *v1.PodSpec) {
		spec.Containers[0].Env = env
	}
}
func withContainer() podSpecOption {
	return func(spec *v1.PodSpec) {
		spec.Containers = append(spec.Containers, v1.Container{Image: "gcr.io/foo/bar:baz"})
	}
}
func withVolumeandSecrets(secretName string) podSpecOption {
	return func(spec *v1.PodSpec) {
		spec.Volumes = []v1.Volume{
			{
				Name: secretName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			},
		}
		spec.Containers[0].VolumeMounts = []v1.VolumeMount{
			{
				Name:      secretName,
				MountPath: "/mount/path",
				ReadOnly:  true,
			},
		}
	}
}
