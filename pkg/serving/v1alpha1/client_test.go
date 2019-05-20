// Copyright © 2019 The Knative Authors
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

package v1alpha1

import (
	"fmt"
	"testing"

	"gotest.tools/assert"

	serving_api "github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"

	"github.com/knative/client/pkg/serving"
)

var testNamespace = "test-ns"

func setup() (serving fake.FakeServingV1alpha1, client KnClient) {
	serving = fake.FakeServingV1alpha1{Fake: &client_testing.Fake{}}
	client = NewKnServingClient(&serving, testNamespace)
	return
}

func TestGetService(t *testing.T) {
	serving, client := setup()
	serviceName := "test-service"

	serving.AddReactor("get", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			service := newService(serviceName)
			name := a.(client_testing.GetAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == serviceName {
				return true, service, nil
			}
			return true, nil, errors.NewNotFound(v1alpha1.Resource("service"), name)
		})

	t.Run("get known service by name returns service", func(t *testing.T) {
		service, err := client.GetService(serviceName)
		assert.NilError(t, err)
		assert.Equal(t, serviceName, service.Name, "service name should be equal")
		validateGroupVersionKind(t, service)
	})

	t.Run("get unknown service name returns error", func(t *testing.T) {
		nonExistingServiceName := "service-that-does-not-exist"
		service, err := client.GetService(nonExistingServiceName)
		assert.Assert(t, service == nil, "no service should be returned")
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistingServiceName)
	})
}

func TestListService(t *testing.T) {
	serving, client := setup()

	t.Run("list service returns a list of services", func(t *testing.T) {
		service1 := newService("service-1")
		service2 := newService("service-2")

		serving.AddReactor("list", "services",
			func(a client_testing.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &v1alpha1.ServiceList{Items: []v1alpha1.Service{*service1, *service2}}, nil
			})

		listServices, err := client.ListServices()
		assert.NilError(t, err)
		assert.Assert(t, len(listServices.Items) == 2)
		assert.Equal(t, listServices.Items[0].Name, "service-1")
		assert.Equal(t, listServices.Items[1].Name, "service-2")
		validateGroupVersionKind(t, listServices)
		validateGroupVersionKind(t, &listServices.Items[0])
		validateGroupVersionKind(t, &listServices.Items[1])
	})
}

