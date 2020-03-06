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

package v1alpha2

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/apis/sources/v1alpha2"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func setupPingSourcesClient(t *testing.T) (sources fake.FakeSourcesV1alpha2, client KnPingSourcesClient) {
	sources = fake.FakeSourcesV1alpha2{Fake: &clienttesting.Fake{}}
	client = NewKnSourcesClient(&sources, "test-ns").PingSourcesClient()
	assert.Equal(t, client.Namespace(), "test-ns")
	return
}

func TestCreatePingSource(t *testing.T) {
	sourcesServer, client := setupPingSourcesClient(t)

	sourcesServer.AddReactor("create", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			newSource := a.(clienttesting.CreateAction).GetObject()
			name := newSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while creating pingsource %s", name)
			}
			return true, newSource, nil
		})

	err := client.CreatePingSource(newPingSource("testsource", "mysvc"))
	assert.NilError(t, err)

	err = client.CreatePingSource(newPingSource("testsource", ""))
	assert.ErrorContains(t, err, "sink")
	assert.ErrorContains(t, err, "required")

	err = client.CreatePingSource(newPingSource("errorSource", "mysvc"))
	assert.ErrorContains(t, err, "errorSource")
}

func TestUpdatePingSource(t *testing.T) {
	sourcesServer, client := setupPingSourcesClient(t)

	sourcesServer.AddReactor("update", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			newSource := a.(clienttesting.UpdateAction).GetObject()
			name := newSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating pingsource %s", name)
			}
			return true, NewPingSourceBuilderFromExisting(newSource.(*v1alpha2.PingSource)).Build(), nil
		})

	err := client.UpdatePingSource(newPingSource("testsource", ""))
	assert.NilError(t, err)

	err = client.UpdatePingSource(newPingSource("errorSource", ""))
	assert.ErrorContains(t, err, "errorSource")
}

func TestDeletePingSource(t *testing.T) {
	sourcesServer, client := setupPingSourcesClient(t)

	sourcesServer.AddReactor("delete", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.DeleteAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating pingsource %s", name)
			}
			return true, nil, nil
		})

	err := client.DeletePingSource("testsource")
	assert.NilError(t, err)

	err = client.DeletePingSource("errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestGetPingSource(t *testing.T) {
	sourcesServer, client := setupPingSourcesClient(t)

	sourcesServer.AddReactor("get", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating pingsource %s", name)
			}
			return true, newPingSource(name, "mysvc"), nil
		})

	source, err := client.GetPingSource("testsource")
	assert.NilError(t, err)
	assert.Equal(t, source.Name, "testsource")
	assert.Equal(t, source.Spec.Sink.Ref.Name, "mysvc")

	_, err = client.GetPingSource("errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestListPingSource(t *testing.T) {
	sourcesServer, client := setupPingSourcesClient(t)

	sourcesServer.AddReactor("list", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			cJSource := newPingSource("testsource", "mysvc")
			return true, &v1alpha2.PingSourceList{Items: []v1alpha2.PingSource{*cJSource}}, nil
		})

	sourceList, err := client.ListPingSource()
	assert.NilError(t, err)
	assert.Equal(t, len(sourceList.Items), 1)
}

func newPingSource(name string, sink string) *v1alpha2.PingSource {
	b := NewPingSourceBuilder(name).
		Schedule("* * * * *").
		JsonData("mydata")

	if sink != "" {
		b.Sink(
			duckv1.Destination{
				Ref: &duckv1.KReference{
					Kind:      "Service",
					Name:      sink,
					Namespace: "default",
				},
			})
	}
	return b.Build()
}
