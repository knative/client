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

package v1alpha1

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1/fake"
)

var testNamespace = "test-ns"

func setup() (sources fake.FakeSourcesV1alpha1, client KnSourcesClient) {
	sources = fake.FakeSourcesV1alpha1{Fake: &client_testing.Fake{}}
	client = NewKnSourcesClient(&sources, testNamespace)
	return
}

func TestDeleteApiServerSource(t *testing.T) {
	var srcName = "new-src"
	sourcesServer, client := setup()

	apisourceNew := newApiServerSource(srcName)

	sourcesServer.AddReactor("create", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
			if name == apisourceNew.Name {
				apisourceNew.Generation = 2
				return true, apisourceNew, nil
			}
			return true, nil, fmt.Errorf("error while creating apiserversource %s", name)
		})

	t.Run("create apiserversource without error", func(t *testing.T) {
		ins, err := client.CreateApiServerSource(apisourceNew)
		assert.NilError(t, err)
		assert.Equal(t, ins.Name, srcName)
		assert.Equal(t, ins.Namespace, testNamespace)
	})

	t.Run("create apiserversource with an error returns an error object", func(t *testing.T) {
		_, err := client.CreateApiServerSource(newApiServerSource("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestCreateApiServerSource(t *testing.T) {
	var srcName = "new-src"
	sourcesServer, client := setup()

	apisourceNew := newApiServerSource(srcName)

	sourcesServer.AddReactor("create", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
			if name == apisourceNew.Name {
				apisourceNew.Generation = 2
				return true, apisourceNew, nil
			}
			return true, nil, fmt.Errorf("error while creating apiserversource %s", name)
		})

	t.Run("create apiserversource without error", func(t *testing.T) {
		ins, err := client.CreateApiServerSource(apisourceNew)
		assert.NilError(t, err)
		assert.Equal(t, ins.Name, srcName)
		assert.Equal(t, ins.Namespace, testNamespace)
	})

	t.Run("create apiserversource with an error returns an error object", func(t *testing.T) {
		_, err := client.CreateApiServerSource(newApiServerSource("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func newApiServerSource(name string) *v1alpha1.ApiServerSource {
	src := &v1alpha1.ApiServerSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
	}
	src.Name = name
	src.Namespace = testNamespace
	return src
}
