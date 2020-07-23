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
	"k8s.io/apimachinery/pkg/runtime/schema"
	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	clientv1alpha2 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2"
)

// KnSinkBindingClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnSourcesClient interface {
	// Get client for Ping sources
	PingSourcesClient() KnPingSourcesClient

	// Get client for sink binding sources
	SinkBindingClient() KnSinkBindingClient

	// Get client for ApiServer sources
	APIServerSourcesClient() KnAPIServerSourcesClient
}

// sourcesClient is a combination of Sources client interface and namespace
// Temporarily help to add sources dependencies
// May be changed when adding real sources features
type sourcesClient struct {
	client    clientv1alpha2.SourcesV1alpha2Interface
	namespace string
}

// NewKnSourcesClient for managing all eventing built-in sources
func NewKnSourcesClient(client clientv1alpha2.SourcesV1alpha2Interface, namespace string) KnSourcesClient {
	return &sourcesClient{
		client:    client,
		namespace: namespace,
	}
}

// Get the client for dealing with Ping sources
func (c *sourcesClient) PingSourcesClient() KnPingSourcesClient {
	return newKnPingSourcesClient(c.client.PingSources(c.namespace), c.namespace)
}

// ApiServerSourcesClient for dealing with ApiServer sources
func (c *sourcesClient) SinkBindingClient() KnSinkBindingClient {
	return newKnSinkBindingClient(c.client.SinkBindings(c.namespace), c.namespace)
}

// ApiServerSourcesClient for dealing with ApiServer sources
func (c *sourcesClient) APIServerSourcesClient() KnAPIServerSourcesClient {
	return newKnAPIServerSourcesClient(c.client.ApiServerSources(c.namespace), c.namespace)
}

// BuiltInSourcesGVKs returns the GVKs for built in sources
func BuiltInSourcesGVKs() []schema.GroupVersionKind {
	return []schema.GroupVersionKind{
		sourcesv1alpha2.SchemeGroupVersion.WithKind("ApiServerSource"),
		sourcesv1alpha2.SchemeGroupVersion.WithKind("ContainerSource"),
		sourcesv1alpha2.SchemeGroupVersion.WithKind("PingSource"),
		sourcesv1alpha2.SchemeGroupVersion.WithKind("SinkBinding"),
	}
}
