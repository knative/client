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

package v1

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingtest "knative.dev/serving/pkg/testing/v1"

	libtest "knative.dev/client/lib/test"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestGitOpsOperations(t *testing.T) {
	c1TempDir, err := ioutil.TempDir("", "kn-files-cluster1")
	assert.NilError(t, err)
	c2TempDir, err := ioutil.TempDir("", "kn-files-cluster2")
	assert.NilError(t, err)
	defer os.RemoveAll(c1TempDir)
	defer os.RemoveAll(c2TempDir)
	// create clients
	fooclient := NewKnServingGitOpsClient("foo-ns", c1TempDir)
	bazclient := NewKnServingGitOpsClient("baz-ns", c1TempDir)
	globalclient := NewKnServingGitOpsClient("", c1TempDir)
	diffClusterClient := NewKnServingGitOpsClient("", "tmp")

	// set up test services
	fooSvc := libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()))
	barSvc := libtest.BuildServiceWithOptions("bar", servingtest.WithConfigSpec(buildConfiguration()))
	fooUpdateSvc := libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()), servingtest.WithEnv(corev1.EnvVar{Name: "a", Value: "mouse"}))

	fooserviceList := getServiceList([]servingv1.Service{*barSvc, *fooSvc})
	allServices := getServiceList([]servingv1.Service{*barSvc, *barSvc, *fooSvc})

	t.Run("get file path for foo service in foo namespace", func(t *testing.T) {
		fp := fooclient.(*knServingGitOpsClient).getKsvcFilePath("foo")
		assert.Equal(t, filepath.Join(c1TempDir, "foo-ns/ksvc/foo.yaml"), fp)
	})
	t.Run("get namespace for bazclient client", func(t *testing.T) {
		ns := bazclient.Namespace(context.Background())
		assert.Equal(t, "baz-ns", ns)
	})
	t.Run("create service foo in foo namespace", func(t *testing.T) {
		err := fooclient.CreateService(context.Background(), fooSvc)
		assert.NilError(t, err)
	})
	t.Run("wait for foo service in foo namespace", func(t *testing.T) {
		err, d := fooclient.WaitForService(context.Background(), "foo", 5*time.Second, nil)
		assert.NilError(t, err)
		assert.Equal(t, 1*time.Second, d)
	})
	t.Run("get service foo", func(t *testing.T) {
		result, err := fooclient.GetService(context.Background(), "foo")
		assert.NilError(t, err)
		assert.DeepEqual(t, fooSvc, result)
	})
	t.Run("create service bar in foo namespace", func(t *testing.T) {
		err := fooclient.CreateService(context.Background(), barSvc)
		assert.NilError(t, err)
	})
	t.Run("create service bar in baz namespace", func(t *testing.T) {
		err := bazclient.CreateService(context.Background(), barSvc)
		assert.NilError(t, err)
	})
	t.Run("list services in foo namespace", func(t *testing.T) {
		result, err := fooclient.ListServices(context.Background())
		assert.NilError(t, err)
		assert.DeepEqual(t, fooserviceList, result)
	})
	t.Run("create service without tmp directory", func(t *testing.T) {
		err := diffClusterClient.CreateService(context.Background(), fooSvc)
		assert.ErrorContains(t, err, "directory 'tmp' not present, please create the directory and try again")
	})
	diffClusterClient = NewKnServingGitOpsClient("", c2TempDir)
	t.Run("create service foo in foo namespace in cluster 2", func(t *testing.T) {
		err := diffClusterClient.CreateService(context.Background(), fooSvc)
		assert.NilError(t, err)
	})
	t.Run("list services in all namespaces in cluster 1", func(t *testing.T) {
		result, err := globalclient.ListServices(context.Background())
		assert.NilError(t, err)
		assert.DeepEqual(t, allServices, result)
	})
	t.Run("update service with retry foo", func(t *testing.T) {
		changed, err := fooclient.UpdateServiceWithRetry(context.Background(), "foo", func(svc *servingv1.Service) (*servingv1.Service, error) {
			return svc, nil
		}, 1)
		assert.Assert(t, changed)
		assert.NilError(t, err)
	})
	t.Run("update service foo", func(t *testing.T) {
		changed, err := fooclient.UpdateService(context.Background(), fooUpdateSvc)
		assert.Assert(t, changed)
		assert.NilError(t, err)
	})
	t.Run("check updated service foo", func(t *testing.T) {
		result, err := fooclient.GetService(context.Background(), "foo")
		assert.NilError(t, err)
		assert.DeepEqual(t, fooUpdateSvc, result)
	})
	t.Run("delete service foo", func(t *testing.T) {
		err := fooclient.DeleteService(context.Background(), "foo", 5*time.Second)
		assert.NilError(t, err)
	})
	t.Run("get service foo", func(t *testing.T) {
		_, err := fooclient.GetService(context.Background(), "foo")
		assert.ErrorType(t, err, apierrors.IsNotFound)
	})
}

func TestGitOpsSingleFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "singlefile")
	assert.NilError(t, err)
	defer os.RemoveAll(tmpDir)
	// create clients
	fooclient := NewKnServingGitOpsClient("", filepath.Join(tmpDir, "test.yaml"))
	barclient := NewKnServingGitOpsClient("", filepath.Join(tmpDir, "test.yml"))
	bazclient := NewKnServingGitOpsClient("", filepath.Join(tmpDir, "test.json"))

	// set up test services
	testSvc := libtest.BuildServiceWithOptions("test", servingtest.WithConfigSpec(buildConfiguration()))
	updateSvc := libtest.BuildServiceWithOptions("test", servingtest.WithConfigSpec(buildConfiguration()), servingtest.WithEnv(corev1.EnvVar{Name: "a", Value: "mouse"}))

	svcList := getServiceList([]servingv1.Service{*updateSvc})

	t.Run("get file path for fooclient", func(t *testing.T) {
		fp := fooclient.(*knServingGitOpsClient).getKsvcFilePath("test")
		assert.Equal(t, filepath.Join(tmpDir, "test.yaml"), fp)
	})
	t.Run("get namespace for fooclient", func(t *testing.T) {
		ns := fooclient.Namespace(context.Background())
		assert.Equal(t, "", ns)
	})
	t.Run("create service in single file mode in different formats", func(t *testing.T) {
		err := fooclient.CreateService(context.Background(), testSvc)
		assert.NilError(t, err)

		err = barclient.CreateService(context.Background(), testSvc)
		assert.NilError(t, err)

		err = bazclient.CreateService(context.Background(), testSvc)
		assert.NilError(t, err)
	})
	t.Run("retrieve services", func(t *testing.T) {
		result, err := fooclient.GetService(context.Background(), "test")
		assert.NilError(t, err)
		assert.DeepEqual(t, testSvc, result)

		result, err = barclient.GetService(context.Background(), "test")
		assert.NilError(t, err)
		assert.DeepEqual(t, testSvc, result)

		result, err = bazclient.GetService(context.Background(), "test")
		assert.NilError(t, err)
		assert.DeepEqual(t, testSvc, result)
	})
	t.Run("update service foo", func(t *testing.T) {
		changed, err := fooclient.UpdateService(context.Background(), updateSvc)
		assert.NilError(t, err)
		assert.Assert(t, changed)

		changed, err = barclient.UpdateService(context.Background(), updateSvc)
		assert.NilError(t, err)
		assert.Assert(t, changed)

		changed, err = bazclient.UpdateService(context.Background(), updateSvc)
		assert.NilError(t, err)
		assert.Assert(t, changed)
	})
	t.Run("list services", func(t *testing.T) {
		result, err := fooclient.ListServices(context.Background())
		assert.NilError(t, err)
		assert.DeepEqual(t, svcList, result)

		result, err = barclient.ListServices(context.Background())
		assert.NilError(t, err)
		assert.DeepEqual(t, svcList, result)

		result, err = bazclient.ListServices(context.Background())
		assert.NilError(t, err)
		assert.DeepEqual(t, svcList, result)
	})
	t.Run("delete service foo", func(t *testing.T) {
		err := fooclient.DeleteService(context.Background(), "test", 5*time.Second)
		assert.NilError(t, err)

		err = barclient.DeleteService(context.Background(), "test", 5*time.Second)
		assert.NilError(t, err)

		err = bazclient.DeleteService(context.Background(), "test", 5*time.Second)
		assert.NilError(t, err)
	})
	t.Run("get service foo", func(t *testing.T) {
		_, err := fooclient.GetService(context.Background(), "test")
		assert.ErrorType(t, err, apierrors.IsNotFound)

		_, err = barclient.GetService(context.Background(), "test")
		assert.ErrorType(t, err, apierrors.IsNotFound)

		_, err = bazclient.GetService(context.Background(), "test")
		assert.ErrorType(t, err, apierrors.IsNotFound)
	})
}

func getServiceList(services []servingv1.Service) *servingv1.ServiceList {
	return &servingv1.ServiceList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
		Items: services,
	}
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

var revisionSpec = servingv1.RevisionSpec{
	PodSpec: corev1.PodSpec{
		Containers: []corev1.Container{{
			Image: "busybox",
		}},
		EnableServiceLinks: ptr.Bool(false),
	},
	TimeoutSeconds: ptr.Int64(300),
}
