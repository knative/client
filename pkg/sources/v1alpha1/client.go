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
	clientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"
)

// KnSinkBindingClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnSourcesClient interface {
	// Get client for ApiServer sources
	APIServerSourcesClient() KnAPIServerSourcesClient
}

// sourcesClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type sourcesClient struct {
	client    clientv1alpha1.SourcesV1alpha1Interface
	namespace string
}

// NewKnSourcesClient for managing all eventing built-in sources
func NewKnSourcesClient(client clientv1alpha1.SourcesV1alpha1Interface, namespace string) KnSourcesClient {
	return &sourcesClient{
		client:    client,
		namespace: namespace,
	}
}

// ApiServerSourcesClient for dealing with ApiServer sources
func (c *sourcesClient) APIServerSourcesClient() KnAPIServerSourcesClient {
	return newKnAPIServerSourcesClient(c.client.ApiServerSources(c.namespace), c.namespace)
}
