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

package v1

import (
	"context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/runtime"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clienttesting "k8s.io/client-go/testing"
	fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var testAPIServerSourceNamespace = "test-ns"

func setupAPIServerSourcesClient(t *testing.T) (fakeSources fake.FakeSourcesV1, client KnAPIServerSourcesClient) {
	fakeSources = fake.FakeSourcesV1{Fake: &clienttesting.Fake{}}
	client = NewKnSourcesClient(&fakeSources, testAPIServerSourceNamespace).APIServerSourcesClient()
	assert.Equal(t, client.Namespace(), testAPIServerSourceNamespace)
	return
}

func TestDeleteApiServerSource(t *testing.T) {
	sourcesServer, client := setupAPIServerSourcesClient(t)

	sourcesServer.AddReactor("delete", "apiserversources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.DeleteAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while deleting ApiServer source %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteAPIServerSource(context.Background(), "foo")
	assert.NilError(t, err)

	err = client.DeleteAPIServerSource(context.Background(), "errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestCreateApiServerSource(t *testing.T) {
	sourcesServer, client := setupAPIServerSourcesClient(t)

	sourcesServer.AddReactor("create", "apiserversources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			newSource := a.(clienttesting.CreateAction).GetObject()
			name := newSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while creating ApiServer source %s", name)
			}
			return true, newSource, nil
		})
	err := client.CreateAPIServerSource(context.Background(), newAPIServerSource("foo", "Event"))
	assert.NilError(t, err)

	err = client.CreateAPIServerSource(context.Background(), newAPIServerSource("errorSource", "Event"))
	assert.ErrorContains(t, err, "errorSource")

}

func TestGetApiServerSource(t *testing.T) {
	sourcesServer, client := setupAPIServerSourcesClient(t)

	sourcesServer.AddReactor("get", "apiserversources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while getting ApiServer source %s", name)
			}
			return true, newAPIServerSource(name, "Event"), nil
		})
	testsource, err := client.GetAPIServerSource(context.Background(), "foo")
	assert.NilError(t, err)
	assert.Equal(t, testsource.Name, "foo")
	assert.Equal(t, testsource.Spec.Sink.Ref.Name, "foosvc")

	_, err = client.GetAPIServerSource(context.Background(), "errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestUpdateApiServerSource(t *testing.T) {
	sourcesServer, client := setupAPIServerSourcesClient(t)

	sourcesServer.AddReactor("update", "apiserversources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			updatedSource := a.(clienttesting.UpdateAction).GetObject()
			name := updatedSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating ApiServer source %s", name)
			}
			return true, NewAPIServerSourceBuilderFromExisting(updatedSource.(*v1.ApiServerSource)).Build(), nil
		})
	err := client.UpdateAPIServerSource(context.Background(), newAPIServerSource("foo", "Event"))
	assert.NilError(t, err)

	err = client.UpdateAPIServerSource(context.Background(), newAPIServerSource("errorSource", "Event"))
	assert.ErrorContains(t, err, "errorSource")
}

func TestListAPIServerSource(t *testing.T) {
	sourcesServer, client := setupAPIServerSourcesClient(t)

	sourcesServer.AddReactor("list", "apiserversources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			cJSource := newAPIServerSource("testsource", "Event")
			return true, &v1.ApiServerSourceList{Items: []v1.ApiServerSource{*cJSource}}, nil
		})

	sourceList, err := client.ListAPIServerSource(context.Background())
	assert.NilError(t, err)
	assert.Equal(t, len(sourceList.Items), 1)
}

func newAPIServerSource(name, resource string) *v1.ApiServerSource {
	b := NewAPIServerSourceBuilder(name).
		ServiceAccount("testsa").
		EventMode("Reference").
		CloudEventOverrides(map[string]string{"type": "foo"}, []string{}).
		Sink(duckv1.Destination{
			Ref: &duckv1.KReference{
				Kind:      "Service",
				Name:      "foosvc",
				Namespace: "default",
			}})

	if resource != "" {
		b.Resources([]v1.APIVersionKindSelector{{
			APIVersion: "v1",
			Kind:       resource,
		}})
	}
	return b.Build()
}
