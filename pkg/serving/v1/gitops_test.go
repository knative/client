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
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	servingtest "knative.dev/serving/pkg/testing/v1"

	"knative.dev/pkg/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	libtest "knative.dev/client/lib/test"
)

var ksvcStore = make(map[string]*servingv1.Service)

type fakeFileClient struct{}

// newFakeKnServingGitOpsClient returns an instance of the
// kn service gitops client for testing
func newFakeKnServingGitOpsClient(namespace, dir string) *knServingGitOpsClient {
	return &knServingGitOpsClient{
		dir:        dir,
		namespace:  namespace,
		fileClient: &fakeFileClient{},
	}
}

func (f *fakeFileClient) writeToFile(obj runtime.Object, filePath string) error {
	svc, isSvc := obj.(*servingv1.Service)
	if !isSvc {
		return fmt.Errorf("obj is not a knative service")
	}
	ksvcStore[filePath] = svc
	return nil
}

func (f *fakeFileClient) readFromFile(key, name string) (*servingv1.Service, error) {
	svc := ksvcStore[key]
	if svc == nil {
		return nil, apierrors.NewNotFound(servingv1.Resource("services"), name)
	}
	return svc, nil
}

func (f *fakeFileClient) removeFile(filePath string) error {
	delete(ksvcStore, filePath)
	return nil
}

func (f *fakeFileClient) listFiles(filepath string) ([]servingv1.Service, error) {
	var services []servingv1.Service

	for k, v := range ksvcStore {
		//only append for requested namespace
		if !strings.Contains(k, filepath) {
			continue
		}
		services = append(services, *v)
	}

	if len(services) == 0 {
		return nil, nil
	}

	sort.SliceStable(services, func(i, j int) bool {
		a := services[i]
		b := services[j]

		return a.ObjectMeta.Name < b.ObjectMeta.Name
	})

	return services, nil
}

func TestGitOpsOperations(t *testing.T) {
	// create clients
	fooclient := newFakeKnServingGitOpsClient("foo-ns", "/kn/files")
	bazclient := newFakeKnServingGitOpsClient("baz-ns", "/kn/files")
	globalclient := newFakeKnServingGitOpsClient("", "/kn/files")
	diffClusterClient := newFakeKnServingGitOpsClient("", "/kn/cluster2/files")

	// set up test services
	fooSvc := libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()))
	barSvc := libtest.BuildServiceWithOptions("bar", servingtest.WithConfigSpec(buildConfiguration()))
	fooUpdateSvc := libtest.BuildServiceWithOptions("foo", servingtest.WithConfigSpec(buildConfiguration()), servingtest.WithEnv(corev1.EnvVar{Name: "a", Value: "mouse"}))

	fooserviceList := getServiceList([]servingv1.Service{*barSvc, *fooSvc})
	allServices := getServiceList([]servingv1.Service{*barSvc, *barSvc, *fooSvc})

	t.Run("get file path for foo service in foo namespace", func(t *testing.T) {
		fp := fooclient.getKsvcFilePath("foo")
		assert.Equal(t, "/kn/files/foo-ns/ksvc/foo.yaml", fp)
	})
	t.Run("get namespace for bazclient client", func(t *testing.T) {
		ns := bazclient.Namespace()
		assert.Equal(t, "baz-ns", ns)
	})
	t.Run("create service foo in foo namespace", func(t *testing.T) {
		err := fooclient.CreateService(fooSvc)
		assert.NilError(t, err)
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
}

func TestForUnsupportedMessage(t *testing.T) {
	fooclient := newFakeKnServingGitOpsClient("foo-ns", "/kn/files")

	t.Run("WatchService", func(t *testing.T) {
		_, err := fooclient.WatchService("foo", 5*time.Second)
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("WatchRevision", func(t *testing.T) {
		_, err := fooclient.WatchRevision("foo", 5*time.Second)
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("ApplyService", func(t *testing.T) {
		_, err := fooclient.ApplyService(nil)
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("GetConfiguration", func(t *testing.T) {
		_, err := fooclient.GetConfiguration("foo")
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("GetRevision", func(t *testing.T) {
		_, err := fooclient.GetRevision("foo")
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("GetBaseRevision", func(t *testing.T) {
		_, err := fooclient.GetBaseRevision(nil)
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("CreateRevision", func(t *testing.T) {
		err := fooclient.CreateRevision(nil)
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("UpdateRevision", func(t *testing.T) {
		err := fooclient.UpdateRevision(nil)
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("DeleteRevision", func(t *testing.T) {
		err := fooclient.DeleteRevision("foo", 5*time.Second)
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("ListRevisions", func(t *testing.T) {
		_, err := fooclient.ListRevisions()
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("ListRoutes", func(t *testing.T) {
		_, err := fooclient.ListRoutes()
		assert.ErrorContains(t, err, operationNotSuportedError)
	})
	t.Run("GetRoute", func(t *testing.T) {
		_, err := fooclient.GetRoute("foo")
		assert.ErrorContains(t, err, operationNotSuportedError)
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
