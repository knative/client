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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// gvrFromUnstructured takes a unstructured object of CRD type and finds GVR from its spec
func gvrFromUnstructured(u *unstructured.Unstructured) (gvr schema.GroupVersionResource, err error) {
	group, err := groupFromUnstructured(u)
	if err != nil {
		return gvr, err
	}

	version, err := versionFromUnstructured(u)
	if err != nil {
		return gvr, err
	}

	resource, err := resourceFromUnstructured(u)
	if err != nil {
		return gvr, err
	}

	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}, nil
}

func groupFromUnstructured(u *unstructured.Unstructured) (string, error) {
	content := u.UnstructuredContent()
	group, found, err := unstructured.NestedString(content, "spec", "group")
	if err != nil || !found {
		return "", fmt.Errorf("can't find group for source GVR: %v", err)
	}
	return group, nil
}

func versionFromUnstructured(u *unstructured.Unstructured) (version string, err error) {
	content := u.UnstructuredContent()
	versions, found, err := unstructured.NestedSlice(content, "spec", "versions")
	if err != nil || !found || len(versions) == 0 {
		// fallback to .spec.version
		version, found, err = unstructured.NestedString(content, "spec", "version")
		if err != nil || !found {
			return version, fmt.Errorf("can't find version for source GVR: %v", err)
		}
	} else {
		for _, v := range versions {
			if vmap, ok := v.(map[string]interface{}); ok {
				// find the version name which is being served
				if vmap["served"] == true {
					version = vmap["name"].(string)
					break
				}
			}
		}
	}
	// if we could find the version at all
	if version == "" {
		err = fmt.Errorf("can't find version for source GVR")
	}
	return version, err
}

func resourceFromUnstructured(u *unstructured.Unstructured) (string, error) {
	content := u.UnstructuredContent()
	resource, found, err := unstructured.NestedString(content, "spec", "names", "plural")
	if err != nil || !found {
		return "", fmt.Errorf("can't find resource for source GVR: %v", err)
	}
	return resource, nil
}

func kindFromUnstructured(u *unstructured.Unstructured) (string, error) {
	content := u.UnstructuredContent()
	kind, found, err := unstructured.NestedString(content, "spec", "names", "kind")
	if !found || err != nil {
		return "", fmt.Errorf("can't find source kind from source CRD: %v", err)
	}
	return kind, nil
}

// TypesFilter for keeping list of sources types to filter upo
type TypesFilter []string

// WithType function for easy filtering on source types
type WithType func(filters *TypesFilter)

// WithTypes for recording the source type filtering function WithType
type WithTypes []WithType

// WithTypeFilter can be used to filter based on source type name
func WithTypeFilter(name string) WithType {
	return func(filters *TypesFilter) {
		*filters = append(*filters, name)
	}
}

// List returns the source type name list recorded via WithTypeFilter
func (types WithTypes) List() []string {
	var stypes TypesFilter
	for _, f := range types {
		f(&stypes)
	}
	return stypes
}

// UnstructuredCRDFromGVK constructs an unstructured object using the given GVK
func UnstructuredCRDFromGVK(gvk schema.GroupVersionKind) *unstructured.Unstructured {
	name := fmt.Sprintf("%ss.%s", strings.ToLower(gvk.Kind), gvk.Group)
	plural := fmt.Sprintf("%ss", strings.ToLower(gvk.Kind))
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": map[string]interface{}{
			"group":   gvk.Group,
			"version": gvk.Version,
			"names": map[string]interface{}{
				"kind":   gvk.Kind,
				"plural": plural,
			},
		},
	})

	return u
}
