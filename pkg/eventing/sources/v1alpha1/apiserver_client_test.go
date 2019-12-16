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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1/fake"
	"knative.dev/pkg/apis/duck/v1beta1"
)

var testApiServerSourceNamespace = "test-ns"

func setupApiServerSourcesClient(t *testing.T) (fakeSources fake.FakeSourcesV1alpha1, client KnApiServerSourcesClient) {
	fakeSources = fake.FakeSourcesV1alpha1{Fake: &client_testing.Fake{}}
	client = NewKnSourcesClient(&fakeSources, testApiServerSourceNamespace).ApiServerSourcesClient()
	assert.Equal(t, client.Namespace(), testApiServerSourceNamespace)
	return
}

func TestDeleteApiServerSource(t *testing.T) {
	sourcesServer, client := setupApiServerSourcesClient(t)

	sourcesServer.AddReactor("delete", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while deleting ApiServer source %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteApiServerSource("foo")
	assert.NilError(t, err)

	err = client.DeleteApiServerSource("errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestCreateApiServerSource(t *testing.T) {
	sourcesServer, client := setupApiServerSourcesClient(t)

	sourcesServer.AddReactor("create", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			newSource := a.(client_testing.CreateAction).GetObject()
			name := newSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while creating ApiServer source %s", name)
			}
			return true, newSource, nil
		})
	err := client.CreateApiServerSource(newApiServerSource("foo", "Event"))
	assert.NilError(t, err)

	err = client.CreateApiServerSource(newApiServerSource("errorSource", "Event"))
	assert.ErrorContains(t, err, "errorSource")

}

func TestGetApiServerSource(t *testing.T) {
	sourcesServer, client := setupApiServerSourcesClient(t)

	sourcesServer.AddReactor("get", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while getting ApiServer source %s", name)
			}
			return true, newApiServerSource(name, "Event"), nil
		})
	testsource, err := client.GetApiServerSource("foo")
	assert.NilError(t, err)
	assert.Equal(t, testsource.Name, "foo")
	assert.Equal(t, testsource.Spec.Sink.Ref.Name, "foosvc")

	_, err = client.GetApiServerSource("errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestUpdateApiServerSource(t *testing.T) {
	sourcesServer, client := setupApiServerSourcesClient(t)

	sourcesServer.AddReactor("update", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			updatedSource := a.(client_testing.UpdateAction).GetObject()
			name := updatedSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating ApiServer source %s", name)
			}
			return true, NewAPIServerSourceBuilderFromExisting(updatedSource.(*v1alpha1.ApiServerSource)).Build(), nil
		})
	err := client.UpdateApiServerSource(newApiServerSource("foo", "Event"))
	assert.NilError(t, err)

	err = client.UpdateApiServerSource(newApiServerSource("errorSource", "Event"))
	assert.ErrorContains(t, err, "errorSource")
}

func newApiServerSource(name, resource string) *v1alpha1.ApiServerSource {
	b := NewAPIServerSourceBuilder(name).ServiceAccount("testsa").Mode("Ref")
	b.Sink(&v1beta1.Destination{
		Ref: &v1.ObjectReference{
			Kind: "Service",
			Name: "foosvc",
		}})

	if resource != "" {
		b.Resources([]v1alpha1.ApiServerResource{{
			APIVersion: "v1",
			Kind:       resource,
			Controller: false,
		}})
	}
	return b.Build()
}
