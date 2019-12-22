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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamic_fake "knative.dev/client/pkg/dynamic/fake"
	eventing_v1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

type resolveCase struct {
	sink        string
	destination *duckv1beta1.Destination
	errContents string
}

func TestResolve(t *testing.T) {
	targetExampleCom, err := apis.ParseURL("http://target.example.com")
	mysvc := &serving_v1alpha1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	defaultBroker := &eventing_v1alpha1.Broker{
		TypeMeta:   metav1.TypeMeta{Kind: "Broker", APIVersion: "eventing.knative.dev/v1alpha1"},
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "default"},
	}

	assert.NilError(t, err)
	cases := []resolveCase{
		{"svc:mysvc", &duckv1beta1.Destination{
			Ref: &v1.ObjectReference{Kind: "Service",
				APIVersion: "serving.knative.dev/v1alpha1",
				Name:       "mysvc",
				Namespace:  "default"}}, ""},
		{"svc:absent", nil, "\"absent\" not found"},
		{"broker:default", &duckv1beta1.Destination{
			Ref: &v1.ObjectReference{Kind: "Broker",
				APIVersion: "eventing.knative.dev/v1alpha1",
				Name:       "default",
				Namespace:  "default",
			}}, ""},
		{"http://target.example.com", &duckv1beta1.Destination{
			URI: targetExampleCom,
		}, ""},
	}
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default", mysvc, defaultBroker)
	for _, c := range cases {
		i := &SinkFlags{c.sink}
		result, err := i.ResolveSink(dynamicClient.RawClient(), "default")
		if c.destination != nil {
			assert.DeepEqual(t, result, c.destination)
			assert.NilError(t, err)
		} else {
			assert.ErrorContains(t, err, c.errContents)
		}
	}
}
