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

	"gotest.tools/assert"

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

	_, err = executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "-o", "json")
	assert.Error(t, err, "'kn service export --with-revisions' requires a mode, please specify one of replay|export")

	_, err = executeServiceExportCommand(t, tc, "export", tc.latestSvc.ObjectMeta.Name, "--with-revisions", "--mode", "k8s", "-o", "yaml")
	assert.Error(t, err, "'kn service export --with-revisions' requires a mode, please specify one of replay|export")
}

func TestServiceExport(t *testing.T) {

	for _, tc := range []testCase{
		{latestSvc: libtest.GetSvcWithOptions("foo", servingtest.WithConfigSpec(cfg()))},
		{latestSvc: libtest.GetSvcWithOptions("foo", servingtest.WithConfigSpec(cfg()), servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "mouse"}))},
		{latestSvc: libtest.GetSvcWithOptions("foo", servingtest.WithConfigSpec(cfg()), libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v2"}))},
		{latestSvc: libtest.GetSvcWithOptions("foo", servingtest.WithConfigSpec(cfg()), servingtest.WithServiceLabel("a", "mouse"), servingtest.WithServiceAnnotation("a", "mouse"))},
		{latestSvc: libtest.GetSvcWithOptions("foo", servingtest.WithConfigSpec(cfg()), servingtest.WithVolume("secretName", "/mountpath", volumeSource("secretName")))},
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

	assert.DeepEqual(t, tc.latestSvc, &actSvc)
}

func TestServiceExportwithMultipleRevisions(t *testing.T) {
	for _, tc := range []testCase{{
		name: "test 2 revisions with traffic split",
		latestSvc: libtest.GetSvcWithOptions(
			"foo", servingtest.WithConfigSpec(cfg()),
			libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v2"}),
			libtest.WithTrafficSplit([]string{"foo-rev-1", "latest"}, []int{50, 50}, []string{"", ""}),
		),
		expectedSvcList: libtest.GetServiceListWithOptions(
			libtest.WithServices(libtest.GetSvcWithOptions("foo", servingtest.WithConfigSpec(cfg()),
				libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v1"}),
				servingtest.WithBYORevisionName("foo-rev-1"),
			)),
			libtest.WithServices(libtest.GetSvcWithOptions("foo", servingtest.WithConfigSpec(cfg()),
				libtest.WithRevisionAnnotations(map[string]string{"client.knative.dev/user-image": "busybox:v2"}),
				libtest.WithTrafficSplit([]string{"foo-rev-1", "latest"}, []int{50, 50}, []string{"", ""}),
			)),
		),
		revisionList: libtest.GetRevisionListWithOptions(
			libtest.WithRevs(libtest.GetRev("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionAnn("client.knative.dev/user-image", "busybox:v1"),
				servingtest.WithRevisionAnn("serving.knative.dev/lastPinned", "1111132"),
			)),
			libtest.WithRevs(libtest.GetRev("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionAnn("client.knative.dev/user-image", "busybox:v2"),
			)),
		),
		expectedKNExport: libtest.GetKNExportWithOptions(
			libtest.WithKNRevs(libtest.GetRev("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionAnn("client.knative.dev/user-image", "busybox:v1"),
			)),
		),
	}, {
		name: "test 2 revisions no traffic split",
		latestSvc: libtest.GetSvcWithOptions(
			"foo", servingtest.WithConfigSpec(cfg()),
			libtest.WithTrafficSplit([]string{"latest"}, []int{100}, []string{""}),
		),
		expectedSvcList: libtest.GetServiceListWithOptions(
			libtest.WithServices(libtest.GetSvcWithOptions(
				"foo", servingtest.WithConfigSpec(cfg()),
				libtest.WithTrafficSplit([]string{"latest"}, []int{100}, []string{""}),
			)),
		),
		revisionList: libtest.GetRevisionListWithOptions(
			libtest.WithRevs(libtest.GetRev("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
			)),
			libtest.WithRevs(libtest.GetRev("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "2"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
			)),
		),
		expectedKNExport: libtest.GetKNExportWithOptions(),
	}, {
		name: "test 3 active revisions with traffic split with no latest revision",
		latestSvc: libtest.GetSvcWithOptions(
			"foo", servingtest.WithConfigSpec(cfg()),
			servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "mouse"}),
			servingtest.WithBYORevisionName("foo-rev-3"),
			libtest.WithTrafficSplit([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []string{"", "", ""}),
		),
		expectedSvcList: libtest.GetServiceListWithOptions(
			libtest.WithServices(
				libtest.GetSvcWithOptions(
					"foo", servingtest.WithConfigSpec(cfg()),
					servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "cat"}),
					servingtest.WithBYORevisionName("foo-rev-1"),
				),
			),
			libtest.WithServices(
				libtest.GetSvcWithOptions(
					"foo", servingtest.WithConfigSpec(cfg()),
					servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "dog"}),
					servingtest.WithBYORevisionName("foo-rev-2"),
				),
			),
			libtest.WithServices(
				libtest.GetSvcWithOptions(
					"foo", servingtest.WithConfigSpec(cfg()),
					servingtest.WithEnv(v1.EnvVar{Name: "a", Value: "mouse"}),
					servingtest.WithBYORevisionName("foo-rev-3"),
					libtest.WithTrafficSplit([]string{"foo-rev-1", "foo-rev-2", "foo-rev-3"}, []int{25, 50, 25}, []string{"", "", ""}),
				),
			),
		),
		revisionList: libtest.GetRevisionListWithOptions(
			libtest.WithRevs(libtest.GetRev("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevEnv(v1.EnvVar{Name: "a", Value: "cat"}),
			)),
			libtest.WithRevs(libtest.GetRev("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "2"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevEnv(v1.EnvVar{Name: "a", Value: "dog"}),
			)),
			libtest.WithRevs(libtest.GetRev("foo-rev-3",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "3"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevEnv(v1.EnvVar{Name: "a", Value: "mouse"}),
			)),
		),
		expectedKNExport: libtest.GetKNExportWithOptions(
			libtest.WithKNRevs(libtest.GetRev("foo-rev-1",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "1"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevEnv(v1.EnvVar{Name: "a", Value: "cat"}),
			)),
			libtest.WithKNRevs(libtest.GetRev("foo-rev-2",
				servingtest.WithRevisionLabel(apiserving.ConfigurationGenerationLabelKey, "2"),
				servingtest.WithRevisionLabel(apiserving.ServiceLabelKey, "foo"),
				libtest.WithRevEnv(v1.EnvVar{Name: "a", Value: "dog"}),
			)),
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

func cfg() *servingv1.ConfigurationSpec {
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

var revisionSpec = servingv1.RevisionSpec{
	PodSpec: v1.PodSpec{
		Containers: []v1.Container{{
			Image: "busybox",
		}},
	},
	TimeoutSeconds: ptr.Int64(300),
}

func volumeSource(secretName string) v1.VolumeSource {
	return v1.VolumeSource{
		Secret: &v1.SecretVolumeSource{
			SecretName: secretName,
		},
	}
}
