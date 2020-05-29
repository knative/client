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

package v1alpha2

import (
	"fmt"

	apisv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
	clientv1alpha2 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"

	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/util"
)

// KnSinkBindingClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnSinkBindingClient interface {
	// Namespace in which this client is operating for
	Namespace() string
	// CreateSinkBinding is used to create an instance of binding
	CreateSinkBinding(binding *v1alpha2.SinkBinding) error
	// DeleteSinkBinding is used to delete an instance of binding
	DeleteSinkBinding(name string) error
	// GetSinkBinding is used to get an instance of binding
	GetSinkBinding(name string) (*v1alpha2.SinkBinding, error)
	// ListSinkBinding returns list of binding CRDs
	ListSinkBindings() (*v1alpha2.SinkBindingList, error)
	// UpdateSinkBinding is used to update an instance of binding
	UpdateSinkBinding(binding *v1alpha2.SinkBinding) error
}

// KnSinkBindingClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type knBindingClient struct {
	client    clientv1alpha2.SinkBindingInterface
	namespace string
}

// NewKnSourcesClient is to invoke Eventing Sources Client API to create object
func newKnSinkBindingClient(client clientv1alpha2.SinkBindingInterface, namespace string) KnSinkBindingClient {
	return &knBindingClient{
		client:    client,
		namespace: namespace,
	}
}

//CreateSinkBinding is used to create an instance of binding
func (c *knBindingClient) CreateSinkBinding(binding *v1alpha2.SinkBinding) error {
	_, err := c.client.Create(binding)
	if err != nil {
		return knerrors.GetError(err)
	}
	return nil
}

