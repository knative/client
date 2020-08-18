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

package dynamic

import (
	"errors"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"knative.dev/client/pkg/util"
)

const (
	crdGroup          = "apiextensions.k8s.io"
	crdVersion        = "v1beta1"
	crdKind           = "CustomResourceDefinition"
	crdKinds          = "customresourcedefinitions"
	sourcesLabelKey   = "duck.knative.dev/source"
	sourcesLabelValue = "true"
	sourceListGroup   = "client.knative.dev"
	sourceListVersion = "v1alpha1"
	sourceListKind    = "SourceList"
)

// KnDynamicClient to client-go Dynamic client. All methods are relative to the
// namespace specified during construction
type KnDynamicClient interface {
	// Namespace in which this client is operating for
	Namespace() string

	// ListCRDs returns list of CRDs with their type and name
	ListCRDs(options metav1.ListOptions) (*unstructured.UnstructuredList, error)

	// ListSourcesTypes returns list of eventing sources CRDs
	ListSourcesTypes() (*unstructured.UnstructuredList, error)

	// ListSources returns list of available source objects
	ListSources(types ...WithType) (*unstructured.UnstructuredList, error)

	// ListSourcesUsingGVKs returns list of available source objects using given list of GVKs
	ListSourcesUsingGVKs(*[]schema.GroupVersionKind, ...WithType) (*unstructured.UnstructuredList, error)

	// RawClient returns the raw dynamic client interface
	RawClient() dynamic.Interface
}

// knDynamicClient is a combination of client-go Dynamic client interface and namespace
type knDynamicClient struct {
	client    dynamic.Interface
	namespace string
}

// NewKnDynamicClient is to invoke Eventing Sources Client API to create object
func NewKnDynamicClient(client dynamic.Interface, namespace string) KnDynamicClient {
	return &knDynamicClient{
		client:    client,
		namespace: namespace,
	}
}

// Return the client's namespace
func (c *knDynamicClient) Namespace() string {
	return c.namespace
}

// TODO(navidshaikh): Use ListConfigs here instead of ListOptions
// ListCRDs returns list of installed CRDs in the cluster and filters based on the given options
func (c *knDynamicClient) ListCRDs(options metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	gvr := schema.GroupVersionResource{
		Group:    crdGroup,
		Version:  crdVersion,
		Resource: crdKinds,
	}

	uList, err := c.client.Resource(gvr).List(options)
	if err != nil {
		return nil, err
	}

	return uList, nil
}

// ListSourcesTypes returns installed knative eventing sources CRDs
func (c *knDynamicClient) ListSourcesTypes() (*unstructured.UnstructuredList, error) {
	options := metav1.ListOptions{}
	sourcesLabels := labels.Set{sourcesLabelKey: sourcesLabelValue}
	options.LabelSelector = sourcesLabels.String()
	return c.ListCRDs(options)
}

func (c knDynamicClient) RawClient() dynamic.Interface {
	return c.client
}

// ListSources returns list of available sources objects
// Provide the list of source types as for example: WithTypes("pingsource", "apiserversource"...) to list
// only given types of source objects
func (c *knDynamicClient) ListSources(types ...WithType) (*unstructured.UnstructuredList, error) {
	var (
		sourceList unstructured.UnstructuredList
		options    metav1.ListOptions
	)
	sourceTypes, err := c.ListSourcesTypes()
	if err != nil {
		return nil, err
	}

	if sourceTypes == nil || len(sourceTypes.Items) == 0 {
		return nil, errors.New("no sources found on the backend, please verify the installation")
	}

	namespace := c.Namespace()
	filters := WithTypes(types).List()
	// For each source type available, find out each source types objects
	for _, source := range sourceTypes.Items {
		// find source kind before hand to fail early
		sourceKind, err := kindFromUnstructured(&source)
		if err != nil {
			return nil, err
		}

		if len(filters) > 0 && !util.SliceContainsIgnoreCase(filters, sourceKind) {
			continue
		}

		// find source's GVR from unstructured source type object
		gvr, err := gvrFromUnstructured(&source)
		if err != nil {
			return nil, err
		}

		// list objects of source type with this GVR
		sList, err := c.client.Resource(gvr).Namespace(namespace).List(options)
		if err != nil {
			return nil, err
		}

		if len(sList.Items) > 0 {
			sourceList.Items = append(sourceList.Items, sList.Items...)
		}
	}
	if len(sourceList.Items) > 0 {
		sourceList.SetGroupVersionKind(schema.GroupVersionKind{Group: sourceListGroup, Version: sourceListVersion, Kind: sourceListKind})
	}
	return &sourceList, nil
}

// ListSourcesUsingGVKs returns list of available source objects using given list of GVKs
func (c *knDynamicClient) ListSourcesUsingGVKs(gvks *[]schema.GroupVersionKind, types ...WithType) (*unstructured.UnstructuredList, error) {
	if gvks == nil {
		return nil, nil
	}

	var (
		sourceList unstructured.UnstructuredList
		options    metav1.ListOptions
	)
	namespace := c.Namespace()
	filters := WithTypes(types).List()

	for _, gvk := range *gvks {
		if len(filters) > 0 && !util.SliceContainsIgnoreCase(filters, gvk.Kind) {
			continue
		}

		gvr := gvk.GroupVersion().WithResource(strings.ToLower(gvk.Kind) + "s")

		// list objects of source type with this GVR
		sList, err := c.client.Resource(gvr).Namespace(namespace).List(options)
		if err != nil {
			return nil, err
		}

		if len(sList.Items) > 0 {
			sourceList.Items = append(sourceList.Items, sList.Items...)
		}
	}
	if len(sourceList.Items) > 0 {
		sourceList.SetGroupVersionKind(schema.GroupVersionKind{Group: sourceListGroup, Version: sourceListVersion, Kind: sourceListKind})
	}
	return &sourceList, nil
}
