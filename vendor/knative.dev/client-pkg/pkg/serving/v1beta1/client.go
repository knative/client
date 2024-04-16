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

package v1beta1

import (
	"context"
	"fmt"

	"knative.dev/client-pkg/pkg/config"

	"k8s.io/client-go/util/retry"

	duckv1 "knative.dev/pkg/apis/duck/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	knerrors "knative.dev/client-pkg/pkg/errors"
	"knative.dev/client-pkg/pkg/util"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"
	clientv1beta1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1beta1"
)

type DomainUpdateFunc func(origDomain *servingv1beta1.DomainMapping) (*servingv1beta1.DomainMapping, error)

// KnServingClient to work with Serving v1beta1 resources
type KnServingClient interface {
	// Namespace in which this client is operating for
	Namespace() string

	// GetDomainMapping
	GetDomainMapping(ctx context.Context, name string) (*servingv1beta1.DomainMapping, error)

	// CreateDomainMapping
	CreateDomainMapping(ctx context.Context, domainMapping *servingv1beta1.DomainMapping) error

	// UpdateDomainMapping
	UpdateDomainMapping(ctx context.Context, domainMapping *servingv1beta1.DomainMapping) error

	// UpdateDomainMappingWithRetry
	UpdateDomainMappingWithRetry(ctx context.Context, name string, updateFunc DomainUpdateFunc, nrRetries int) error

	// DeleteDomainMapping
	DeleteDomainMapping(ctx context.Context, name string) error

	// ListDomainMappings
	ListDomainMappings(ctx context.Context) (*servingv1beta1.DomainMappingList, error)
}

type knServingClient struct {
	client    clientv1beta1.ServingV1beta1Interface
	namespace string
}

// NewKnServingClient create a new client facade for the provided namespace
func NewKnServingClient(client clientv1beta1.ServingV1beta1Interface, namespace string) KnServingClient {
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
func (cl *knServingClient) GetDomainMapping(ctx context.Context, name string) (*servingv1beta1.DomainMapping, error) {
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
func (cl *knServingClient) CreateDomainMapping(ctx context.Context, domainMapping *servingv1beta1.DomainMapping) error {
	_, err := cl.client.DomainMappings(cl.namespace).Create(ctx, domainMapping, v1.CreateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return updateServingGvk(domainMapping)
}

// UpdateDomainMapping updates provided DomainMapping
func (cl *knServingClient) UpdateDomainMapping(ctx context.Context, domainMapping *servingv1beta1.DomainMapping) error {
	_, err := cl.client.DomainMappings(cl.namespace).Update(ctx, domainMapping, v1.UpdateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return updateServingGvk(domainMapping)
}

func (cl *knServingClient) UpdateDomainMappingWithRetry(ctx context.Context, name string, updateFunc DomainUpdateFunc, nrRetries int) error {
	return updateDomainMappingWithRetry(ctx, cl, name, updateFunc, nrRetries)
}

func updateDomainMappingWithRetry(ctx context.Context, cl KnServingClient, name string, updateFunc DomainUpdateFunc, nrRetries int) error {
	b := config.DefaultRetry
	b.Steps = nrRetries
	err := retry.RetryOnConflict(b, func() error {
		return updateDomain(ctx, cl, name, updateFunc)
	})
	return err
}

func updateDomain(ctx context.Context, c KnServingClient, name string, updateFunc DomainUpdateFunc) error {
	sub, err := c.GetDomainMapping(ctx, name)
	if err != nil {
		return err
	}
	if sub.GetDeletionTimestamp() != nil {
		return fmt.Errorf("can't update domain mapping %s because it has been marked for deletion", name)
	}
	updatedSource, err := updateFunc(sub.DeepCopy())
	if err != nil {
		return err
	}

	return c.UpdateDomainMapping(ctx, updatedSource)
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
func (cl *knServingClient) ListDomainMappings(ctx context.Context) (*servingv1beta1.DomainMappingList, error) {
	domainMappingList, err := cl.client.DomainMappings(cl.namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	dmListNew := domainMappingList.DeepCopy()
	err = updateServingGvk(dmListNew)
	if err != nil {
		return nil, err
	}
	dmListNew.Items = make([]servingv1beta1.DomainMapping, len(domainMappingList.Items))
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
	return util.UpdateGroupVersionKindWithScheme(obj, servingv1beta1.SchemeGroupVersion, scheme.Scheme)
}

// DomainMappingBuilder is for building the domainMapping
type DomainMappingBuilder struct {
	domainMapping *servingv1beta1.DomainMapping
}

// NewDomainMappingBuilder for building domainMapping object
func NewDomainMappingBuilder(name string) *DomainMappingBuilder {
	return &DomainMappingBuilder{domainMapping: &servingv1beta1.DomainMapping{
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
func (b *DomainMappingBuilder) TLS(cert string) *DomainMappingBuilder {
	if cert == "" {
		return b
	}
	b.domainMapping.Spec.TLS = &servingv1beta1.SecretTLS{SecretName: cert}
	return b
}

// Build to return an instance of domainMapping object
func (b *DomainMappingBuilder) Build() *servingv1beta1.DomainMapping {
	return b.domainMapping
}
