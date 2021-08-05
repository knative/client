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

	duckv1 "knative.dev/pkg/apis/duck/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/util"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"
	clientv1alpha1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

// KnServingClient to work with Serving v1alpha1 resources
type KnServingClient interface {
	// Namespace in which this client is operating for
	Namespace() string

	// GetDomainMapping
	GetDomainMapping(ctx context.Context, name string) (*servingv1alpha1.DomainMapping, error)

	// CreateDomainMapping
	CreateDomainMapping(ctx context.Context, domainMapping *servingv1alpha1.DomainMapping) error

	// UpdateDomainMapping
	UpdateDomainMapping(ctx context.Context, domainMapping *servingv1alpha1.DomainMapping) error

	// DeleteDomainMapping
	DeleteDomainMapping(ctx context.Context, name string) error

	// ListDomainMappings
	ListDomainMappings(ctx context.Context) (*servingv1alpha1.DomainMappingList, error)
}

type knServingClient struct {
	client    clientv1alpha1.ServingV1alpha1Interface
	namespace string
}

// NewKnServingClient create a new client facade for the provided namespace
func NewKnServingClient(client clientv1alpha1.ServingV1alpha1Interface, namespace string) KnServingClient {
	return &knServingClient{
		client:    client,
		namespace: namespace,
	}
}

// Namespace in which this client is operating for
func (cl *knServingClient) Namespace() string {
	return cl.namespace
}

// GetDomainMapping gets DomainMapping by name
func (cl *knServingClient) GetDomainMapping(ctx context.Context, name string) (*servingv1alpha1.DomainMapping, error) {
	dm, err := cl.client.DomainMappings(cl.namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	err = updateServingGvk(dm)
	if err != nil {
		return nil, err
	}
	return dm, nil
}

// CreateDomainMapping creates provided DomainMapping
func (cl *knServingClient) CreateDomainMapping(ctx context.Context, domainMapping *servingv1alpha1.DomainMapping) error {
	_, err := cl.client.DomainMappings(cl.namespace).Create(ctx, domainMapping, v1.CreateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return updateServingGvk(domainMapping)
}

// UpdateDomainMapping updates provided DomainMapping
func (cl *knServingClient) UpdateDomainMapping(ctx context.Context, domainMapping *servingv1alpha1.DomainMapping) error {
	_, err := cl.client.DomainMappings(cl.namespace).Update(ctx, domainMapping, v1.UpdateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return updateServingGvk(domainMapping)
}

// DeleteDomainMapping deletes DomainMapping by name
func (cl *knServingClient) DeleteDomainMapping(ctx context.Context, name string) error {
	err := cl.client.DomainMappings(cl.namespace).Delete(ctx, name, v1.DeleteOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return nil
}

// ListDomainMappings lists all DomainMappings
func (cl *knServingClient) ListDomainMappings(ctx context.Context) (*servingv1alpha1.DomainMappingList, error) {
	domainMappingList, err := cl.client.DomainMappings(cl.namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	dmListNew := domainMappingList.DeepCopy()
	err = updateServingGvk(dmListNew)
	if err != nil {
		return nil, err
	}
	dmListNew.Items = make([]servingv1alpha1.DomainMapping, len(domainMappingList.Items))
	for idx, domainMapping := range domainMappingList.Items {
		domainMappingClone := domainMapping.DeepCopy()
		err := updateServingGvk(domainMappingClone)
		if err != nil {
			return nil, err
		}
		dmListNew.Items[idx] = *domainMappingClone
	}
	return dmListNew, nil
}

func updateServingGvk(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, servingv1alpha1.SchemeGroupVersion, scheme.Scheme)
}

// DomainMappingBuilder is for building the domainMapping
type DomainMappingBuilder struct {
	domainMapping *servingv1alpha1.DomainMapping
}

// NewDomainMappingBuilder for building domainMapping object
func NewDomainMappingBuilder(name string) *DomainMappingBuilder {
	return &DomainMappingBuilder{domainMapping: &servingv1alpha1.DomainMapping{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}}
}

// Namespace for domainMapping builder
func (b *DomainMappingBuilder) Namespace(ns string) *DomainMappingBuilder {
	b.domainMapping.Namespace = ns
	return b
}

// Reference for domainMapping builder
func (b *DomainMappingBuilder) Reference(reference duckv1.KReference) *DomainMappingBuilder {
	b.domainMapping.Spec.Ref = reference
	return b
}

// TLS for domainMapping builder
func (b *DomainMappingBuilder) TLS(tls string) *DomainMappingBuilder {
	if tls == "" {
		return b
	}
	b.domainMapping.Spec.TLS = &servingv1alpha1.SecretTLS{SecretName: tls}
	return b
}

// Build to return an instance of domainMapping object
func (b *DomainMappingBuilder) Build() *servingv1alpha1.DomainMapping {
	return b.domainMapping
}
