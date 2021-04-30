/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"
	fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func setupFakeContainerSourcesClient() (fakeSvr fake.FakeSourcesV1, client KnContainerSourcesClient) {
	fakeE := fake.FakeSourcesV1{Fake: &clienttesting.Fake{}}
	cli := NewKnSourcesClient(&fakeE, "test-ns").ContainerSourcesClient()
	return fakeE, cli
}

func TestDeleteContainerSourceSource(t *testing.T) {
	sourcesServer, client := setupFakeContainerSourcesClient()

	sourcesServer.AddReactor("delete", "containersources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.DeleteAction).GetName()
			fmt.Printf("name=%s \n", name)
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while deleting ContainerSource source %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteContainerSource("foo", context.Background())
	assert.NilError(t, err)

	err = client.DeleteContainerSource("errorSource", context.Background())
	assert.ErrorContains(t, err, "errorSource")
}

func TestCreateContainerSourceSource(t *testing.T) {
	sourcesServer, client := setupFakeContainerSourcesClient()

	sourcesServer.AddReactor("create", "containersources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			newSource := a.(clienttesting.CreateAction).GetObject()
			name := newSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while creating ContainerSource source %s", name)
			}
			return true, newSource, nil
		})
	err := client.CreateContainerSource(context.Background(), newContainerSource("foo", "Event"))
	assert.NilError(t, err)

	err = client.CreateContainerSource(context.Background(), newContainerSource("errorSource", "Event"))
	assert.ErrorContains(t, err, "errorSource")

}

func TestGetContainerSource(t *testing.T) {
	sourcesServer, client := setupFakeContainerSourcesClient()

	sourcesServer.AddReactor("get", "containersources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while getting Container source %s", name)
			}
			return true, newContainerSource(name, "Event"), nil
		})
	testsource, err := client.GetContainerSource(context.Background(), "foo")
	assert.NilError(t, err)
	assert.Equal(t, testsource.Name, "foo")
	assert.Equal(t, testsource.Spec.Sink.Ref.Name, "foosvc")

	_, err = client.GetContainerSource(context.Background(), "errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestUpdateContainerSource(t *testing.T) {
	sourcesServer, client := setupFakeContainerSourcesClient()

	sourcesServer.AddReactor("update", "containersources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			updatedSource := a.(clienttesting.UpdateAction).GetObject()
			name := updatedSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating Container source %s", name)
			}
			return true, NewContainerSourceBuilderFromExisting(updatedSource.(*v1.ContainerSource)).Build(), nil
		})
	err := client.UpdateContainerSource(context.Background(), newContainerSource("foo", "Event"))
	assert.NilError(t, err)

	err = client.UpdateContainerSource(context.Background(), newContainerSource("errorSource", "Event"))
	assert.ErrorContains(t, err, "errorSource")
}

func TestListContainerSource(t *testing.T) {
	sourcesServer, client := setupFakeContainerSourcesClient()

	sourcesServer.AddReactor("list", "containersources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			cJSource := newContainerSource("testsource", "Event")
			return true, &v1.ContainerSourceList{Items: []v1.ContainerSource{*cJSource}}, nil
		})

	sourceList, err := client.ListContainerSources(context.Background())
	assert.NilError(t, err)
	assert.Equal(t, len(sourceList.Items), 1)
}

func newContainerSource(name, container string) *v1.ContainerSource {
	b := NewContainerSourceBuilder(name).
		PodSpec(corev1.PodSpec{}).
		Sink(duckv1.Destination{
			Ref: &duckv1.KReference{
				Kind:      "Service",
				Name:      "foosvc",
				Namespace: "default",
			}})

	return b.Build()
}
