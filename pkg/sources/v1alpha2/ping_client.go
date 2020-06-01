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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/apis/sources/v1alpha2"

	clientv1alpha2 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// Interface for interacting with a Ping source
type KnPingSourcesClient interface {

	// GetPingSource fetches a Ping source by its name
	GetPingSource(name string) (*v1alpha2.PingSource, error)

	// CreatePingSource creates a Ping source
	CreatePingSource(pingSource *v1alpha2.PingSource) error

	// UpdatePingSource updates a Ping source
	UpdatePingSource(pingSource *v1alpha2.PingSource) error

	// DeletePingSource deletes a Ping source
	DeletePingSource(name string) error

	// ListPingSource lists all Ping sources
	// TODO: Support list configs like in service list
	ListPingSource() (*v1alpha2.PingSourceList, error)

	// Get namespace for this source
	Namespace() string
}

// knSourcesClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type pingSourcesClient struct {
	client    clientv1alpha2.PingSourceInterface
	namespace string
}

// NewKnSourcesClient is to invoke Eventing Sources Client API to create object
func newKnPingSourcesClient(client clientv1alpha2.PingSourceInterface, namespace string) KnPingSourcesClient {
	return &pingSourcesClient{
		client:    client,
		namespace: namespace,
	}
}

// Get the namespace for which this client has been created
func (c *pingSourcesClient) Namespace() string {
	return c.namespace
}

func (c *pingSourcesClient) CreatePingSource(pingsource *v1alpha2.PingSource) error {
	if pingsource.Spec.Sink.Ref == nil && pingsource.Spec.Sink.URI == nil {
		return fmt.Errorf("a sink is required for creating a source")
	}
	_, err := c.client.Create(pingsource)
	return err
}

func (c *pingSourcesClient) UpdatePingSource(pingSource *v1alpha2.PingSource) error {
	_, err := c.client.Update(pingSource)
	return err
}

func (c *pingSourcesClient) DeletePingSource(name string) error {
	return c.client.Delete(name, &metav1.DeleteOptions{})
}

func (c *pingSourcesClient) GetPingSource(name string) (*v1alpha2.PingSource, error) {
	return c.client.Get(name, metav1.GetOptions{})
}

// ListPingSource returns the available Ping sources
func (c *pingSourcesClient) ListPingSource() (*v1alpha2.PingSourceList, error) {
	sourceList, err := c.client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return updatePingSourceListGVK(sourceList)
}

func updatePingSourceListGVK(sourceList *v1alpha2.PingSourceList) (*v1alpha2.PingSourceList, error) {
	sourceListNew := sourceList.DeepCopy()
	err := updateSourceGVK(sourceListNew)
	if err != nil {
		return nil, err
	}

	sourceListNew.Items = make([]v1alpha2.PingSource, len(sourceList.Items))
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
	pingSource *v1alpha2.PingSource
}

func NewPingSourceBuilder(name string) *PingSourceBuilder {
	return &PingSourceBuilder{pingSource: &v1alpha2.PingSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

func NewPingSourceBuilderFromExisting(pingsource *v1alpha2.PingSource) *PingSourceBuilder {
	return &PingSourceBuilder{pingSource: pingsource.DeepCopy()}
}

func (b *PingSourceBuilder) Schedule(schedule string) *PingSourceBuilder {
	b.pingSource.Spec.Schedule = schedule
	return b
}

func (b *PingSourceBuilder) JsonData(data string) *PingSourceBuilder {
	b.pingSource.Spec.JsonData = data
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

func (b *PingSourceBuilder) Build() *v1alpha2.PingSource {
	return b.pingSource
}
