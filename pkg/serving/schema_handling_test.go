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
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestGVKUpdate(t *testing.T) {
	service := v1alpha1.Service{}
	err := UpdateGroupVersionKind(&service, v1alpha1.SchemeGroupVersion)
	if err != nil {
		t.Fatalf("cannot update GVK to a service %v", err)
	}
	if service.Kind != "Service" {
		t.Fatalf("wrong kind '%s'", service.Kind)
	}
	if service.APIVersion != v1alpha1.SchemeGroupVersion.Group+"/"+v1alpha1.SchemeGroupVersion.Version {
		t.Fatalf("wrong version '%s'", service.APIVersion)
	}
}

func TestGVKUpdateNegative(t *testing.T) {
	service := v1alpha1.Service{}
	err := UpdateGroupVersionKind(&service, schema.GroupVersion{Group: "bla", Version: "blub"})
	if err == nil {
		t.Fatal("expect an error for an unregistered group version")
	}
}
