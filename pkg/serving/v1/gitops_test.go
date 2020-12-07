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
	"os"
	"testing"
	"time"

	"gotest.tools/assert"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingtest "knative.dev/serving/pkg/testing/v1"

	libtest "knative.dev/client/lib/test"
	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestGitOpsOperations(t *testing.T) {
	// create clients
	fooclient := NewKnServingGitOpsClient("foo-ns", "./tmp/kn/files")
	bazclient := NewKnServingGitOpsClient("baz-ns", "./tmp/kn/files")
	globalclient := NewKnServingGitOpsClient("", "./tmp/kn/files")
	diffClusterClient := NewKnServingGitOpsClient("", "./tmp/kn/cluster2/files")

	// set up test services
	fooSvc := libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()))
	barSvc := libtest.BuildServiceWithOptions("bar", servingtest.WithConfigSpec(buildConfiguration()))
	fooUpdateSvc := libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()), servingtest.WithEnv(corev1.EnvVar{Name: "a", Value: "mouse"}))

	fooserviceList := getServiceList([]servingv1.Service{*barSvc, *fooSvc})
	allServices := getServiceList([]servingv1.Service{*barSvc, *barSvc, *fooSvc})

	t.Run("get file path for foo service in foo namespace", func(t *testing.T) {
		fp := fooclient.(*knServingGitOpsClient).getKsvcFilePath("foo")
		assert.Equal(t, "tmp/kn/files/foo-ns/ksvc/foo.yaml", fp)
	})
	t.Run("get namespace for bazclient client", func(t *testing.T) {
		ns := bazclient.Namespace()
		assert.Equal(t, "baz-ns", ns)
	})
	t.Run("create service foo in foo namespace", func(t *testing.T) {
		err := fooclient.CreateService(fooSvc)
		assert.NilError(t, err)
	})
	t.Run("wait for foo service in foo namespace", func(t *testing.T) {
		err, d := fooclient.WaitForService("foo", 5*time.Second, nil)
		assert.NilError(t, err)
		assert.Equal(t, 1*time.Second, d)
	})
	t.Run("get service foo", func(t *testing.T) {
		result, err := fooclient.GetService("foo")
		assert.NilError(t, err)
		assert.DeepEqual(t, fooSvc, result)
	})
	t.Run("create service bar in foo namespace", func(t *testing.T) {
		err := fooclient.CreateService(barSvc)
		assert.NilError(t, err)
	})
	t.Run("create service bar in baz namespace", func(t *testing.T) {
		err := bazclient.CreateService(barSvc)
		assert.NilError(t, err)
	})
	t.Run("list services in foo namespace", func(t *testing.T) {
		result, err := fooclient.ListServices()
		assert.NilError(t, err)
		assert.DeepEqual(t, fooserviceList, result)
	})
	t.Run("create service foo in foo namespace in cluster 2", func(t *testing.T) {
		err := diffClusterClient.CreateService(fooSvc)
		assert.NilError(t, err)
	})
	t.Run("list services in all namespaces in cluster 1", func(t *testing.T) {
		result, err := globalclient.ListServices()
		assert.NilError(t, err)
		assert.DeepEqual(t, allServices, result)
	})
	t.Run("update service with retry foo", func(t *testing.T) {
		err := fooclient.UpdateServiceWithRetry("foo", func(svc *servingv1.Service) (*servingv1.Service, error) {
			return svc, nil
		}, 1)
		assert.NilError(t, err)
	})
	t.Run("update service foo", func(t *testing.T) {
		err := fooclient.UpdateService(fooUpdateSvc)
		assert.NilError(t, err)
	})
	t.Run("check updated service foo", func(t *testing.T) {
		result, err := fooclient.GetService("foo")
		assert.NilError(t, err)
		assert.DeepEqual(t, fooUpdateSvc, result)
	})
	t.Run("delete service foo", func(t *testing.T) {
		err := fooclient.DeleteService("foo", 5*time.Second)
		assert.NilError(t, err)
	})
	t.Run("get service foo", func(t *testing.T) {
		_, err := fooclient.GetService("foo")
		assert.ErrorType(t, err, apierrors.IsNotFound)
	})
	//clean up
	os.RemoveAll("./tmp")
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
