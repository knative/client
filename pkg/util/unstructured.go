// Copyright © 2020 The Knative Authors
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

package util

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func ToUnstructuredList(obj runtime.Object) (*unstructured.UnstructuredList, error) {
	unstructuredList := &unstructured.UnstructuredList{}
	unstructuredList.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
	if meta.IsListType(obj) {
		items, err := meta.ExtractList(obj)
		if err != nil {
			return nil, err
		}
		for _, obj := range items {
			ud, err := toUnstructured(obj)
			if err != nil {
				return nil, err
			}
			unstructuredList.Items = append(unstructuredList.Items, *ud)
		}

	} else {

	}
	return unstructuredList, nil

}

func toUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	ud := &unstructured.Unstructured{}
	if err := json.Unmarshal(b, ud); err != nil {
		return nil, err
	}
	return ud, nil
}
