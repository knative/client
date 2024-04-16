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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client-pkg/pkg/dynamic"

	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	sourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"
	dynamicclientfake "knative.dev/pkg/injection/clients/dynamicclient/fake"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// CreateFakeKnDynamicClient gives you a dynamic client for testing containing the given objects.
func CreateFakeKnDynamicClient(testNamespace string, objects ...runtime.Object) dynamic.KnDynamicClient {
	scheme := runtime.NewScheme()
	servingv1.AddToScheme(scheme)
	eventingv1.AddToScheme(scheme)
	messagingv1.AddToScheme(scheme)
	sourcesv1.AddToScheme(scheme)
	sourcesv1beta2.AddToScheme(scheme)
	apiextensionsv1.AddToScheme(scheme)
	_, dynamicClient := dynamicclientfake.With(context.TODO(), scheme, objects...)
	return dynamic.NewKnDynamicClient(dynamicClient, testNamespace)
}
