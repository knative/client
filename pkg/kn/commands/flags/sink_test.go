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

package flags

import (
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
)

type resolveCase struct {
	sink        string
	destination *duckv1.Destination
	errContents string
}

func TestResolve(t *testing.T) {
	targetExampleCom, err := apis.ParseURL("http://target.example.com")
	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	defaultBroker := &eventingv1alpha1.Broker{
		TypeMeta:   metav1.TypeMeta{Kind: "Broker", APIVersion: "eventing.knative.dev/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "default"},
	}

	assert.NilError(t, err)
	cases := []resolveCase{
		{"svc:mysvc", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Service",
				APIVersion: "serving.knative.dev/v1",
				Namespace:  "default",
				Name:       "mysvc"}}, ""},
		{"service:mysvc", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Service",
				APIVersion: "serving.knative.dev/v1",
				Namespace:  "default",
				Name:       "mysvc"}}, ""},
		{"svc:absent", nil, "\"absent\" not found"},
		{"broker:default", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Broker",
				APIVersion: "eventing.knative.dev/v1alpha1",
				Namespace:  "default",
				Name:       "default"}}, ""},
		{"http://target.example.com", &duckv1.Destination{
			URI: targetExampleCom,
		}, ""},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc, defaultBroker)
	for _, c := range cases {
		i := &SinkFlags{c.sink}
		result, err := i.ResolveSink(dynamicClient, "default")
		if c.destination != nil {
			assert.DeepEqual(t, result, c.destination)
			assert.NilError(t, err)
		} else {
			assert.ErrorContains(t, err, c.errContents)
		}
	}
}
