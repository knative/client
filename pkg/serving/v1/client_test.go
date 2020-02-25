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

package v1

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"knative.dev/serving/pkg/apis/serving"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"

	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"

	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"
)

var testNamespace = "test-ns"

func setup() (serving servingv1fake.FakeServingV1, client KnServingClient) {
	serving = servingv1fake.FakeServingV1{Fake: &clienttesting.Fake{}}
	client = NewKnServingClient(&serving, testNamespace)
	return
}

func TestGetService(t *testing.T) {
	serving, client := setup()
	serviceName := "test-service"

	serving.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			service := newService(serviceName)
			name := a.(clienttesting.GetAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == serviceName {
				return true, service, nil
			}
			return true, nil, errors.NewNotFound(servingv1.Resource("service"), name)
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
			func(a clienttesting.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &servingv1.ServiceList{Items: []servingv1.Service{*service1, *service2}}, nil
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
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(clienttesting.CreateAction).GetObject().(metav1.Object).GetName()
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
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(clienttesting.UpdateAction).GetObject().(metav1.Object).GetName()
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
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.DeleteAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == serviceName {
				return true, nil, nil
			}
			return true, nil, errors.NewNotFound(servingv1.Resource("service"), name)
		})
	serving.AddWatchReactor("services",
		func(a clienttesting.Action) (bool, watch.Interface, error) {
			watchAction := a.(clienttesting.WatchAction)
			name, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
			if !found {
				return true, nil, errors.NewNotFound(servingv1.Resource("service"), name)
			}
			w := wait.NewFakeWatch(getServiceDeleteEvents("test-service"))
			w.Start()
			return true, w, nil
		})

	t.Run("delete existing service returns no error", func(t *testing.T) {
		err := client.DeleteService(serviceName, time.Duration(10)*time.Second)
		assert.NilError(t, err)
	})

	t.Run("trying to delete non-existing service returns error", func(t *testing.T) {
		err := client.DeleteService(nonExistingServiceName, time.Duration(10)*time.Second)
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistingServiceName)
	})
}

