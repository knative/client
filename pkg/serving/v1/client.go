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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"knative.dev/pkg/apis"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"

	"knative.dev/client/pkg/serving"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	apiserving "knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	clientv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"

	clienterrors "knative.dev/client/pkg/errors"
)

// Func signature for an updating function which returns the updated service object
// or an error
type serviceUpdateFunc func(origService *servingv1.Service) (*servingv1.Service, error)

// Kn interface to serving. All methods are relative to the
// namespace specified during construction
type KnServingClient interface {

	// Namespace in which this client is operating for
	Namespace() string

	// Get a service by its unique name
	GetService(name string) (*servingv1.Service, error)

	// List services
	ListServices(opts ...ListConfig) (*servingv1.ServiceList, error)

	// Create a new service
	CreateService(service *servingv1.Service) error

	// UpdateService updates the given service. For a more robust variant with automatic
	// conflict resolution see UpdateServiceWithRetry
	UpdateService(service *servingv1.Service) error

	// UpdateServiceWithRetry updates service and retries if there is a version conflict.
	// The updateFunc receives a deep copy of the existing service and can add update it in
	// place.
	UpdateServiceWithRetry(name string, updateFunc serviceUpdateFunc, nrRetries int) error

	// Delete a service by name
	DeleteService(name string, timeout time.Duration) error

	// Wait for a service to become ready, but not longer than provided timeout.
	// Return error and how long has been waited
	WaitForService(name string, timeout time.Duration, msgCallback wait.MessageCallback) (error, time.Duration)

	// Get a configuration by name
	GetConfiguration(name string) (*servingv1.Configuration, error)

	// Get a revision by name
	GetRevision(name string) (*servingv1.Revision, error)

	// Get the "base" revision for a Service; the one that corresponds to the
	// current template.
	GetBaseRevision(service *servingv1.Service) (*servingv1.Revision, error)

	// List revisions
	ListRevisions(opts ...ListConfig) (*servingv1.RevisionList, error)

	// Delete a revision
	DeleteRevision(name string, timeout time.Duration) error

	// Get a route by its unique name
	GetRoute(name string) (*servingv1.Route, error)

	// List routes
	ListRoutes(opts ...ListConfig) (*servingv1.RouteList, error)
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
		lo.Labels[apiserving.ServiceLabelKey] = service
	}
}

// WithLabel filters on the provided label
func WithLabel(labelKey, labelValue string) ListConfig {
	return func(lo *listConfigCollector) {
		lo.Labels[labelKey] = labelValue
	}
}

type knServingClient struct {
	client    clientv1.ServingV1Interface
	namespace string
}

// Create a new client facade for the provided namespace
func NewKnServingClient(client clientv1.ServingV1Interface, namespace string) KnServingClient {
	return &knServingClient{
		client:    client,
		namespace: namespace,
	}
}

// Return the client's namespace
func (cl *knServingClient) Namespace() string {
	return cl.namespace
}

