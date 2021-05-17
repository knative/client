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
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	v1 "k8s.io/api/core/v1"
	libtest "knative.dev/client/lib/test"
	clientv1alpha1 "knative.dev/client/pkg/apis/client/v1alpha1"
	knclient "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/util/mock"
	"knative.dev/pkg/ptr"
	apiserving "knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingtest "knative.dev/serving/pkg/testing/v1"
	"sigs.k8s.io/yaml"
)

var revisionSpec = servingv1.RevisionSpec{
	PodSpec: v1.PodSpec{
		Containers: []v1.Container{{
			Image: "busybox",
		}},
		EnableServiceLinks: ptr.Bool(false),
	},
	TimeoutSeconds: ptr.Int64(300),
}

type testCase struct {
	name             string
	latestSvc        *servingv1.Service
	expectedSvcList  *servingv1.ServiceList
	revisionList     *servingv1.RevisionList
	expectedKNExport *clientv1alpha1.Export
}

func TestServiceExportError(t *testing.T) {
	tc := &testCase{latestSvc: getService("foo")}

	_, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name)
	assert.Error(t, err, "'kn service export' requires output format")
}

func TestServiceExport(t *testing.T) {

	for _, tc := range []testCase{
		{latestSvc: libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()))},
		{latestSvc: libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()), servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "mouse"}))},
		{latestSvc: libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()), libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v2"}))},
		{latestSvc: libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()), servingtest.WithServiceLabel("a", "mouse"), servingtest.WithServiceAnnotation("a", "mouse"))},
		{latestSvc: libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()), servingtest.WithVolume("secretName", "/mountpath", volumeSource("secretName")))},
	} {
		exportServiceTestForReplay(t, &tc)
		tc.expectedKNExport = libtest.BuildKNExportWithOptions()
		exportServiceTest(t, &tc, true)
		//test default
		exportServiceTest(t, &tc, false)
	}
}

func exportServiceTestForReplay(t *testing.T, tc *testCase) {
	output, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--mode", "replay", "-o", "yaml")
	assert.NilError(t, err)

	actSvc := servingv1.Service{}
	err = yaml.Unmarshal([]byte(output), &actSvc)
	assert.NilError(t, err)

	assert.DeepEqual(t, tc.latestSvc, &actSvc)
}

func exportServiceTest(t *testing.T, tc *testCase, addMode bool) {
	args := []string{"export", tc.latestSvc.ObjectMeta.Name, "-o", "json"}
	if addMode {
		args = append(args, []string{"--mode", "export"}...)
	}
	output, err := executeServiceExportCommand(t, tc, args...)
	assert.NilError(t, err)

	tc.expectedKNExport.Spec.Service = *tc.latestSvc

	actKNExport := &clientv1alpha1.Export{}
	err = json.Unmarshal([]byte(output), actKNExport)
	assert.NilError(t, err)

	assert.DeepEqual(t, tc.expectedKNExport, actKNExport)
}