func getServiceDeleteEvents(name string) []watch.Event {
	return []watch.Event{
		{watch.Added, wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", "msg1")},
		{watch.Modified, wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", "msg2")},
		{watch.Deleted, wait.CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "")},
	}
}

func TestGetRevision(t *testing.T) {
	serving, client := setup()

	const (
		revisionName            = "test-revision"
		notExistingRevisionName = "no-revision"
	)

	serving.AddReactor("get", "revisions",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			revision := newRevision(revisionName)
			name := a.(clienttesting.GetAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == revisionName {
				return true, revision, nil
			}
			return true, nil, errors.NewNotFound(servingv1.Resource("revision"), name)
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

	revisions := []servingv1.Revision{*newRevision("revision-1"), *newRevision("revision-2")}
	serving.AddReactor("list", "revisions",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			return true, &servingv1.RevisionList{Items: revisions}, nil
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
	fakeServing, client := setup()

	serviceName := "service"

	fakeServing.AddReactor("list", "revisions",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			revisions := []servingv1.Revision{
				*newRevision("revision-1", serving.ServiceLabelKey, "service"),
				*newRevision("revision-2"),
			}

			lAction := a.(clienttesting.ListAction)
			assert.Equal(t, testNamespace, a.GetNamespace())
			restrictions := lAction.GetListRestrictions()
			assert.Check(t, restrictions.Fields.Empty())
			servicesLabels := labels.Set{serving.ServiceLabelKey: serviceName}
			assert.Check(t, restrictions.Labels.Matches(servicesLabels))
			return true, &servingv1.RevisionList{Items: revisions}, nil
		})

	t.Run("list revisions for a service returns a list of revisions associated with this this service and no error",
		func(t *testing.T) {
			revisions, err := client.ListRevisions(WithService(serviceName))
			assert.NilError(t, err)

			assert.Assert(t, cmp.Len(revisions.Items, 1))
			assert.Equal(t, revisions.Items[0].Name, "revision-1")
			assert.Equal(t, revisions.Items[0].Labels[serving.ServiceLabelKey], "service")
			validateGroupVersionKind(t, revisions)
			validateGroupVersionKind(t, &revisions.Items[0])
		})
}

func TestGetRoute(t *testing.T) {
	serving, client := setup()
	routeName := "test-route"

	serving.AddReactor("get", "routes",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			route := newRoute(routeName)
			name := a.(clienttesting.GetAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == routeName {
				return true, route, nil
			}
			return true, nil, errors.NewNotFound(servingv1.Resource("route"), name)
		})

	t.Run("get known route by name returns route", func(t *testing.T) {
		route, err := client.GetRoute(routeName)
		assert.NilError(t, err)
		assert.Equal(t, routeName, route.Name, "route name should be equal")
		validateGroupVersionKind(t, route)
	})

	t.Run("get unknown route name returns error", func(t *testing.T) {
		nonExistingRouteName := "r@ute-that-d$es-n#t-exist"
		route, err := client.GetRoute(nonExistingRouteName)
		assert.Assert(t, route == nil, "no route should be returned")
		assert.ErrorContains(t, err, "not found")
		assert.ErrorContains(t, err, nonExistingRouteName)
	})
}

func TestListRoutes(t *testing.T) {
	serving, client := setup()

	singleRouteName := "route-2"
	singleRoute := *newRoute(singleRouteName)
	routes := []servingv1.Route{*newRoute("route-1"), singleRoute, *newRoute("route-3")}
	serving.AddReactor("list", "routes",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			lAction := a.(clienttesting.ListAction)
			restrictions := lAction.GetListRestrictions()
			assert.Assert(t, restrictions.Labels.Empty())
			if !restrictions.Fields.Empty() {
				nameField := fields.Set{"metadata.name": singleRouteName}
				assert.Check(t, restrictions.Labels.Matches(nameField))
				return true, &servingv1.RouteList{Items: []servingv1.Route{singleRoute}}, nil
			}
			return true, &servingv1.RouteList{Items: routes}, nil
		})

	t.Run("list routes returns a list of routes and no error", func(t *testing.T) {

		routes, err := client.ListRoutes()
		assert.NilError(t, err)

		assert.Assert(t, len(routes.Items) == 3)
		assert.Equal(t, routes.Items[0].Name, "route-1")
		assert.Equal(t, routes.Items[1].Name, singleRouteName)
		assert.Equal(t, routes.Items[2].Name, "route-3")
		validateGroupVersionKind(t, routes)
		for i := 0; i < len(routes.Items); i++ {
			validateGroupVersionKind(t, &routes.Items[i])
		}
	})

	t.Run("list routes with a name filter a list with one route and no error", func(t *testing.T) {

		routes, err := client.ListRoutes(WithName(singleRouteName))
		assert.NilError(t, err)

		assert.Assert(t, len(routes.Items) == 1)
		assert.Equal(t, routes.Items[0].Name, singleRouteName)
		validateGroupVersionKind(t, routes)
		validateGroupVersionKind(t, &routes.Items[0])
	})
}

func TestWaitForService(t *testing.T) {
	serving, client := setup()

	serviceName := "test-service"

	serving.AddWatchReactor("services",
		func(a clienttesting.Action) (bool, watch.Interface, error) {
			watchAction := a.(clienttesting.WatchAction)
			_, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
			if !found {
				return true, nil, fmt.Errorf("no field selector on metadata.name found")
			}
			w := wait.NewFakeWatch(getServiceEvents(serviceName))
			w.Start()
			return true, w, nil
		})
	serving.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			getAction := a.(clienttesting.GetAction)
			assert.Equal(t, getAction.GetName(), serviceName)
			return true, newService(serviceName), nil
		})

	t.Run("wait on a service to become ready with success", func(t *testing.T) {
		err, duration := client.WaitForService(serviceName, 60*time.Second, wait.NoopMessageCallback())
		assert.NilError(t, err)
		assert.Assert(t, duration > 0)
	})
}

type baseRevisionCase struct {
	templateName          string
	templateImage         string
	latestCreated         string
	requestedRevisionName string

	foundRevisionImage string
	errText            string
}

