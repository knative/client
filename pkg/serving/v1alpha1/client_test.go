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

	serving_api "github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/knative/client/pkg/serving"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KnClient Tests")
}

var namespace = "test-ns"

var _ = Describe("Client", func() {

	var (
		knClient KnClient
		serving  fake.FakeServingV1alpha1
	)

	BeforeEach(func() {
		serving = fake.FakeServingV1alpha1{Fake: &client_testing.Fake{}}
		knClient = NewKnServingClient(&serving, namespace)
	})

	// -----------------------------------------------------------------------------------------
	Describe("Getting a service", func() {
		const (
			serviceName            = "test-service"
			notExistingServiceName = "no-service"
		)
		var (
			service *v1alpha1.Service
		)

		BeforeEach(func() {
			service = newService(serviceName)
			serving.AddReactor("get", "services",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					name := a.(client_testing.GetAction).GetName()
					// Sanity check
					Expect(name).ToNot(BeEmpty())
					Expect(a.GetNamespace()).To(Equal(namespace))
					if name == serviceName {
						return true, service, nil
					}
					return true, nil, errors.NewNotFound(v1alpha1.Resource("service"), name)
				})
		})

		Context("with a name that exists", func() {
			It("returns a service with this name and without error", func() {
				getService, err := knClient.GetService(serviceName)
				Expect(err).To(BeNil())
				Expect(getService.Name).To(Equal(service.Name))
				validateGvk(getService)
			})
		})

		Context("with a name that does not exist", func() {
			It("returns a nil object and error with a descriptive error message", func() {
				getService, err := knClient.GetService(notExistingServiceName)
				Expect(getService).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring(notExistingServiceName))
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})
	})

	// -----------------------------------------------------------------------------------------
	Describe("Listing services", func() {
		var (
			service1 = newService("service-1")
			service2 = newService("service-2")
		)

		BeforeEach(func() {
			serving.AddReactor("list", "services",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					Expect(a.GetNamespace()).To(Equal(namespace))
					return true, &v1alpha1.ServiceList{Items: []v1alpha1.Service{*service1, *service2}}, nil
				})
		})

		It("returns a list of services", func() {
			listServices, err := knClient.ListServices()
			Expect(err).To(BeNil())
			Expect(listServices.Items).To(HaveLen(2))
			Expect(listServices.Items[0].Name).To(Equal("service-1"))
			Expect(listServices.Items[1].Name).To(Equal("service-2"))
			validateGvk(listServices)
			validateGvk(&listServices.Items[0])
			validateGvk(&listServices.Items[1])
		})
	})

	// -----------------------------------------------------------------------------------------
	Describe("Creating a service", func() {
		var (
			serviceNew = newService("new-service")
		)

		BeforeEach(func() {
			serving.AddReactor("create", "services",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					Expect(a.GetNamespace()).To(Equal(namespace))
					name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
					if name == serviceNew.Name {
						serviceNew.Generation = 2
						return true, serviceNew, nil
					}
					return true, nil, fmt.Errorf("error while creating service %s", name)
				})
		})

		Context("without error", func() {
			It("creates a new service", func() {
				err := knClient.CreateService(serviceNew)
				Expect(err).To(BeNil())
				Expect(serviceNew.Generation).To(Equal(int64(2)))
				validateGvk(serviceNew)
			})
		})

		Context("with error", func() {
			It("returns error object", func() {
				err := knClient.CreateService(newService("unknown"))
				Expect(err).NotTo(BeNil())
			})
		})
	})

	// -----------------------------------------------------------------------------------------
	Describe("Updating a service", func() {
		var (
			serviceUpdate = newService("update-service")
		)

		BeforeEach(func() {
			serving.AddReactor("update", "services",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					Expect(a.GetNamespace()).To(Equal(namespace))
					name := a.(client_testing.UpdateAction).GetObject().(metav1.Object).GetName()
					if name == serviceUpdate.Name {
						serviceUpdate.Generation = 3
						return true, serviceUpdate, nil
					}
					return true, nil, fmt.Errorf("error while updating service %s", name)
				})
		})

		Context("without error", func() {
			It("updates the service", func() {
				err := knClient.UpdateService(serviceUpdate)
				Expect(err).To(BeNil())
				Expect(serviceUpdate.Generation).To(Equal(int64(3)))
				validateGvk(serviceUpdate)
			})
		})

		Context("with error", func() {
			It("returns error object", func() {
				err := knClient.UpdateService(newService("unknown"))
				Expect(err).NotTo(BeNil())
			})
		})
	})

	// -----------------------------------------------------------------------------------------
	Describe("Deleting a service", func() {
		const (
			serviceName            = "test-service"
			notExistingServiceName = "no-service"
		)

		BeforeEach(func() {
			serving.AddReactor("delete", "services",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					name := a.(client_testing.DeleteAction).GetName()
					// Sanity check
					Expect(name).ToNot(BeEmpty())
					Expect(a.GetNamespace()).To(Equal(namespace))
					if name == serviceName {
						return true, nil, nil
					}
					return true, nil, errors.NewNotFound(v1alpha1.Resource("service"), name)
				})
		})

		Context("with a name that exists", func() {
			It("returns no error", func() {
				err := knClient.DeleteService(serviceName)
				Expect(err).To(BeNil())
			})
		})

		Context("with a name that does not exist", func() {
			It("returns an error", func() {
				err := knClient.DeleteService(notExistingServiceName)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring(notExistingServiceName))
			})
		})
	})

	// -----------------------------------------------------------------------------------------
	Describe("Getting a revision", func() {
		const (
			revisionName            = "test-revision"
			notExistingRevisionName = "no-revision"
		)
		var (
			revision *v1alpha1.Revision
		)

		BeforeEach(func() {
			revision = newRevision(revisionName)
			serving.AddReactor("get", "revisions",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					name := a.(client_testing.GetAction).GetName()
					// Sanity check
					Expect(name).ToNot(BeEmpty())
					Expect(a.GetNamespace()).To(Equal(namespace))
					if name == revisionName {
						return true, revision, nil
					}
					return true, nil, errors.NewNotFound(v1alpha1.Resource("revision"), name)
				})
		})

		Context("with a name that exists", func() {
			It("returns a revision with this name and without error", func() {
				getRevision, err := knClient.GetRevision(revisionName)
				Expect(err).To(BeNil())
				Expect(getRevision.Name).To(Equal(revision.Name))
				validateGvk(getRevision)
			})
		})

		Context("with a name that does not exist", func() {
			It("returns a nil object and error with a descriptive error message", func() {
				getRevisions, err := knClient.GetRevision(notExistingRevisionName)
				Expect(getRevisions).To(BeNil())
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring(notExistingRevisionName))
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})
	})

	// -----------------------------------------------------------------------------------------
	Describe("Listing revisions", func() {
		var (
			revisions = []v1alpha1.Revision{*newRevision("revision-1"), *newRevision("revision-2")}
		)

		BeforeEach(func() {
			serving.AddReactor("list", "revisions",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					Expect(a.GetNamespace()).To(Equal(namespace))
					return true, &v1alpha1.RevisionList{Items: revisions}, nil
				})
		})

		It("returns a list of services", func() {
			listRevisions, err := knClient.ListRevisions()
			Expect(err).To(BeNil())
			Expect(listRevisions.Items).To(HaveLen(2))
			Expect(listRevisions.Items[0].Name).To(Equal("revision-1"))
			Expect(listRevisions.Items[1].Name).To(Equal("revision-2"))
			validateGvk(listRevisions)
			validateGvk(&listRevisions.Items[0])
			validateGvk(&listRevisions.Items[1])
		})
	})

	// -----------------------------------------------------------------------------------------
	Describe("Listing revisions by service", func() {
		var (
			revisions = []v1alpha1.Revision{
				*newRevision("revision-1", serving_api.ServiceLabelKey, "service"),
				*newRevision("revision-2"),
			}
			serviceName = "service"
		)

		BeforeEach(func() {
			serving.AddReactor("list", "revisions",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					lAction := a.(client_testing.ListAction)
					Expect(a.GetNamespace()).To(Equal(namespace))
					restrictions := lAction.GetListRestrictions()
					Expect(restrictions.Fields.Empty()).To(BeTrue())
					servicesLabels := labels.Set{serving_api.ServiceLabelKey: serviceName}
					Expect(restrictions.Labels.Matches(servicesLabels)).To(BeTrue())
					return true, &v1alpha1.RevisionList{Items: revisions}, nil
				})
		})

		It("returns a list of revisions associated with this this service", func() {
			listRevisions, err := knClient.ListRevisionsForService(serviceName)
			Expect(err).To(BeNil())
			Expect(listRevisions.Items).To(HaveLen(1))
			Expect(listRevisions.Items[0].Name).To(Equal("revision-1"))
			Expect(listRevisions.Items[0].Labels[serving_api.ServiceLabelKey]).To(Equal("service"))
			validateGvk(listRevisions)
			validateGvk(&listRevisions.Items[0])
		})
	})

})

func validateGvk(obj runtime.Object) {
	gvkExpected, err := serving.GetGroupVersionKind(obj, v1alpha1.SchemeGroupVersion)
	Expect(err).To(BeNil())
	gvkGiven := obj.GetObjectKind().GroupVersionKind()
	Expect(gvkGiven).To(Equal(*gvkExpected))
}

func newService(name string) *v1alpha1.Service {
	return &v1alpha1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
}

func newRevision(name string, labels ...string) *v1alpha1.Revision {
	rev := &v1alpha1.Revision{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	labelMap := make(map[string]string)
	if len(labels) > 0 {
		for i := 0; i < len(labels); i += 2 {
			labelMap[labels[i]] = labels[i+1]
		}
		rev.Labels = labelMap
	}
	return rev
}
