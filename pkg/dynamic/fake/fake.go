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
	k8s_fake "k8s.io/client-go/dynamic/fake"

	"knative.dev/client/pkg/dynamic"
	eventing_v1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// CreateFakeKnDynamicClient gives you a dynamic client for testing contianing the given objects.
func CreateFakeKnDynamicClient(testNamespace string, objects ...runtime.Object) dynamic.KnDynamicClient {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1alpha1", Kind: "Service"}, &serving_v1alpha1.Service{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "eventing.knative.dev", Version: "v1alpha1", Kind: "Broker"}, &eventing_v1alpha1.Broker{})
	client := k8s_fake.NewSimpleDynamicClient(scheme, objects...)
	return dynamic.NewKnDynamicClient(client, testNamespace)
}
