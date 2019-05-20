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
	"io"
	"time"

	"github.com/knative/pkg/apis"

	"github.com/knative/client/pkg/serving"
	"github.com/knative/client/pkg/wait"

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

	// List all services
	ListServices() (*v1alpha1.ServiceList, error)

	// Create a new service
	CreateService(service *v1alpha1.Service) error

	// Update the given service
	UpdateService(service *v1alpha1.Service) error

	// Delete a service by name
	DeleteService(name string) error

	// Wait for a service to become ready, but not longer than provided timeout
	WaitForService(name string, timeout time.Duration, out io.Writer) error

	// Get a revision by name
	GetRevision(name string) (*v1alpha1.Revision, error)

	// List all revisions
	ListRevisions() (*v1alpha1.RevisionList, error)

	// Get all revisions for a specific service
	ListRevisionsForService(serviceName string) (*v1alpha1.RevisionList, error)

	// Delete a revision
	DeleteRevision(name string) error
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
		return nil, err
	}
	err = serving.UpdateGroupVersionKind(service, v1alpha1.SchemeGroupVersion)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// List all services
func (cl *knClient) ListServices() (*v1alpha1.ServiceList, error) {
	serviceList, err := cl.client.Services(cl.namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
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
		return err
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
	return cl.client.Services(cl.namespace).Delete(
		serviceName,
		&v1.DeleteOptions{},
	)
}

// Wait for a service to become ready, but not longer than provided timeout
func (cl *knClient) WaitForService(name string, timeout time.Duration, out io.Writer) error {
	waitForReady := newServiceWaitForReady(cl.client.Services(cl.namespace).Watch)
	return waitForReady.Wait(name, timeout, out)
}

// Get a revision by name
func (cl *knClient) GetRevision(name string) (*v1alpha1.Revision, error) {
	revision, err := cl.client.Revisions(cl.namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = updateServingGvk(revision)
	if err != nil {
		return nil, err
	}
	return revision, nil
}

// Delete a revision by name
func (cl *knClient) DeleteRevision(name string) error {
	return cl.client.Revisions(cl.namespace).Delete(name, &v1.DeleteOptions{})
}

// List all revisions
func (cl *knClient) ListRevisions() (*v1alpha1.RevisionList, error) {
	revisionList, err := cl.client.Revisions(cl.namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return updateServingGvkForRevisionList(revisionList)
}

func (cl *knClient) ListRevisionsForService(serviceName string) (*v1alpha1.RevisionList, error) {
	listOptions := v1.ListOptions{}
	listOptions.LabelSelector = labels.Set(
		map[string]string{api_serving.ServiceLabelKey: serviceName}).String()

	revisionList, err := cl.client.Revisions(cl.namespace).List(listOptions)
	if err != nil {
		return nil, err
	}
	return updateServingGvkForRevisionList(revisionList)
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
