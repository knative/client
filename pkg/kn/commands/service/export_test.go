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

type testCase struct {
	name                 string
	latestSvc            *servingv1.Service
	expectedSvcList      *servingv1.ServiceList
	revisionList         *servingv1.RevisionList
	expectedRevisionList *servingv1.RevisionList
}

func TestServiceExportError(t *testing.T) {
	tc := &testCase{latestSvc: getService("foo")}

	_, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name)
	assert.Error(t, err, "'kn service export' requires output format")

	_, err = executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "-o", "json")
	assert.Error(t, err, "'kn service export --with-revisions' requires a mode, please specify one of kubernetes|resources.")

	_, err = executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "--mode", "k8s", "-o", "yaml")
	assert.Error(t, err, "'kn service export --with-revisions' requires a mode, please specify one of kubernetes|resources.")
}

func TestServiceExport(t *testing.T) {

	for _, tc := range []testCase{
		{latestSvc: getServiceWithOptions(getService("foo"), withServicePodSpecOption(withContainer()))},
		{latestSvc: getServiceWithOptions(getService("foo"), withServicePodSpecOption(withContainer(), withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}})))},
		{latestSvc: getServiceWithOptions(getService("foo"), withConfigurationLabels(map[string]string{"a": "mouse"}), withConfigurationAnnotations(map[string]string{"a": "mouse"}), withServicePodSpecOption(withContainer()))},
		{latestSvc: getServiceWithOptions(getService("foo"), withLabels(map[string]string{"a": "mouse"}), withAnnotations(map[string]string{"a": "mouse"}), withServicePodSpecOption(withContainer()))},
		{latestSvc: getServiceWithOptions(getService("foo"), withServicePodSpecOption(withContainer(), withVolumeandSecrets("secretName")))},
	} {
		exportServiceTest(t, &tc)
	}
}

func exportServiceTest(t *testing.T, tc *testCase) {
	output, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "-o", "yaml")
	assert.NilError(t, err)

	actSvc := servingv1.Service{}
	err = yaml.Unmarshal([]byte(output), &actSvc)
	assert.NilError(t, err)

	stripUnwantedFields(tc.latestSvc)
	assert.DeepEqual(t, tc.latestSvc, &actSvc)
}

func TestServiceExportwithMultipleRevisions(t *testing.T) {
	for _, tc := range []testCase{{
		name: "test 2 revisions with traffic split",
		latestSvc: getServiceWithOptions(
			getService("foo"),
			withAnnotations(map[string]string{"serving.knative.dev/creator": "ut", "serving.knative.dev/lastModifier": "ut"}),
			withTrafficSplit([]string{"foo-rev-1", ""}, []int{50, 50}, []bool{false, true}),
			withServicePodSpecOption(withContainer()),
		),
		expectedSvcList: getServiceListWithOptions(
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				withServicePodSpecOption(withContainer()),
				withServiceRevisionName("foo-rev-1"),
			),
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				withServicePodSpecOption(withContainer()),
				withTrafficSplit([]string{"foo-rev-1", ""}, []int{50, 50}, []bool{false, true}),
			),
		),
		revisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("1"),
				withRevisionAnnotations(map[string]string{"serving.knative.dev/lastPinned": "1111132"}),
				withRevisionName("foo-rev-1"),
				withRevisionPodSpecOption(withContainer()),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("2"),
				withRevisionName("foo-rev-2"),
				withRevisionPodSpecOption(withContainer()),
			),
		),
		expectedRevisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionName("foo-rev-1"),
				withRevisionGeneration("1"),
				withRevisionPodSpecOption(withContainer()),
			),
		),
	}, {
		name: "test 2 revisions no traffic split",
		latestSvc: getServiceWithOptions(
			getService("foo"),
			withTrafficSplit([]string{""}, []int{100}, []bool{true}),
			withServicePodSpecOption(withContainer()),
		),
		expectedSvcList: getServiceListWithOptions(
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				withServicePodSpecOption(withContainer()),
				withTrafficSplit([]string{""}, []int{100}, []bool{true}),
			),
		),
		revisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("1"),
				withRevisionName("foo-rev-1"),
				withRevisionPodSpecOption(withContainer()),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("2"),
				withRevisionName("foo-rev-2"),
				withRevisionPodSpecOption(withContainer()),
			),
		),
	}, {
		name: "test 3 active revisions with traffic split with no latest revision",
		latestSvc: getServiceWithOptions(
			getService("foo"),
			withTrafficSplit([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []bool{false, false, false}),
			withServiceRevisionName("foo-rev-3"),
			withServicePodSpecOption(
				withContainer(),
				withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}}),
			),
		),
		expectedSvcList: getServiceListWithOptions(
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				withServicePodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "cat"}}),
				),
				withServiceRevisionName("foo-rev-1"),
			),
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				withServicePodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "dog"}}),
				),
				withServiceRevisionName("foo-rev-2"),
			),
			withServices(
				getService("foo"),
				withUnwantedFieldsStripped(),
				withServicePodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "mouse"}}),
				),
				withServiceRevisionName("foo-rev-3"),
				withTrafficSplit([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []bool{false, false, false}),
			),
		),
		revisionList: getRevisionListWithOptions(
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("1"),
				withRevisionName("foo-rev-1"),
				withRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "cat"}}),
				),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("2"),
				withRevisionName("foo-rev-2"),
				withRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "dog"}}),
				),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionGeneration("3"),
				withRevisionName("foo-rev-3"),
				withRevisionPodSpecOption(
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
				withRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "cat"}}),
				),
			),
			withRevisions(
				withRevisionLabels(map[string]string{apiserving.ServiceLabelKey: "foo"}),
				withRevisionName("foo-rev-2"),
				withRevisionGeneration("2"),
				withRevisionPodSpecOption(
					withContainer(),
					withEnv([]v1.EnvVar{{Name: "a", Value: "dog"}}),
				),
			),
		),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			exportWithRevisionsforKubernetesTest(t, &tc)
			exportWithRevisionsTest(t, &tc)
		})
	}
}