func TestServiceExportwithMultipleRevisions(t *testing.T) {
	for _, tc := range []testCase{{
		name: "test 2 revisions with traffic split",
		latestSvc: libtest.BuildServiceWithOptions(
			"foo", servingtest.WithConfigSpec(buildConfiguration()),
			libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v2"}),
			libtest.WithTrafficSpec([]string{"foo-rev-1", "latest"}, []int{50, 50}, []string{"", ""}),
		),
		expectedSvcList: libtest.BuildServiceListWithOptions(
			libtest.WithService(libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()),
				libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v1"}),
				servingtest.WithBYORevisionName("foo-rev-1"),
			)),
			libtest.WithService(libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()),
				libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v2"}),
				libtest.WithTrafficSpec([]string{"foo-rev-1", "latest"}, []int{50, 50}, []string{"", ""}),
			)),
		),
		revisionList: libtest.BuildRevisionListWithOptions(
			libtest.WithRevision(*(libtest.BuildRevision("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionAnn("client.knative.dev/user-image", "busybox:v1"),
				servingtest.WithRevisionAnn("serving.knative.dev/lastPinned", "1111132"),
			))),
			libtest.WithRevision(*(libtest.BuildRevision("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionAnn("client.knative.dev/user-image", "busybox:v2"),
			))),
		),
		expectedKNExport: libtest.BuildKNExportWithOptions(
			libtest.WithKNRevision(*(libtest.BuildRevision("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionAnn("client.knative.dev/user-image", "busybox:v1"),
			))),
		),
	}, {
		name: "test 2 revisions no traffic split",
		latestSvc: libtest.BuildServiceWithOptions(
			"foo", servingtest.WithConfigSpec(buildConfiguration()),
			libtest.WithTrafficSpec([]string{"latest"}, []int{100}, []string{""}),
		),
		expectedSvcList: libtest.BuildServiceListWithOptions(
			libtest.WithService(libtest.BuildServiceWithOptions(
				"foo", servingtest.WithConfigSpec(buildConfiguration()),
				libtest.WithTrafficSpec([]string{"latest"}, []int{100}, []string{""}),
			)),
		),
		revisionList: libtest.BuildRevisionListWithOptions(
			libtest.WithRevision(*(libtest.BuildRevision("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
			))),
			libtest.WithRevision(*(libtest.BuildRevision("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "2"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
			))),
		),
		expectedKNExport: libtest.BuildKNExportWithOptions(),
	}, {
		name: "test 3 active revisions with traffic split with no latest revision",
		latestSvc: libtest.BuildServiceWithOptions(
			"foo", servingtest.WithConfigSpec(buildConfiguration()),
			servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "mouse"}),
			servingtest.WithBYORevisionName("foo-rev-3"),
			libtest.WithTrafficSpec([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []string{"", "", ""}),
		),
		expectedSvcList: libtest.BuildServiceListWithOptions(
			libtest.WithService(
				libtest.BuildServiceWithOptions(
					"foo", servingtest.WithConfigSpec(buildConfiguration()),
					servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "cat"}),
					servingtest.WithBYORevisionName("foo-rev-1"),
				),
			),
			libtest.WithService(
				libtest.BuildServiceWithOptions(
					"foo", servingtest.WithConfigSpec(buildConfiguration()),
					servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "dog"}),
					servingtest.WithBYORevisionName("foo-rev-2"),
				),
			),
			libtest.WithService(
				libtest.BuildServiceWithOptions(
					"foo", servingtest.WithConfigSpec(buildConfiguration()),
					servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "mouse"}),
					servingtest.WithBYORevisionName("foo-rev-3"),
					libtest.WithTrafficSpec([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []string{"", "", ""}),
				),
			),
		),
		revisionList: libtest.BuildRevisionListWithOptions(
			libtest.WithRevision(*(libtest.BuildRevision("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevisionEnv(v1.EnvVar{Name: "a", Value: "cat"}),
			))),
			libtest.WithRevision(*(libtest.BuildRevision("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "2"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevisionEnv(v1.EnvVar{Name: "a", Value: "dog"}),
			))),
			libtest.WithRevision(*(libtest.BuildRevision("foo-rev-3",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "3"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevisionEnv(v1.EnvVar{Name: "a", Value: "mouse"}),
			))),
		),
		expectedKNExport: libtest.BuildKNExportWithOptions(
			libtest.WithKNRevision(*(libtest.BuildRevision("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevisionEnv(v1.EnvVar{Name: "a", Value: "cat"}),
			))),
			libtest.WithKNRevision(*(libtest.BuildRevision("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "2"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevisionEnv(v1.EnvVar{Name: "a", Value: "dog"}),
			))),
		),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			exportWithRevisionsforKubernetesTest(t, &tc)
			exportWithRevisionsTest(t, &tc)
		})
	}
}

func exportWithRevisionsforKubernetesTest(t *testing.T, tc *testCase) {
	output, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "--mode", "replay", "-o", "json")
	assert.NilError(t, err)

	actSvcList := servingv1.ServiceList{}
	err = json.Unmarshal([]byte(output), &actSvcList)
	assert.NilError(t, err)
	assert.DeepEqual(t, tc.expectedSvcList, &actSvcList)
}

func exportWithRevisionsTest(t *testing.T, tc *testCase) {
	output, err := executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "--mode", "export", "-o", "json")
	assert.NilError(t, err)

	tc.expectedKNExport.Spec.Service = *tc.latestSvc

	actKNExport := &clientv1alpha1.Export{}
	err = json.Unmarshal([]byte(output), actKNExport)
	assert.NilError(t, err)

	assert.DeepEqual(t, tc.expectedKNExport, actKNExport)
}

func executeServiceExportCommand(t *testing.T, tc *testCase, options ...string) (string, error) {
	client := knclient.NewMockKnServiceClient(t)
	r := client.Recorder()

	r.GetService(tc.latestSvc.ObjectMeta.Name, tc.latestSvc, nil)
	r.ListRevisions(mock.Any(), tc.revisionList, nil)

	return executeServiceCommand(client, options...)
}

func buildConfiguration() *servingv1.ConfigurationSpec {
	c := &servingv1.Configuration{
		Spec: servingv1.ConfigurationSpec{
			Template: servingv1.RevisionTemplateSpec{
				Spec: *revisionSpec.DeepCopy(),
			},
		},
	}
	c.SetDefaults(context.Background())
	return &c.Spec
}

func volumeSource(secretName string) v1.VolumeSource {
	return v1.VolumeSource{
		Secret: &v1.SecretVolumeSource{
			SecretName: secretName,
		},
	}
}
