// Copyright © 2021 The Knative Authors
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
	"context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"

	"knative.dev/client/pkg/util"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"
	servingv1alpha1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
)

const (
	testNamespace         = "test-ns"
	domainMappingResource = "domainmappings"
)

func setup() (serving servingv1alpha1fake.FakeServingV1alpha1, client KnServingClient) {
	serving = servingv1alpha1fake.FakeServingV1alpha1{Fake: &clienttesting.Fake{}}
	client = NewKnServingClient(&serving, testNamespace)
	return
}

func TestGetDomainMapping(t *testing.T) {
	serving, client := setup()
	serviceName := "foo"
	domainName := "foo.bar"

	serving.AddReactor("get", domainMappingResource,
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			dm := createDomainMapping(domainName, createServiceRef(serviceName, testNamespace))
			name := a.(clienttesting.GetAction).GetName()

			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == domainName {
				return true, dm, nil
			}
			return true, nil, errors.NewNotFound(servingv1alpha1.Resource("dm"), name)
		})

	t.Run("get domain mapping by name returns object", func(t *testing.T) {
		domainMapping, err := client.GetDomainMapping(context.Background(), domainName)
		assert.NilError(t, err)
		assert.Equal(t, domainName, domainMapping.Name, "domain mapping name should be equal")
		validateGroupVersionKind(t, domainMapping)
	})

	t.Run("get non-existing domain mapping by name returns error", func(t *testing.T) {
		nonExistingName := "does-not-exist"
		service, err := client.GetDomainMapping(context.Background(), nonExistingName)
		assert.Assert(t, service == nil, "no domain mapping should be returned")
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistingName)
	})
}

func TestCreateDomainMapping(t *testing.T) {
	serving, client := setup()
	serviceName := "foo"
	domainName := "foo.bar"
	domainMapping := createDomainMapping(domainName, createServiceRef(serviceName, testNamespace))
	serving.AddReactor("create", domainMappingResource,
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(clienttesting.CreateAction).GetObject().(metav1.Object).GetName()
			if name == domainMapping.Name {
				domainMapping.Generation = 2
				return true, domainMapping, nil
			}
			return true, nil, fmt.Errorf("error while creating service %s", name)
		})

	t.Run("create domain mapping without error creates a new object", func(t *testing.T) {
		err := client.CreateDomainMapping(context.Background(), domainMapping)
		assert.NilError(t, err)
		assert.Equal(t, domainMapping.Generation, int64(2))
		validateGroupVersionKind(t, domainMapping)
	})

	t.Run("create  domain mapping with an error returns an error object", func(t *testing.T) {
		err := client.CreateDomainMapping(context.Background(), createDomainMapping("unknown", createServiceRef(serviceName, testNamespace)))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestUpdateDomainMapping(t *testing.T) {
	serving, client := setup()
	serviceName := "foo"
	domainName := "foo.bar"
	domainMappingUpdate := createDomainMapping(domainName, createServiceRef(serviceName, testNamespace))
	domainMappingUpdate.ObjectMeta.Generation = 2

	serving.AddReactor("update", domainMappingResource,
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(clienttesting.UpdateAction).GetObject().(metav1.Object).GetName()
			if name == domainMappingUpdate.Name {
				dmResult := createDomainMapping(domainName, createServiceRef(serviceName, testNamespace))
				dmResult.Generation = 3
				return true, dmResult, nil
			}
			return true, nil, fmt.Errorf("error while updating service %s", name)
		})

	t.Run("update domain mapping without error", func(t *testing.T) {
		err := client.UpdateDomainMapping(context.Background(), domainMappingUpdate)
		assert.NilError(t, err)
		validateGroupVersionKind(t, domainMappingUpdate)
	})

	t.Run("update domain mapping with error", func(t *testing.T) {
		err := client.UpdateDomainMapping(context.Background(), createDomainMapping("unknown", createServiceRef(serviceName, testNamespace)))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestDeleteDomainMapping(t *testing.T) {
	serving, client := setup()
	domainName := "foo.bar"

	serving.AddReactor("delete", domainMappingResource,
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.DeleteAction).GetName()
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == domainName {
				return true, nil, nil
			}
			return true, nil, errors.NewNotFound(servingv1alpha1.Resource(domainMappingResource), name)
		})

	t.Run("delete domain mapping returns no error", func(t *testing.T) {
		err := client.DeleteDomainMapping(context.Background(), domainName)
		assert.NilError(t, err)
	})

	t.Run("delete non-existing domain mapping returns error", func(t *testing.T) {
		nonExistingName := "does-not-exist"
		err := client.DeleteDomainMapping(context.Background(), nonExistingName)
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistingName)
		assert.ErrorType(t, err, &errors.StatusError{})
	})
}

func TestListDomainMappings(t *testing.T) {
	serving, client := setup()
	t.Run("list domain mappings returns a list of objects", func(t *testing.T) {
		dm1 := createDomainMapping("dm-1", createServiceRef("svc1", testNamespace))
		dm2 := createDomainMapping("dm-2", createServiceRef("svc2", testNamespace))
		dm3 := createDomainMapping("dm-3", createServiceRef("svc3", testNamespace))
		serving.AddReactor("list", domainMappingResource,
			func(a clienttesting.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &servingv1alpha1.DomainMappingList{Items: []servingv1alpha1.DomainMapping{*dm1, *dm2, *dm3}}, nil
			})
		listServices, err := client.ListDomainMappings(context.Background())
		assert.NilError(t, err)
		assert.Assert(t, len(listServices.Items) == 3)
		assert.Equal(t, listServices.Items[0].Name, "dm-1")
		assert.Equal(t, listServices.Items[1].Name, "dm-2")
		assert.Equal(t, listServices.Items[2].Name, "dm-3")
		validateGroupVersionKind(t, listServices)
		validateGroupVersionKind(t, &listServices.Items[0])
		validateGroupVersionKind(t, &listServices.Items[1])
		validateGroupVersionKind(t, &listServices.Items[2])
	})
}

func validateGroupVersionKind(t *testing.T, obj runtime.Object) {
	gvkExpected, err := util.GetGroupVersionKind(obj, servingv1alpha1.SchemeGroupVersion, scheme.Scheme)
	assert.NilError(t, err)
	gvkGiven := obj.GetObjectKind().GroupVersionKind()
	fmt.Println(gvkGiven.String())
	assert.Equal(t, *gvkExpected, gvkGiven, "GVK should be the same")
}

func createDomainMapping(name string, ref duckv1.KReference) *servingv1alpha1.DomainMapping {
	return NewDomainMappingBuilder(name).Namespace("default").Reference(ref).Build()
}

func createServiceRef(service, namespace string) duckv1.KReference {
	return duckv1.KReference{Name: service,
		Kind:       "Service",
		APIVersion: "serving.knative.dev/v1",
		Namespace:  namespace,
	}
}
