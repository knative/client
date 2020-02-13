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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	crdGroup          = "apiextensions.k8s.io"
	crdVersion        = "v1beta1"
	crdKind           = "CustomResourceDefinition"
	crdKinds          = "customresourcedefinitions"
	sourcesLabelKey   = "duck.knative.dev/source"
	sourcesLabelValue = "true"
)

// SourceListFilters defines flags used for kn source list to filter sources on types
type SourceListFilters struct {
	filters []string
}

// Add attaches the SourceListFilters flags to given command
func (s *SourceListFilters) Add(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&s.filters, "type", "t", nil, "Filter list on given source type. This flag can be given multiple times.")
}

// KnDynamicClient to client-go Dynamic client. All methods are relative to the
// namespace specified during construction
type KnDynamicClient interface {
	// Namespace in which this client is operating for
	Namespace() string

	// ListCRDs returns list of CRDs with their type and name
	ListCRDs(options metav1.ListOptions) (*unstructured.UnstructuredList, error)

	// ListSourceCRDs returns list of eventing sources CRDs
	ListSourcesTypes() (*unstructured.UnstructuredList, error)

	// ListSources returns list of available sources COs
	ListSources(f *SourceListFilters) (*unstructured.UnstructuredList, error)

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

// ListSources returns list of available sources COs
func (c *knDynamicClient) ListSources(f *SourceListFilters) (*unstructured.UnstructuredList, error) {
	var sourceList unstructured.UnstructuredList
	options := metav1.ListOptions{}
	sourceTypes, err := c.ListSourcesTypes()
	if err != nil {
		return nil, err
	}

	namespace := c.Namespace()
	// For each source type available, find out CO
	for _, source := range sourceTypes.Items {
		//  only find COs if this source type is given in filter
		if f != nil && f.filters != nil {
			// find source kind before hand to fail early
			sourceKind, err := kindFromUnstructured(&source)
			if err != nil {
				return nil, err
			}

			// if this source is not given in filter flags continue
			if !sliceContainsIgnoreCase(sourceKind, f.filters) {
				continue
			}
		}

		gvr, err := gvrFromUnstructured(&source)
		if err != nil {
			return nil, err
		}

		sList, err := c.client.Resource(gvr).Namespace(namespace).List(options)
		if err != nil {
			return nil, err
		}

		if len(sList.Items) > 0 {
			sourceList.Items = append(sourceList.Items, sList.Items...)
		}
		sourceList.SetGroupVersionKind(sList.GetObjectKind().GroupVersionKind())
	}
	return &sourceList, nil
}

func gvrFromUnstructured(u *unstructured.Unstructured) (gvr schema.GroupVersionResource, err error) {
	content := u.UnstructuredContent()
	group, found, err := unstructured.NestedString(content, "spec", "group")
	if err != nil || !found {
		return gvr, fmt.Errorf("can't find source GVR: %v", err)
	}
	version, found, err := unstructured.NestedString(content, "spec", "version")
	if err != nil || !found {
		return gvr, fmt.Errorf("can't find source GVR: %v", err)
	}
	resource, found, err := unstructured.NestedString(content, "spec", "names", "plural")
	if err != nil || !found {
		return gvr, fmt.Errorf("can't find source GVR: %v", err)
	}
	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}, nil
}

func kindFromUnstructured(u *unstructured.Unstructured) (string, error) {
	content := u.UnstructuredContent()
	kind, found, err := unstructured.NestedString(content, "spec", "names", "kind")
	if !found || err != nil {
		return "", fmt.Errorf("can't find source kind: %v", err)
	}
	return kind, nil
}

func sliceContainsIgnoreCase(s string, slice []string) bool {
	for _, each := range slice {
		if strings.EqualFold(s, each) {
			return true
		}
	}
	return false
}
