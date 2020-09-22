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

package v1alpha2

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func setupFakeContainerSourcesClient() (fakeSvr fake.FakeSourcesV1alpha2, client KnContainerSourcesClient) {
	fakeE := fake.FakeSourcesV1alpha2{Fake: &clienttesting.Fake{}}
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

	err := client.DeleteContainerSource("foo")
	assert.NilError(t, err)

	err = client.DeleteContainerSource("errorSource")
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
	err := client.CreateContainerSource(newContainerSource("foo", "Event"))
	assert.NilError(t, err)

	err = client.CreateContainerSource(newContainerSource("errorSource", "Event"))
	assert.ErrorContains(t, err, "errorSource")

}

func newContainerSource(name, container string) *v1alpha2.ContainerSource {
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
