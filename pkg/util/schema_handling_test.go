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
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"
)

func TestGVKUpdate(t *testing.T) {
	service := servingv1.Service{}
	err := UpdateGroupVersionKindWithScheme(&service, servingv1.SchemeGroupVersion, scheme.Scheme)
	if err != nil {
		t.Fatalf("cannot update GVK to a service %v", err)
	}
	if service.Kind != "Service" {
		t.Fatalf("wrong kind '%s'", service.Kind)
	}
	if service.APIVersion != servingv1.SchemeGroupVersion.Group+"/"+servingv1.SchemeGroupVersion.Version {
		t.Fatalf("wrong version '%s'", service.APIVersion)
	}
}

func TestGVKUpdateNegative(t *testing.T) {
	service := servingv1.Service{}
	err := UpdateGroupVersionKindWithScheme(&service, schema.GroupVersion{Group: "bla", Version: "blub"}, scheme.Scheme)
	if err == nil {
		t.Fatal("expect an error for an unregistered group version")
	}
}
