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

package v1beta2

import (
	"context"
	"fmt"

	"knative.dev/client-pkg/pkg/config"

	"k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/client-pkg/pkg/util"
	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	knerrors "knative.dev/client-pkg/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"

	clientv1beta2 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1beta2"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

type PingSourceUpdateFunc func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error)

// Interface for interacting with a Ping source
type KnPingSourcesClient interface {

	// GetPingSource fetches a Ping source by its name
	GetPingSource(ctx context.Context, name string) (*sourcesv1beta2.PingSource, error)

	// CreatePingSource creates a Ping source
	CreatePingSource(ctx context.Context, pingSource *sourcesv1beta2.PingSource) error

	// UpdatePingSource updates a Ping source
	UpdatePingSource(ctx context.Context, pingSource *sourcesv1beta2.PingSource) error

	// UpdatePingSourceWithRetry updates a Ping source and retries on conflict
	UpdatePingSourceWithRetry(ctx context.Context, name string, updateFunc PingSourceUpdateFunc, nrRetries int) error

	// DeletePingSource deletes a Ping source
	DeletePingSource(ctx context.Context, name string) error

	// ListPingSource lists all Ping sources
	// TODO: Support list configs like in service list
	ListPingSource(ctx context.Context) (*sourcesv1beta2.PingSourceList, error)

	// Get namespace for this source
	Namespace() string
}

// knSourcesClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type pingSourcesClient struct {
	client    clientv1beta2.PingSourceInterface
	namespace string
}

// NewKnSourcesClient is to invoke Eventing Sources Client API to create object
func newKnPingSourcesClient(client clientv1beta2.PingSourceInterface, namespace string) KnPingSourcesClient {
	return &pingSourcesClient{
		client:    client,
		namespace: namespace,
	}
}

// Get the namespace for which this client has been created
func (c *pingSourcesClient) Namespace() string {
	return c.namespace
}

func (c *pingSourcesClient) CreatePingSource(ctx context.Context, pingsource *sourcesv1beta2.PingSource) error {
	if pingsource.Spec.Sink.Ref == nil && pingsource.Spec.Sink.URI == nil {
		return fmt.Errorf("a sink is required for creating a source")
	}
	_, err := c.client.Create(ctx, pingsource, metav1.CreateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return nil
}

func (c *pingSourcesClient) UpdatePingSource(ctx context.Context, pingSource *sourcesv1beta2.PingSource) error {
	_, err := c.client.Update(ctx, pingSource, metav1.UpdateOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return nil
}

func (c *pingSourcesClient) UpdatePingSourceWithRetry(ctx context.Context, name string, updateFunc PingSourceUpdateFunc, nrRetries int) error {
	return updatePingSourceWithRetry(ctx, c, name, updateFunc, nrRetries)
}

func updatePingSourceWithRetry(ctx context.Context, c KnPingSourcesClient, name string, updateFunc PingSourceUpdateFunc, nrRetries int) error {
	b := config.DefaultRetry
	b.Steps = nrRetries
	err := retry.RetryOnConflict(b, func() error {
		return updatePingSource(ctx, c, name, updateFunc)
	})
	return err
}

func updatePingSource(ctx context.Context, c KnPingSourcesClient, name string, updateFunc PingSourceUpdateFunc) error {
	source, err := c.GetPingSource(ctx, name)
	if err != nil {
		return err
	}
	if source.GetDeletionTimestamp() != nil {
		return fmt.Errorf("can't update ping source %s because it has been marked for deletion", name)
	}
	updatedSource, err := updateFunc(source.DeepCopy())
	if err != nil {
		return err
	}

	return c.UpdatePingSource(ctx, updatedSource)
}

func (c *pingSourcesClient) DeletePingSource(ctx context.Context, name string) error {
	err := c.client.Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return knerrors.GetError(err)
	}
	return nil
}

func (c *pingSourcesClient) GetPingSource(ctx context.Context, name string) (*sourcesv1beta2.PingSource, error) {
	source, err := c.client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}
	err = updateSourceGVK(source)
	if err != nil {
		return nil, err
	}
	return source, nil
}

// ListPingSource returns the available Ping sources
func (c *pingSourcesClient) ListPingSource(ctx context.Context) (*sourcesv1beta2.PingSourceList, error) {
	sourceList, err := c.client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, knerrors.GetError(err)
	}

	return updatePingSourceListGVK(sourceList)
}

func updateSourceGVK(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, sourcesv1beta2.SchemeGroupVersion, scheme.Scheme)
}

func updatePingSourceListGVK(sourceList *sourcesv1beta2.PingSourceList) (*sourcesv1beta2.PingSourceList, error) {
	sourceListNew := sourceList.DeepCopy()
	err := updateSourceGVK(sourceListNew)
	if err != nil {
		return nil, err
	}

	sourceListNew.Items = make([]sourcesv1beta2.PingSource, len(sourceList.Items))
	for idx, source := range sourceList.Items {
		sourceClone := source.DeepCopy()
		err := updateSourceGVK(sourceClone)
		if err != nil {
			return nil, err
		}
		sourceListNew.Items[idx] = *sourceClone
	}
	return sourceListNew, nil
}

// Builder for building up Ping sources

type PingSourceBuilder struct {
	pingSource *sourcesv1beta2.PingSource
}

func NewPingSourceBuilder(name string) *PingSourceBuilder {
	return &PingSourceBuilder{pingSource: &sourcesv1beta2.PingSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

func NewPingSourceBuilderFromExisting(pingsource *sourcesv1beta2.PingSource) *PingSourceBuilder {
	return &PingSourceBuilder{pingSource: pingsource.DeepCopy()}
}

func (b *PingSourceBuilder) Schedule(schedule string) *PingSourceBuilder {
	b.pingSource.Spec.Schedule = schedule
	return b
}

func (b *PingSourceBuilder) Data(data string) *PingSourceBuilder {
	b.pingSource.Spec.Data = data
	return b
}

func (b *PingSourceBuilder) DataBase64(data string) *PingSourceBuilder {
	b.pingSource.Spec.DataBase64 = data
	return b
}

func (b *PingSourceBuilder) Sink(sink duckv1.Destination) *PingSourceBuilder {
	b.pingSource.Spec.Sink = sink
	return b
}

// CloudEventOverrides adds given Cloud Event override extensions map to source spec
func (b *PingSourceBuilder) CloudEventOverrides(ceo map[string]string, toRemove []string) *PingSourceBuilder {
	if ceo == nil && len(toRemove) == 0 {
		return b
	}

	ceOverrides := b.pingSource.Spec.CloudEventOverrides
	if ceOverrides == nil {
		ceOverrides = &duckv1.CloudEventOverrides{Extensions: map[string]string{}}
		b.pingSource.Spec.CloudEventOverrides = ceOverrides
	}
	for k, v := range ceo {
		ceOverrides.Extensions[k] = v
	}
	for _, r := range toRemove {
		delete(ceOverrides.Extensions, r)
	}

	return b
}

func (b *PingSourceBuilder) Build() *sourcesv1beta2.PingSource {
	return b.pingSource
}
