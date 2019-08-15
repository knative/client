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
	"time"

	"github.com/knative/pkg/apis"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/knative/client/pkg/serving"
	"github.com/knative/client/pkg/wait"

	kn_errors "github.com/knative/client/pkg/errors"
	api_serving "github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	client_v1alpha1 "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Kn interface to serving. All methods are relative to the
// namespace specified during construction
type KnClient interface {

	// Get a service by its unique name
	GetService(name string) (*v1alpha1.Service, error)

	// List services
	ListServices(opts ...ListConfig) (*v1alpha1.ServiceList, error)

	// Create a new service
	CreateService(service *v1alpha1.Service) error

	// Update the given service
	UpdateService(service *v1alpha1.Service) error

	// Delete a service by name
	DeleteService(name string) error

	// Wait for a service to become ready, but not longer than provided timeout
	WaitForService(name string, timeout time.Duration) error

	// Get a configuration by name
	GetConfiguration(name string) (*v1alpha1.Configuration, error)

	// Get a revision by name
	GetRevision(name string) (*v1alpha1.Revision, error)

	// List revisions
	ListRevisions(opts ...ListConfig) (*v1alpha1.RevisionList, error)

	// Delete a revision
	DeleteRevision(name string) error

	// Get a route by its unique name
	GetRoute(name string) (*v1alpha1.Route, error)

	// List routes
	ListRoutes(opts ...ListConfig) (*v1alpha1.RouteList, error)
}

type listConfigCollector struct {
	// Labels to filter on
	Labels labels.Set

	// Labels to filter on
	Fields fields.Set
}

// Config function for builder pattern
type ListConfig func(config *listConfigCollector)

type ListConfigs []ListConfig

// add selectors to a list options
func (opts ListConfigs) toListOptions() v1.ListOptions {
	listConfig := listConfigCollector{labels.Set{}, fields.Set{}}
	for _, f := range opts {
		f(&listConfig)
	}
	options := v1.ListOptions{}
	if len(listConfig.Fields) > 0 {
		options.FieldSelector = listConfig.Fields.String()
	}
	if len(listConfig.Labels) > 0 {
		options.LabelSelector = listConfig.Labels.String()
	}
	return options
}

// Filter list on the provided name
func WithName(name string) ListConfig {
	return func(lo *listConfigCollector) {
		lo.Fields["metadata.name"] = name
	}
}

// Filter on the service name
func WithService(service string) ListConfig {
	return func(lo *listConfigCollector) {
		lo.Labels[api_serving.ServiceLabelKey] = service
	}
}

type knClient struct {
	client    client_v1alpha1.ServingV1alpha1Interface
	namespace string
}

// Create a new client facade for the provided namespace
func NewKnServingClient(client client_v1alpha1.ServingV1alpha1Interface, namespace string) KnClient {
	return &knClient{
		client:    client,
		namespace: namespace,
	}
}

