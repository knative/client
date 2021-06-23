/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/util"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"
	clientv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// KnContainerSourcesClient interface for working with ApiServer sources
type KnContainerSourcesClient interface {

	// Get an ContainerSource by name
	GetContainerSource(ctx context.Context, name string) (*v1.ContainerSource, error)

	// Create an ContainerSource by object
	CreateContainerSource(ctx context.Context, containerSrc *v1.ContainerSource) error

	// Update an ContainerSource by object
	UpdateContainerSource(ctx context.Context, containerSrc *v1.ContainerSource) error

	// Delete an ContainerSource by name
	DeleteContainerSource(name string, ctx context.Context) error

	// List ContainerSource
	ListContainerSources(ctx context.Context) (*v1.ContainerSourceList, error)

	// Get namespace for this client
	Namespace() string
}

// knSourcesClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type containerSourcesClient struct {
	client    clientv1.ContainerSourceInterface
	namespace string
}

// newKnContainerSourcesClient is to invoke Eventing Sources Client API to create object
func newKnContainerSourcesClient(client clientv1.ContainerSourceInterface, namespace string) KnContainerSourcesClient {
	return &containerSourcesClient{
		client:    client,
		namespace: namespace,
	}
}

//GetContainerSource returns containerSrc object if present
func (c *containerSourcesClient) GetContainerSource(ctx context.Context, name string) (*v1.ContainerSource, error) {
	containerSrc, err := c.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}

	return containerSrc, nil
}

//CreateContainerSource is used to create an instance of ContainerSource
func (c *containerSourcesClient) CreateContainerSource(ctx context.Context, containerSrc *v1.ContainerSource) error {
	_, err := c.client.Create(ctx, containerSrc, metav1.CreateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}

	return nil
}

//UpdateContainerSource is used to update an instance of ContainerSource
func (c *containerSourcesClient) UpdateContainerSource(ctx context.Context, containerSrc *v1.ContainerSource) error {
	_, err := c.client.Update(ctx, containerSrc, metav1.UpdateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}

	return nil
}

//DeleteContainerSource is used to create an instance of ContainerSource
func (c *containerSourcesClient) DeleteContainerSource(name string, ctx context.Context) error {
	return c.client.Delete(ctx, name, metav1.DeleteOptions{})
}

// Return the client's namespace
func (c *containerSourcesClient) Namespace() string {
	return c.namespace
}

// ListContainerSource returns the available container sources
func (c *containerSourcesClient) ListContainerSources(ctx context.Context) (*v1.ContainerSourceList, error) {
	sourceList, err := c.client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}

	containerListNew := sourceList.DeepCopy()
	err = updateContainerSourceGvk(containerListNew)
	if err != nil {
		return nil, err
	}

	containerListNew.Items = make([]v1.ContainerSource, len(sourceList.Items))
	for idx, binding := range sourceList.Items {
		bindingClone := binding.DeepCopy()
		err := updateSinkBindingGvk(bindingClone)
		if err != nil {
			return nil, err
		}
		containerListNew.Items[idx] = *bindingClone
	}

	return containerListNew, nil
}

// update with the v1 group + version
func updateContainerSourceGvk(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, v1.SchemeGroupVersion, scheme.Scheme)
}

// ContainerSourceBuilder is for building the source
type ContainerSourceBuilder struct {
	ContainerSource *v1.ContainerSource
}

// NewContainerSourceBuilder for building Container source object
func NewContainerSourceBuilder(name string) *ContainerSourceBuilder {
	return &ContainerSourceBuilder{ContainerSource: &v1.ContainerSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

// NewContainerSourceBuilderFromExisting for building the object from existing ContainerSource object
func NewContainerSourceBuilderFromExisting(ContainerSource *v1.ContainerSource) *ContainerSourceBuilder {
	return &ContainerSourceBuilder{ContainerSource: ContainerSource.DeepCopy()}
}

// Sink or destination of the source
func (b *ContainerSourceBuilder) Sink(sink duckv1.Destination) *ContainerSourceBuilder {
	b.ContainerSource.Spec.Sink = sink
	return b
}

func (b *ContainerSourceBuilder) CloudEventOverrides(ceo map[string]string, toRemove []string) *ContainerSourceBuilder {
	if ceo == nil && len(toRemove) == 0 {
		return b
	}

	ceOverrides := b.ContainerSource.Spec.CloudEventOverrides
	if ceOverrides == nil {
		ceOverrides = &duckv1.CloudEventOverrides{Extensions: map[string]string{}}
		b.ContainerSource.Spec.CloudEventOverrides = ceOverrides
	}
	for k, v := range ceo {
		ceOverrides.Extensions[k] = v
	}
	for _, r := range toRemove {
		delete(ceOverrides.Extensions, r)
	}

	return b
}

// Build the ContainerSource object
func (b *ContainerSourceBuilder) Build() *v1.ContainerSource {
	return b.ContainerSource
}

// PodSpec defines the PodSpec
func (b *ContainerSourceBuilder) PodSpec(podSpec corev1.PodSpec) *ContainerSourceBuilder {
	b.ContainerSource.Spec.Template.Spec = podSpec
	return b
}
