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
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client/pkg/dynamic"

	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	dynamicclientfake "knative.dev/pkg/injection/clients/dynamicclient/fake"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// CreateFakeKnDynamicClient gives you a dynamic client for testing containing the given objects.
func CreateFakeKnDynamicClient(testNamespace string, objects ...runtime.Object) dynamic.KnDynamicClient {
	if !testing.Testing() {
		panic("For test usage only!")
	}
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = servingv1.AddToScheme(scheme)
	_ = eventingv1.AddToScheme(scheme)
	_ = messagingv1.AddToScheme(scheme)
	_ = sourcesv1.AddToScheme(scheme)
	_ = apiextensionsv1.AddToScheme(scheme)
	_, dynamicClient := dynamicclientfake.With(context.TODO(), scheme, objects...)
	return dynamic.NewKnDynamicClient(dynamicClient, testNamespace)
}