//DeleteSinkBinding is used to delete an instance of binding
func (c *knBindingClient) DeleteSinkBinding(name string) error {
	err := c.client.Delete(name, &apisv1.DeleteOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return nil
}

//GetSinkBinding is used to get an instance of binding
func (c *knBindingClient) GetSinkBinding(name string) (*v1alpha2.SinkBinding, error) {
	binding, err := c.client.Get(name, apisv1.GetOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	return binding, nil
}

func (c *knBindingClient) ListSinkBindings() (*v1alpha2.SinkBindingList, error) {
	bindingList, err := c.client.List(apisv1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	bindingListNew := bindingList.DeepCopy()
	err = updateSinkBindingGvk(bindingListNew)
	if err != nil {
		return nil, err
	}

	bindingListNew.Items = make([]v1alpha2.SinkBinding, len(bindingList.Items))
	for idx, binding := range bindingList.Items {
		bindingClone := binding.DeepCopy()
		err := updateSinkBindingGvk(bindingClone)
		if err != nil {
			return nil, err
		}
		bindingListNew.Items[idx] = *bindingClone
	}
	return bindingListNew, nil
}

//CreateSinkBinding is used to create an instance of binding
func (c *knBindingClient) UpdateSinkBinding(binding *v1alpha2.SinkBinding) error {
	_, err := c.client.Update(binding)
	if err != nil {
		return knerrors.GetError(err)
	}
	return nil
}

// Return the client's namespace
func (c *knBindingClient) Namespace() string {
	return c.namespace
}

// update with the v1alpha2 group + version
func updateSinkBindingGvk(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, v1alpha2.SchemeGroupVersion, scheme.Scheme)
}

// SinkBindingBuilder is for building the binding
type SinkBindingBuilder struct {
	binding        *v1alpha2.SinkBinding
	sGvk           *schema.GroupVersionKind
	sName          string
	sLabelSelector map[string]string
	sNamespace     string

	// When set directly:
	subject *tracker.Reference
}

// NewSinkBindingBuilder for building binding object
func NewSinkBindingBuilder(name string) *SinkBindingBuilder {
	return &SinkBindingBuilder{binding: &v1alpha2.SinkBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

// NewSinkBindingBuilderFromExisting for building the object from existing SinkBinding object
func NewSinkBindingBuilderFromExisting(binding *v1alpha2.SinkBinding) *SinkBindingBuilder {
	return &SinkBindingBuilder{binding: binding.DeepCopy()}
}

// Namespace for this binding
func (b *SinkBindingBuilder) Namespace(ns string) *SinkBindingBuilder {
	b.binding.Namespace = ns
	return b
}

// Subscriber for the binding to send to (it's a Sink actually)
func (b *SinkBindingBuilder) Subject(subject *tracker.Reference) *SinkBindingBuilder {
	b.subject = subject
	return b
}

// Add a GVK of the subject
func (b *SinkBindingBuilder) SubjectGVK(gvk *schema.GroupVersionKind) *SinkBindingBuilder {
	b.sGvk = gvk
	return b
}

// Add a subject name for building up the name
func (b *SinkBindingBuilder) SubjectName(name string) *SinkBindingBuilder {
	b.sName = name
	return b
}

// Add a subject namespace for building up the name
func (b *SinkBindingBuilder) SubjectNamespace(ns string) *SinkBindingBuilder {
	b.sNamespace = ns
	return b
}

// Add a label match part for building up the subject
func (b *SinkBindingBuilder) AddSubjectMatchLabel(labelKey, labelValue string) *SinkBindingBuilder {
	if b.sLabelSelector == nil {
		b.sLabelSelector = map[string]string{}
	}
	b.sLabelSelector[labelKey] = labelValue
	return b
}

// Broker to set the broker of binding object
func (b *SinkBindingBuilder) Sink(sink *duckv1.Destination) *SinkBindingBuilder {
	b.binding.Spec.Sink = *sink
	return b
}

// CloudEventOverrides adds given Cloud Event override extensions map to source spec
func (b *SinkBindingBuilder) CloudEventOverrides(ceo map[string]string, toRemove []string) *SinkBindingBuilder {
	if ceo == nil && len(toRemove) == 0 {
		return b
	}

	ceOverrides := b.binding.Spec.CloudEventOverrides
	if ceOverrides == nil {
		ceOverrides = &duckv1.CloudEventOverrides{Extensions: map[string]string{}}
		b.binding.Spec.CloudEventOverrides = ceOverrides
	}
	for k, v := range ceo {
		ceOverrides.Extensions[k] = v
	}
	for _, r := range toRemove {
		delete(ceOverrides.Extensions, r)
	}

	return b
}

// Build to return an instance of binding object
func (b *SinkBindingBuilder) Build() (*v1alpha2.SinkBinding, error) {
	// If set directly, return the sink binding directly
	if b.subject != nil {
		b.binding.Spec.Subject = *b.subject
		return b.binding, nil
	}

	if b.sGvk == nil && b.sName == "" && b.sLabelSelector == nil {
		// None of the subject methods has been called, so no subject build up
		return b.binding, nil
	}

	// Otherwise, validate and build up the subject
	if b.sGvk == nil {
		return nil, fmt.Errorf("no group-version-kind provided for creating binding %s", b.binding.Name)
	}

	if b.sName != "" && b.sLabelSelector != nil {
		return nil, fmt.Errorf("either a subject name or label selector can be used for creating binding %s, but not both (subject name: %s, label selector: %v", b.binding.Name, b.sName, b.sLabelSelector)
	}

	subject := b.prepareBaseSubject()
	if b.sName != "" {
		subject.Name = b.sName
	} else {
		subject.Selector = &metav1.LabelSelector{
			MatchLabels: b.sLabelSelector,
		}
	}

	b.binding.Spec.Subject = subject
	return b.binding, nil
}

func (b *SinkBindingBuilder) prepareBaseSubject() tracker.Reference {
	subject := tracker.Reference{
		APIVersion: b.sGvk.GroupVersion().String(),
		Kind:       b.sGvk.Kind,
	}
	if b.sNamespace != "" {
		subject.Namespace = b.sNamespace
	}
	return subject
}
