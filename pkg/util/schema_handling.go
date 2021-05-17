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

package util

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func UpdateGroupVersionKindWithScheme(obj runtime.Object, gv schema.GroupVersion, scheme *runtime.Scheme) error {
	gvk, err := GetGroupVersionKind(obj, gv, scheme)
	if err != nil {
		return err
	}
	obj.GetObjectKind().SetGroupVersionKind(*gvk)
	return nil
}

func GetGroupVersionKind(obj runtime.Object, gv schema.GroupVersion, scheme *runtime.Scheme) (*schema.GroupVersionKind, error) {
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	for _, gvk := range gvks {
		if gvk.GroupVersion() == gv {
			return &gvk, nil
		}
	}
	return nil, fmt.Errorf("no group version %s registered in %s", gv, scheme.Name())
}
