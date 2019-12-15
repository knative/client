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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	client_v1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"

	kn_errors "knative.dev/client/pkg/errors"
)

// Interface for working with ApiServer sources
type KnApiServerSourcesClient interface {

	// Get an ApiServerSource by object
	CreateApiServerSource(apisvrsrc *v1alpha1.ApiServerSource) error

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

//CreateApiServerSource is used to create an instance of ApiServerSource
func (c *apiServerSourcesClient) CreateApiServerSource(apisvrsrc *v1alpha1.ApiServerSource) error {
	_, err := c.client.Create(apisvrsrc)
	if err != nil {
		return kn_errors.GetError(err)
	}

	return nil
}

//DeleteApiServerSource is used to create an instance of ApiServerSource
func (c *apiServerSourcesClient) DeleteApiServerSource(name string) error {
	err := c.client.Delete(name, &v1.DeleteOptions{})
	return err
}

// Return the client's namespace
func (c *apiServerSourcesClient) Namespace() string {
	return c.namespace
}
