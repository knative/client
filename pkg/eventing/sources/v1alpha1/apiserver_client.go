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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kn_errors "knative.dev/client/pkg/errors"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	client_v1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

// Interface for working with ApiServer sources
type KnApiServerSourcesClient interface {

	// Get an ApiServerSource by name
	GetApiServerSource(name string) (*v1alpha1.ApiServerSource, error)

	// Create an ApiServerSource by object
	CreateApiServerSource(apiSource *v1alpha1.ApiServerSource) error

	// Update an ApiServerSource by object
	UpdateApiServerSource(apiSource *v1alpha1.ApiServerSource) error

	// Delete an ApiServerSource by name
	DeleteApiServerSource(name string) error

	// Get namespace for this client
	Namespace() string
}

// knSourcesClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type apiServerSourcesClient struct {
	client    client_v1alpha1.ApiServerSourceInterface
	namespace string
}

// NewKnSourcesClient is to invoke Eventing Sources Client API to create object
func newKnApiServerSourcesClient(client client_v1alpha1.ApiServerSourceInterface, namespace string) KnApiServerSourcesClient {
	return &apiServerSourcesClient{
		client:    client,
		namespace: namespace,
	}
}

//GetApiServerSource returns apiSource object if present
func (c *apiServerSourcesClient) GetApiServerSource(name string) (*v1alpha1.ApiServerSource, error) {
	apiSource, err := c.client.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, kn_errors.GetError(err)
	}

	return apiSource, nil
}

//CreateApiServerSource is used to create an instance of ApiServerSource
func (c *apiServerSourcesClient) CreateApiServerSource(apiSource *v1alpha1.ApiServerSource) error {
	_, err := c.client.Create(apiSource)
	if err != nil {
		return kn_errors.GetError(err)
	}

	return nil
}

//UpdateApiServerSource is used to update an instance of ApiServerSource
func (c *apiServerSourcesClient) UpdateApiServerSource(apiSource *v1alpha1.ApiServerSource) error {
	_, err := c.client.Update(apiSource)
	if err != nil {
		return kn_errors.GetError(err)
	}

	return nil
}

//DeleteApiServerSource is used to create an instance of ApiServerSource
func (c *apiServerSourcesClient) DeleteApiServerSource(name string) error {
	err := c.client.Delete(name, &metav1.DeleteOptions{})
	return err
}

// Return the client's namespace
func (c *apiServerSourcesClient) Namespace() string {
	return c.namespace
}

// APIServerSourceBuilder is for building the source
type APIServerSourceBuilder struct {
	apiServerSource *v1alpha1.ApiServerSource
}

func NewAPIServerSourceBuilder(name string) *APIServerSourceBuilder {
	return &APIServerSourceBuilder{apiServerSource: &v1alpha1.ApiServerSource{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}}
}

func NewAPIServerSourceBuilderFromExisting(apiServerSource *v1alpha1.ApiServerSource) *APIServerSourceBuilder {
	return &APIServerSourceBuilder{apiServerSource: apiServerSource.DeepCopy()}
}

func (b *APIServerSourceBuilder) Resources(resources []v1alpha1.ApiServerResource) *APIServerSourceBuilder {
	b.apiServerSource.Spec.Resources = resources
	return b
}

func (b *APIServerSourceBuilder) ServiceAccount(sa string) *APIServerSourceBuilder {
	b.apiServerSource.Spec.ServiceAccountName = sa
	return b
}

func (b *APIServerSourceBuilder) Mode(mode string) *APIServerSourceBuilder {
	b.apiServerSource.Spec.Mode = mode
	return b
}

func (b *APIServerSourceBuilder) Sink(sink *duckv1beta1.Destination) *APIServerSourceBuilder {
	b.apiServerSource.Spec.Sink = sink
	return b
}

func (b *APIServerSourceBuilder) Build() *v1alpha1.ApiServerSource {
	return b.apiServerSource
}