func TestCreateService(t *testing.T) {
	serving, client := setup()

	serviceNew := newService("new-service")
	serving.AddReactor("create", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
			if name == serviceNew.Name {
				serviceNew.Generation = 2
				return true, serviceNew, nil
			}
			return true, nil, fmt.Errorf("error while creating service %s", name)
		})

	t.Run("create service without error creates a new service", func(t *testing.T) {
		err := client.CreateService(serviceNew)
		assert.NilError(t, err)
		assert.Equal(t, serviceNew.Generation, int64(2))
		validateGroupVersionKind(t, serviceNew)
	})

	t.Run("create service with an error returns an error object", func(t *testing.T) {
		err := client.CreateService(newService("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestUpdateService(t *testing.T) {
	serving, client := setup()
	serviceUpdate := newService("update-service")

	serving.AddReactor("update", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.UpdateAction).GetObject().(metav1.Object).GetName()
			if name == serviceUpdate.Name {
				serviceUpdate.Generation = 3
				return true, serviceUpdate, nil
			}
			return true, nil, fmt.Errorf("error while updating service %s", name)
		})

	t.Run("updating a service without error", func(t *testing.T) {
		err := client.UpdateService(serviceUpdate)
		assert.NilError(t, err)
		assert.Equal(t, int64(3), serviceUpdate.Generation)
		validateGroupVersionKind(t, serviceUpdate)
	})

	t.Run("updating a service with error", func(t *testing.T) {
		err := client.UpdateService(newService("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestDeleteService(t *testing.T) {
	serving, client := setup()
	const (
		serviceName            = "test-service"
		nonExistingServiceName = "no-service"
	)

	serving.AddReactor("delete", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == serviceName {
				return true, nil, nil
			}
			return true, nil, errors.NewNotFound(v1alpha1.Resource("service"), name)
		})

	t.Run("delete existing service returns no error", func(t *testing.T) {
		err := client.DeleteService(serviceName)
		assert.NilError(t, err)
	})

	t.Run("trying to delete non-existing service returns error", func(t *testing.T) {
		err := client.DeleteService(nonExistingServiceName)
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistingServiceName)
	})
}

func TestGetRevision(t *testing.T) {
	serving, client := setup()

	const (
		revisionName            = "test-revision"
		notExistingRevisionName = "no-revision"
	)

	serving.AddReactor("get", "revisions",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			revision := newRevision(revisionName)
			name := a.(client_testing.GetAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == revisionName {
				return true, revision, nil
			}
			return true, nil, errors.NewNotFound(v1alpha1.Resource("revision"), name)
		})

	t.Run("get existing revision returns revision and no error", func(t *testing.T) {
		revision, err := client.GetRevision(revisionName)
		assert.NilError(t, err)
		assert.Equal(t, revisionName, revision.Name)
		validateGroupVersionKind(t, revision)
	})

	t.Run("trying to get a revision with a name that does not exist returns an error", func(t *testing.T) {
		revision, err := client.GetRevision(notExistingRevisionName)
		assert.Assert(t, revision == nil)
		assert.ErrorContains(t, err, notExistingRevisionName)
		assert.ErrorContains(t, err, "not found")
	})
}

func TestListRevisions(t *testing.T) {
	serving, client := setup()

	revisions := []v1alpha1.Revision{*newRevision("revision-1"), *newRevision("revision-2")}
	serving.AddReactor("list", "revisions",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			return true, &v1alpha1.RevisionList{Items: revisions}, nil
		})

	t.Run("list revisions returns a list of revisions and no error", func(t *testing.T) {

		revisions, err := client.ListRevisions()
		assert.NilError(t, err)

		assert.Assert(t, len(revisions.Items) == 2)
		assert.Equal(t, revisions.Items[0].Name, "revision-1")
		assert.Equal(t, revisions.Items[1].Name, "revision-2")
		validateGroupVersionKind(t, revisions)
		validateGroupVersionKind(t, &revisions.Items[0])
		validateGroupVersionKind(t, &revisions.Items[1])
	})
}

func TestListRevisionForService(t *testing.T) {
	serving, client := setup()

	serviceName := "service"

	serving.AddReactor("list", "revisions",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			revisions := []v1alpha1.Revision{
				*newRevision("revision-1", serving_api.ServiceLabelKey, "service"),
				*newRevision("revision-2"),
			}

			lAction := a.(client_testing.ListAction)
			assert.Equal(t, testNamespace, a.GetNamespace())
			restrictions := lAction.GetListRestrictions()
			assert.Assert(t, restrictions.Fields.Empty())
			servicesLabels := labels.Set{serving_api.ServiceLabelKey: serviceName}
			assert.Assert(t, restrictions.Labels.Matches(servicesLabels))
			return true, &v1alpha1.RevisionList{Items: revisions}, nil
		})

	t.Run("list revisions for a service returns a list of revisions associated with this this service and no error",
		func(t *testing.T) {
			revisions, err := client.ListRevisionsForService(serviceName)
			assert.NilError(t, err)

			assert.Equal(t, len(revisions.Items), 1)
			assert.Equal(t, revisions.Items[0].Name, "revision-1")
			assert.Equal(t, revisions.Items[0].Labels[serving_api.ServiceLabelKey], "service")
			validateGroupVersionKind(t, revisions)
			validateGroupVersionKind(t, &revisions.Items[0])
		})
}

func validateGroupVersionKind(t *testing.T, obj runtime.Object) {
	gvkExpected, err := serving.GetGroupVersionKind(obj, v1alpha1.SchemeGroupVersion)
	assert.NilError(t, err)
	gvkGiven := obj.GetObjectKind().GroupVersionKind()
	assert.Equal(t, *gvkExpected, gvkGiven, "GVK should be the same")
}

func newService(name string) *v1alpha1.Service {
	return &v1alpha1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace}}
}

func newRevision(name string, labels ...string) *v1alpha1.Revision {
	rev := &v1alpha1.Revision{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace}}
	labelMap := make(map[string]string)
	if len(labels) > 0 {
		for i := 0; i < len(labels); i += 2 {
			labelMap[labels[i]] = labels[i+1]
		}
		rev.Labels = labelMap
	}
	return rev
}
