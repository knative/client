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
	client_v1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"
)

// KnSourcesClient to Eventing Sources. All methods are relative to the
// namespace specified during construction
type KnSourcesClient interface {
	// Namespace in which this client is operating for
	Namespace() string
}

// Create a new client facade for the provided namespace
func NewKnSourcesClient(client client_v1alpha1.SourcesV1alpha1Interface, namespace string) KnSourcesClient {
	return &knSourcesClient{
		client:    client,
		namespace: namespace,
	}
}

type knSourcesClient struct {
	client    client_v1alpha1.SourcesV1alpha1Interface
	namespace string
}

// Return the client's namespace
func (cl *knSourcesClient) Namespace() string {
	return cl.namespace
}
