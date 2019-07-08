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

package serving

import (
	"errors"
	"fmt"

	"github.com/knative/serving/pkg/client/clientset/versioned/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Update the GVK on the given object, based on the GVK registered in into the serving scheme
// for the given GroupVersion
func UpdateGroupVersionKind(obj runtime.Object, gv schema.GroupVersion) error {
	gvk, err := GetGroupVersionKind(obj, gv)
	if err != nil {
		return err
	}
	obj.GetObjectKind().SetGroupVersionKind(*gvk)
	return nil
}

func GetGroupVersionKind(obj runtime.Object, gv schema.GroupVersion) (*schema.GroupVersionKind, error) {
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	for _, gvk := range gvks {
		if gvk.GroupVersion() == gv {
			return &gvk, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("no group version %s registered in %s", gv, scheme.Scheme.Name()))
}
