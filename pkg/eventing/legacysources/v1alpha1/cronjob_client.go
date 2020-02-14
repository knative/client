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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"

	"knative.dev/client/pkg/util"
	"knative.dev/eventing/pkg/legacyclient/clientset/versioned/scheme"
	client_v1alpha1 "knative.dev/eventing/pkg/legacyclient/clientset/versioned/typed/legacysources/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// Interface for interacting with a cronjob source
type KnCronJobSourcesClient interface {

	// Get a single cronjob source by name
	GetCronJobSource(name string) (*v1alpha1.CronJobSource, error)

	// Create a cronjob source by providing the schedule, data and sink
	CreateCronJobSource(cronjobSource *v1alpha1.CronJobSource) error

	// Update a cronjob source by providing the schedule, data and sink
	UpdateCronJobSource(cronjobSource *v1alpha1.CronJobSource) error

	// Delete a cronjob source by name
	DeleteCronJobSource(name string) error

	// List CronJob sources
	// TODO: Support list configs like in service list
	ListCronJobSource() (*v1alpha1.CronJobSourceList, error)

	// Get namespace for this source
	Namespace() string
}

// knSourcesClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type cronJobSourcesClient struct {
	client    client_v1alpha1.CronJobSourceInterface
	namespace string
}

// NewKnSourcesClient is to invoke Eventing Sources Client API to create object
func newKnCronJobSourcesClient(client client_v1alpha1.CronJobSourceInterface, namespace string) KnCronJobSourcesClient {
	return &cronJobSourcesClient{
		client:    client,
		namespace: namespace,
	}
}

// Get the namespace for which this client has been created
func (c *cronJobSourcesClient) Namespace() string {
	return c.namespace
}

func (c *cronJobSourcesClient) CreateCronJobSource(cronjobSource *v1alpha1.CronJobSource) error {
	if cronjobSource.Spec.Sink == nil {
		return fmt.Errorf("a sink is required for creating a source")
	}
	_, err := c.client.Create(cronjobSource)
	return err
}

func (c *cronJobSourcesClient) UpdateCronJobSource(cronjobSource *v1alpha1.CronJobSource) error {
	_, err := c.client.Update(cronjobSource)
	return err
}

func (c *cronJobSourcesClient) DeleteCronJobSource(name string) error {
	return c.client.Delete(name, &metav1.DeleteOptions{})
}

func (c *cronJobSourcesClient) GetCronJobSource(name string) (*v1alpha1.CronJobSource, error) {
	return c.client.Get(name, metav1.GetOptions{})
}

// ListCronJobSource returns the available CronJob type sources
func (c *cronJobSourcesClient) ListCronJobSource() (*v1alpha1.CronJobSourceList, error) {
	sourceList, err := c.client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return updateCronJobSourceListGVK(sourceList)
}

func updateCronJobSourceListGVK(sourceList *v1alpha1.CronJobSourceList) (*v1alpha1.CronJobSourceList, error) {
	sourceListNew := sourceList.DeepCopy()
	err := updateSourceGVK(sourceListNew)
	if err != nil {
		return nil, err
	}

	sourceListNew.Items = make([]v1alpha1.CronJobSource, len(sourceList.Items))
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

func updateSourceGVK(obj runtime.Object) error {
	return util.UpdateGroupVersionKindWithScheme(obj, v1alpha1.SchemeGroupVersion, scheme.Scheme)
}

// Builder for building up cronjob sources

type CronJobSourceBuilder struct {
	cronjobSource *v1alpha1.CronJobSource
}

func NewCronJobSourceBuilder(name string) *CronJobSourceBuilder {
	return &CronJobSourceBuilder{cronjobSource: &v1alpha1.CronJobSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

func NewCronJobSourceBuilderFromExisting(cronjobsource *v1alpha1.CronJobSource) *CronJobSourceBuilder {
	return &CronJobSourceBuilder{cronjobSource: cronjobsource.DeepCopy()}
}

func (b *CronJobSourceBuilder) Schedule(schedule string) *CronJobSourceBuilder {
	b.cronjobSource.Spec.Schedule = schedule
	return b
}

func (b *CronJobSourceBuilder) Data(data string) *CronJobSourceBuilder {
	b.cronjobSource.Spec.Data = data
	return b
}

func (b *CronJobSourceBuilder) Sink(sink *duckv1beta1.Destination) *CronJobSourceBuilder {
	b.cronjobSource.Spec.Sink = sink
	return b
}

func (b *CronJobSourceBuilder) ServiceAccount(sa string) *CronJobSourceBuilder {
	b.cronjobSource.Spec.ServiceAccountName = sa
	return b
}

func (b *CronJobSourceBuilder) ResourceRequestsCPU(cpu string) *CronJobSourceBuilder {
	b.cronjobSource.Spec.Resources.Requests.ResourceCPU = cpu
	return b
}

func (b *CronJobSourceBuilder) ResourceRequestsMemory(memory string) *CronJobSourceBuilder {
	b.cronjobSource.Spec.Resources.Requests.ResourceMemory = memory
	return b
}

func (b *CronJobSourceBuilder) ResourceLimitsCPU(cpu string) *CronJobSourceBuilder {
	b.cronjobSource.Spec.Resources.Limits.ResourceCPU = cpu
	return b
}

func (b *CronJobSourceBuilder) ResourceLimitsMemory(memory string) *CronJobSourceBuilder {
	b.cronjobSource.Spec.Resources.Limits.ResourceMemory = memory
	return b
}

func (b *CronJobSourceBuilder) Build() *v1alpha1.CronJobSource {
	return b.cronjobSource
}