// Get a service by its unique name
func (cl *knServingClient) GetService(name string) (*servingv1.Service, error) {
	service, err := cl.client.Services(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, clienterrors.GetError(err)
	}
	err = updateServingGvk(service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (cl *knServingClient) WatchService(name string, timeout time.Duration) (watch.Interface, error) {
	return wait.NewWatcher(cl.client.Services(cl.namespace).Watch,
		cl.client.RESTClient(), cl.namespace, "services", name, timeout)
}

func (cl *knServingClient) WatchRevision(name string, timeout time.Duration) (watch.Interface, error) {
	return wait.NewWatcher(cl.client.Revisions(cl.namespace).Watch,
		cl.client.RESTClient(), cl.namespace, "revision", name, timeout)
}

// List services
func (cl *knServingClient) ListServices(config ...ListConfig) (*servingv1.ServiceList, error) {
	serviceList, err := cl.client.Services(cl.namespace).List(ListConfigs(config).toListOptions())
	if err != nil {
		return nil, clienterrors.GetError(err)
	}
	serviceListNew := serviceList.DeepCopy()
	err = updateServingGvk(serviceListNew)
	if err != nil {
		return nil, err
	}

	serviceListNew.Items = make([]servingv1.Service, len(serviceList.Items))
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
func (cl *knServingClient) CreateService(service *servingv1.Service) error {
	_, err := cl.client.Services(cl.namespace).Create(service)
	if err != nil {
		return clienterrors.GetError(err)
	}
	return updateServingGvk(service)
}

// Update the given service
func (cl *knServingClient) UpdateService(service *servingv1.Service) error {
	_, err := cl.client.Services(cl.namespace).Update(service)
	if err != nil {
		return err
	}
	return updateServingGvk(service)
}

// Update the given service with a retry in case of a conflict
func (cl *knServingClient) UpdateServiceWithRetry(name string, updateFunc serviceUpdateFunc, nrRetries int) error {
	return updateServiceWithRetry(cl, name, updateFunc, nrRetries)
}

// Extracted to be usable with the Mocking client
func updateServiceWithRetry(cl KnServingClient, name string, updateFunc serviceUpdateFunc, nrRetries int) error {
	var retries = 0
	for {
		service, err := cl.GetService(name)
		if err != nil {
			return err
		}
		if service.GetDeletionTimestamp() != nil {
			return fmt.Errorf("can't update service %s because it has been marked for deletion", name)
		}
		updatedService, err := updateFunc(service.DeepCopy())
		if err != nil {
			return err
		}

		err = cl.UpdateService(updatedService)
		if err != nil {
			// Retry to update when a resource version conflict exists
			if apierrors.IsConflict(err) && retries < nrRetries {
				retries++
				// Wait a second before doing the retry
				time.Sleep(time.Second)
				continue
			}
			return fmt.Errorf("giving up after %d retries: %w", nrRetries, err)
		}
		return nil
	}
}

// Delete a service by name
// Param `timeout` represents a duration to wait for a delete op to finish.
// For `timeout == 0` delete is performed async without any wait.
func (cl *knServingClient) DeleteService(serviceName string, timeout time.Duration) error {
	if timeout == 0 {
		return cl.deleteService(serviceName, v1.DeletePropagationBackground)
	}
	waitC := make(chan error)
	go func() {
		waitForEvent := wait.NewWaitForEvent("service", cl.WatchService, func(evt *watch.Event) bool { return evt.Type == watch.Deleted })
		err, _ := waitForEvent.Wait(serviceName, wait.Options{Timeout: &timeout}, wait.NoopMessageCallback())
		waitC <- err
	}()
	err := cl.deleteService(serviceName, v1.DeletePropagationForeground)
	if err != nil {
		return err
	}
	return <-waitC
}

func (cl *knServingClient) deleteService(serviceName string, propagationPolicy v1.DeletionPropagation) error {
	err := cl.client.Services(cl.namespace).Delete(
		serviceName,
		&v1.DeleteOptions{PropagationPolicy: &propagationPolicy},
	)
	if err != nil {
		return clienterrors.GetError(err)
	}

	return nil
}

// Wait for a service to become ready, but not longer than provided timeout
func (cl *knServingClient) WaitForService(name string, timeout time.Duration, msgCallback wait.MessageCallback) (error, time.Duration) {
	waitForReady := wait.NewWaitForReady("service", cl.WatchService, serviceConditionExtractor)
	return waitForReady.Wait(name, wait.Options{Timeout: &timeout}, msgCallback)
}

// Get the configuration for a service
func (cl *knServingClient) GetConfiguration(name string) (*servingv1.Configuration, error) {
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
func (cl *knServingClient) GetRevision(name string) (*servingv1.Revision, error) {
	revision, err := cl.client.Revisions(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, clienterrors.GetError(err)
	}
	err = updateServingGvk(revision)
	if err != nil {
		return nil, err
	}
	return revision, nil
}

type NoBaseRevisionError struct {
	msg string
}

func (e NoBaseRevisionError) Error() string {
	return e.msg
}

var noBaseRevisionError = &NoBaseRevisionError{"base revision not found"}

// Get a "base" revision. This is the revision corresponding to the template of
// a Service. It may not be findable with our heuristics, in which case this
// method returns Errors()["no-base-revision"]. If it simply doesn't exist (like
// it wasn't yet created or was deleted), return the usual not found error.
func (cl *knServingClient) GetBaseRevision(service *servingv1.Service) (*servingv1.Revision, error) {
	return getBaseRevision(cl, service)
}

func getBaseRevision(cl KnServingClient, service *servingv1.Service) (*servingv1.Revision, error) {
	template := service.Spec.Template
	// First, try to get it by name. If the template has a particular name, the
	// base revision is the one created with that name.
	if template.Name != "" {
		return cl.GetRevision(template.Name)
	}
	// Next, let's try the LatestCreatedRevision, and see if that matches the
	// template, at least in terms of the image (which is what we care about here).
	if service.Status.LatestCreatedRevisionName != "" {
		latestCreated, err := cl.GetRevision(service.Status.LatestCreatedRevisionName)
		if err != nil {
			return nil, err
		}
		latestContainer, err := serving.ContainerOfRevisionSpec(&latestCreated.Spec)
		if err != nil {
			return nil, err
		}
		container, err := serving.ContainerOfRevisionTemplate(&template)
		if err != nil {
			return nil, err
		}
		if latestContainer.Image != container.Image {
			// At this point we know the latestCreatedRevision is out of date.
			return nil, noBaseRevisionError
		}
		// There is still some chance the latestCreatedRevision is out of date,
		// but we can't check the whole thing for equality because of
		// server-side defaulting. Since what we probably want it for is to
		// check the image digest anyway, keep it as good enough.
		return latestCreated, nil
	}
	return nil, noBaseRevisionError
}

// Delete a revision by name
func (cl *knServingClient) DeleteRevision(name string, timeout time.Duration) error {
	revision, err := cl.client.Revisions(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return clienterrors.GetError(err)
	}
	if revision.GetDeletionTimestamp() != nil {
		return fmt.Errorf("can't delete revision '%s' because it has been already marked for deletion", name)
	}
	if timeout == 0 {
		return cl.deleteRevision(name)
	}
	waitC := make(chan error)
	go func() {
		waitForEvent := wait.NewWaitForEvent("revision", cl.WatchRevision, func(evt *watch.Event) bool { return evt.Type == watch.Deleted })
		err, _ := waitForEvent.Wait(name, wait.Options{Timeout: &timeout}, wait.NoopMessageCallback())
		waitC <- err
	}()
	err = cl.deleteRevision(name)
	if err != nil {
		return clienterrors.GetError(err)
	}

	return <-waitC
}

func (cl *knServingClient) deleteRevision(name string) error {
	err := cl.client.Revisions(cl.namespace).Delete(name, &v1.DeleteOptions{})
	if err != nil {
		return clienterrors.GetError(err)
	}

	return nil
}

// List revisions
func (cl *knServingClient) ListRevisions(config ...ListConfig) (*servingv1.RevisionList, error) {
	revisionList, err := cl.client.Revisions(cl.namespace).List(ListConfigs(config).toListOptions())
	if err != nil {
		return nil, clienterrors.GetError(err)
	}
	return updateServingGvkForRevisionList(revisionList)
}

// Get a route by its unique name
func (cl *knServingClient) GetRoute(name string) (*servingv1.Route, error) {
	route, err := cl.client.Routes(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = updateServingGvk(route)
	if err != nil {
		return nil, err
	}
	return route, nil
}

// List routes
func (cl *knServingClient) ListRoutes(config ...ListConfig) (*servingv1.RouteList, error) {
	routeList, err := cl.client.Routes(cl.namespace).List(ListConfigs(config).toListOptions())
	if err != nil {
		return nil, err
	}
	return updateServingGvkForRouteList(routeList)
}

// update all the list + all items contained in the list with
// the proper GroupVersionKind specific to Knative serving
func updateServingGvkForRevisionList(revisionList *servingv1.RevisionList) (*servingv1.RevisionList, error) {
	revisionListNew := revisionList.DeepCopy()
	err := updateServingGvk(revisionListNew)
	if err != nil {
		return nil, err
	}

	revisionListNew.Items = make([]servingv1.Revision, len(revisionList.Items))
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
func updateServingGvkForRouteList(routeList *servingv1.RouteList) (*servingv1.RouteList, error) {
	routeListNew := routeList.DeepCopy()
	err := updateServingGvk(routeListNew)
	if err != nil {
		return nil, err
	}

	routeListNew.Items = make([]servingv1.Route, len(routeList.Items))
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

// update with the servingv1 group + version
func updateServingGvk(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, servingv1.SchemeGroupVersion, scheme.Scheme)
}

func serviceConditionExtractor(obj runtime.Object) (apis.Conditions, error) {
	service, ok := obj.(*servingv1.Service)
	if !ok {
		return nil, fmt.Errorf("%v is not a service", obj)
	}
	return apis.Conditions(service.Status.Conditions), nil
}
