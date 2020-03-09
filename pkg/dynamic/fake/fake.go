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

package fake

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/dynamic"
)

// CreateFakeKnDynamicClient gives you a dynamic client for testing containing the given objects.
func CreateFakeKnDynamicClient(testNamespace string, objects ...runtime.Object) dynamic.KnDynamicClient {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1", Kind: "Service"}, &servingv1.Service{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "eventing.knative.dev", Version: "v1alpha1", Kind: "Broker"}, &eventingv1alpha1.Broker{})
	client := dynamicfake.NewSimpleDynamicClient(scheme, objects...)
	return dynamic.NewKnDynamicClient(client, testNamespace)
}