// Get a service by its unique name
func (cl *knClient) GetService(name string) (*v1alpha1.Service, error) {
	service, err := cl.client.Services(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	err = serving.UpdateGroupVersionKind(service, v1alpha1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// List services
func (cl *knClient) ListServices(config ...ListConfig) (*v1alpha1.ServiceList, error) {
	serviceList, err := cl.client.Services(cl.namespace).List(ListConfigs(config).toListOptions())
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	serviceListNew := serviceList.DeepCopy()
	err = updateServingGvk(serviceListNew)
	if err != nil {
		return nil, err
	}

	serviceListNew.Items = make([]v1alpha1.Service, len(serviceList.Items))
	for idx, service := range serviceList.Items {
		serviceClone := service.DeepCopy()
		err := updateServingGvk(serviceClone)
		if err != nil {
			return nil, err
		}
		serviceListNew.Items[idx] = *serviceClone
	}
	return serviceListNew, nil
}

// Create a new service
func (cl *knClient) CreateService(service *v1alpha1.Service) error {
	_, err := cl.client.Services(cl.namespace).Create(service)
	if err != nil {
		return kn_errors.GetError(err)
	}
	return updateServingGvk(service)
}

// Update the given service
func (cl *knClient) UpdateService(service *v1alpha1.Service) error {
	_, err := cl.client.Services(cl.namespace).Update(service)
	if err != nil {
		return err
	}
	return updateServingGvk(service)
}

// Delete a service by name
func (cl *knClient) DeleteService(serviceName string) error {
	err := cl.client.Services(cl.namespace).Delete(
		serviceName,
		&v1.DeleteOptions{},
	)
	if err != nil {
		return kn_errors.GetError(err)
	}

	return nil
}

// Wait for a service to become ready, but not longer than provided timeout
func (cl *knClient) WaitForService(name string, timeout time.Duration) error {
	waitForReady := newServiceWaitForReady(cl.client.Services(cl.namespace).Watch)
	return waitForReady.Wait(name, timeout)
}

// Get the configuration for a service
func (cl *knClient) GetConfiguration(name string) (*v1alpha1.Configuration, error) {
	configuration, err := cl.client.Configurations(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = updateServingGvk(configuration)
	if err != nil {
		return nil, err
	}
	return configuration, nil
}

// Get a revision by name
func (cl *knClient) GetRevision(name string) (*v1alpha1.Revision, error) {
	revision, err := cl.client.Revisions(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	err = updateServingGvk(revision)
	if err != nil {
		return nil, err
	}
	return revision, nil
}

// Delete a revision by name
func (cl *knClient) DeleteRevision(name string) error {
	err := cl.client.Revisions(cl.namespace).Delete(name, &v1.DeleteOptions{})
	if err != nil {
		return kn_errors.GetError(err)
	}

	return nil
}

// List revisions
func (cl *knClient) ListRevisions(config ...ListConfig) (*v1alpha1.RevisionList, error) {
	revisionList, err := cl.client.Revisions(cl.namespace).List(ListConfigs(config).toListOptions())
	if err != nil {
		return nil, kn_errors.GetError(err)
	}
	return updateServingGvkForRevisionList(revisionList)
}

// Get a route by its unique name
func (cl *knClient) GetRoute(name string) (*v1alpha1.Route, error) {
	route, err := cl.client.Routes(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = serving.UpdateGroupVersionKind(route, v1alpha1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// List routes
func (cl *knClient) ListRoutes(config ...ListConfig) (*v1alpha1.RouteList, error) {
	routeList, err := cl.client.Routes(cl.namespace).List(ListConfigs(config).toListOptions())
	if err != nil {
		return nil, err
	}
	return updateServingGvkForRouteList(routeList)
}

// update all the list + all items contained in the list with
// the proper GroupVersionKind specific to Knative serving
func updateServingGvkForRevisionList(revisionList *v1alpha1.RevisionList) (*v1alpha1.RevisionList, error) {
	revisionListNew := revisionList.DeepCopy()
	err := updateServingGvk(revisionListNew)
	if err != nil {
		return nil, err
	}

	revisionListNew.Items = make([]v1alpha1.Revision, len(revisionList.Items))
	for idx := range revisionList.Items {
		revision := revisionList.Items[idx].DeepCopy()
		err := updateServingGvk(revision)
		if err != nil {
			return nil, err
		}
		revisionListNew.Items[idx] = *revision
	}
	return revisionListNew, nil
}

// update all the list + all items contained in the list with
// the proper GroupVersionKind specific to Knative serving
func updateServingGvkForRouteList(routeList *v1alpha1.RouteList) (*v1alpha1.RouteList, error) {
	routeListNew := routeList.DeepCopy()
	err := updateServingGvk(routeListNew)
	if err != nil {
		return nil, err
	}

	routeListNew.Items = make([]v1alpha1.Route, len(routeList.Items))
	for idx := range routeList.Items {
		revision := routeList.Items[idx].DeepCopy()
		err := updateServingGvk(revision)
		if err != nil {
			return nil, err
		}
		routeListNew.Items[idx] = *revision
	}
	return routeListNew, nil
}

// update with the v1alpha1 group + version
func updateServingGvk(obj runtime.Object) error {
	return serving.UpdateGroupVersionKind(obj, v1alpha1.SchemeGroupVersion)
}

// Create wait arguments for a Knative service which can be used to wait for
// a create/update options to be finished
// Can be used by `service_create` and `service_update`, hence this extra file
func newServiceWaitForReady(watch wait.WatchFunc) wait.WaitForReady {
	return wait.NewWaitForReady(
		"service",
		watch,
		serviceConditionExtractor)
}

func serviceConditionExtractor(obj runtime.Object) (apis.Conditions, error) {
	service, ok := obj.(*v1alpha1.Service)
	if !ok {
		return nil, fmt.Errorf("%v is not a service", obj)
	}
	return apis.Conditions(service.Status.Conditions), nil
}