func exportWithRevisionsforKubernetesTest(t *testing.T, tc *testCase) {
	output, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "--mode", "kubernetes", "-o", "json")
	assert.NilError(t, err)

	actSvcList := servingv1.ServiceList{}
	err = json.Unmarshal([]byte(output), &actSvcList)
	assert.NilError(t, err)
	assert.DeepEqual(t, tc.expectedSvcList, &actSvcList)
}

func exportWithRevisionsTest(t *testing.T, tc *testCase) {
	output, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "--mode", "resources", "-o", "json")
	assert.NilError(t, err)

	stripUnwantedFields(tc.latestSvc)
	expOut := strings.Builder{}
	expSvcJSON, err := json.MarshalIndent(tc.latestSvc, "", "    ")
	assert.NilError(t, err)
	expOut.Write(expSvcJSON)
	expOut.WriteString("\n")

	if tc.expectedRevisionList != nil {
		expRevsJSON, err := json.MarshalIndent(tc.expectedRevisionList, "", "    ")
		assert.NilError(t, err)
		expOut.Write(expRevsJSON)
		expOut.WriteString("\n")
	}

	assert.Equal(t, expOut.String(), output)
}

func executeServiceExportCommand(t *testing.T, tc *testCase, options ...string) (string, error) {
	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	r.GetService(tc.latestSvc.ObjectMeta.Name, tc.latestSvc, nil)
	r.ListRevisions(mock.Any(), tc.revisionList, nil)

	return executeServiceCommand(client, options...)
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
		svc.Spec.Template.ObjectMeta.Labels = labels
	}
}
func withAnnotations(Annotations map[string]string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.ObjectMeta.Annotations = Annotations
	}
}
func withConfigurationAnnotations(Annotations map[string]string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.Template.ObjectMeta.Annotations = Annotations
	}
}
func withServiceRevisionName(name string) expectedServiceOption {
	return func(svc *servingv1.Service) {
		svc.Spec.Template.ObjectMeta.Name = name
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
func withTrafficSplit(revisions []string, percentages []int, latest []bool) expectedServiceOption {
	return func(svc *servingv1.Service) {
		var trafficTargets []servingv1.TrafficTarget
		for i, rev := range revisions {
			trafficTargets = append(trafficTargets, servingv1.TrafficTarget{
				Percent: ptr.Int64(int64(percentages[i])),
			})
			if latest[i] {
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
func withServicePodSpecOption(options ...podSpecOption) expectedServiceOption {
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
func withRevisionPodSpecOption(options ...podSpecOption) expectedRevisionOption {
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