func TestGetBaseRevision(t *testing.T) {
	servingFake, client := setup()
	cases := []baseRevisionCase{
		{"foo-asdf", "gcr.io/foo/bar", "foo-asdf", "foo-asdf", "gcr.io/foo/bar", ""},
		{"", "gcr.io/foo/bar", "foo-asdf", "foo-asdf", "gcr.io/foo/bar", ""},
		{"foo-qwer", "gcr.io/foo/bar", "foo-asdf", "foo-qwer", "gcr.io/foo/bar", ""},
		{"", "gcr.io/foo/bar", "foo-asdf", "foo-asdf", "gcr.io/foo/baz", "base revision not found"},
	}
	var c baseRevisionCase
	servingFake.AddReactor("get", "revisions", func(a clienttesting.Action) (bool, runtime.Object, error) {
		revision := &servingv1.Revision{ObjectMeta: metav1.ObjectMeta{Name: c.requestedRevisionName, Namespace: testNamespace}}
		revision.Spec.Containers = []corev1.Container{{}}
		name := a.(clienttesting.GetAction).GetName()
		assert.Equal(t, name, c.requestedRevisionName)

		if c.foundRevisionImage != "" {
			revision.Spec.Containers[0].Image = c.foundRevisionImage
			return true, revision, nil
		}
		return true, nil, errors.NewNotFound(servingv1.Resource("revision"), name)
	})
	for _, c = range cases {
		service := servingv1.Service{}
		service.Spec.Template = servingv1.RevisionTemplateSpec{}
		service.Spec.Template.Name = c.templateName
		service.Status.LatestCreatedRevisionName = c.latestCreated
		service.Spec.Template.Spec.Containers = []corev1.Container{{}}
		service.Spec.Template.Spec.Containers[0].Image = c.templateImage

		r, err := client.GetBaseRevision(&service)
		if err == nil {
			assert.Equal(t, r.Spec.Containers[0].Image, c.foundRevisionImage)
		} else {
			assert.Assert(t, c.errText != "")
			assert.ErrorContains(t, err, c.errText)
		}
	}
}

func TestGetConfiguration(t *testing.T) {
	serving, client := setup()

	const (
		configName                   = "test-config"
		notExistingConfigurationName = "no-revision"
	)

	serving.AddReactor("get", "configurations",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			configuration := &servingv1.Configuration{ObjectMeta: metav1.ObjectMeta{Name: configName, Namespace: testNamespace}}
			name := a.(clienttesting.GetAction).GetName()
			// Sanity check
			assert.Assert(t, name != "")
			assert.Equal(t, testNamespace, a.GetNamespace())
			if name == configName {
				return true, configuration, nil
			}
			return true, nil, errors.NewNotFound(servingv1.Resource("configuration"), name)
		})

	t.Run("getting existing configuration returns configuration and no error", func(t *testing.T) {
		configuration, err := client.GetConfiguration(configName)
		assert.NilError(t, err)
		assert.Equal(t, configName, configuration.Name)
		validateGroupVersionKind(t, configuration)
	})

	t.Run("trying to get a configuration with a name that does not exist returns an error", func(t *testing.T) {
		configuration, err := client.GetConfiguration(notExistingConfigurationName)
		assert.Assert(t, configuration == nil)
		assert.ErrorContains(t, err, notExistingConfigurationName)
		assert.ErrorContains(t, err, "not found")
	})
}

func validateGroupVersionKind(t *testing.T, obj runtime.Object) {
	gvkExpected, err := util.GetGroupVersionKind(obj, servingv1.SchemeGroupVersion, scheme.Scheme)
	assert.NilError(t, err)
	gvkGiven := obj.GetObjectKind().GroupVersionKind()
	assert.Equal(t, *gvkExpected, gvkGiven, "GVK should be the same")
}

func newService(name string) *servingv1.Service {
	return &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace}}
}

func newRevision(name string, labels ...string) *servingv1.Revision {
	rev := &servingv1.Revision{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace}}
	labelMap := make(map[string]string)
	if len(labels) > 0 {
		for i := 0; i < len(labels); i += 2 {
			labelMap[labels[i]] = labels[i+1]
		}
		rev.Labels = labelMap
	}
	return rev
}

func newRoute(name string) *servingv1.Route {
	return &servingv1.Route{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace}}
}

func getServiceEvents(name string) []watch.Event {
	return []watch.Event{
		{watch.Added, wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", "msg1")},
		{watch.Modified, wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", "msg2")},
		{watch.Modified, wait.CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "")},
	}
}
