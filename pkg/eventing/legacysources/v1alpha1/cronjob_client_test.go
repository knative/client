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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
	"knative.dev/eventing/pkg/legacyclient/clientset/versioned/typed/legacysources/v1alpha1/fake"
	"knative.dev/pkg/apis/duck/v1beta1"
)

func setupCronJobSourcesClient(t *testing.T) (sources fake.FakeSourcesV1alpha1, client KnCronJobSourcesClient) {
	sources = fake.FakeSourcesV1alpha1{Fake: &client_testing.Fake{}}
	client = NewKnSourcesClient(&sources, "test-ns").CronJobSourcesClient()
	assert.Equal(t, client.Namespace(), "test-ns")
	return
}

func TestCreateCronJobSource(t *testing.T) {
	sourcesServer, client := setupCronJobSourcesClient(t)

	sourcesServer.AddReactor("create", "cronjobsources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			newSource := a.(client_testing.CreateAction).GetObject()
			name := newSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while creating cronjobsource %s", name)
			}
			return true, newSource, nil
		})

	err := client.CreateCronJobSource(newCronJobSource("testsource", "mysvc"))
	assert.NilError(t, err)

	err = client.CreateCronJobSource(newCronJobSource("testsource", ""))
	assert.ErrorContains(t, err, "sink")
	assert.ErrorContains(t, err, "required")

	err = client.CreateCronJobSource(newCronJobSource("errorSource", "mysvc"))
	assert.ErrorContains(t, err, "errorSource")
}

func TestUpdateCronJobSource(t *testing.T) {
	sourcesServer, client := setupCronJobSourcesClient(t)

	sourcesServer.AddReactor("update", "cronjobsources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			newSource := a.(client_testing.UpdateAction).GetObject()
			name := newSource.(metav1.Object).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating cronjobsource %s", name)
			}
			return true, NewCronJobSourceBuilderFromExisting(newSource.(*v1alpha1.CronJobSource)).Build(), nil
		})

	err := client.UpdateCronJobSource(newCronJobSource("testsource", ""))
	assert.NilError(t, err)

	err = client.UpdateCronJobSource(newCronJobSource("errorSource", ""))
	assert.ErrorContains(t, err, "errorSource")
}

func TestDeleteCronJobSource(t *testing.T) {
	sourcesServer, client := setupCronJobSourcesClient(t)

	sourcesServer.AddReactor("delete", "cronjobsources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating cronjobsource %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteCronJobSource("testsource")
	assert.NilError(t, err)

	err = client.DeleteCronJobSource("errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestGetCronJobSource(t *testing.T) {
	sourcesServer, client := setupCronJobSourcesClient(t)

	sourcesServer.AddReactor("get", "cronjobsources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "errorSource" {
				return true, nil, fmt.Errorf("error while updating cronjobsource %s", name)
			}
			return true, newCronJobSource(name, "mysvc"), nil
		})

	source, err := client.GetCronJobSource("testsource")
	assert.NilError(t, err)
	assert.Equal(t, source.Name, "testsource")
	assert.Equal(t, source.Spec.Sink.Ref.Name, "mysvc")

	_, err = client.GetCronJobSource("errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestListCronJobSource(t *testing.T) {
	sourcesServer, client := setupCronJobSourcesClient(t)

	sourcesServer.AddReactor("list", "cronjobsources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			cJSource := newCronJobSource("testsource", "mysvc")
			return true, &v1alpha1.CronJobSourceList{Items: []v1alpha1.CronJobSource{*cJSource}}, nil
		})

	sourceList, err := client.ListCronJobSource()
	assert.NilError(t, err)
	assert.Equal(t, len(sourceList.Items), 1)
}

func newCronJobSource(name string, sink string) *v1alpha1.CronJobSource {
	b := NewCronJobSourceBuilder(name).
		Schedule("* * * * *").
		Data("mydata")

	if sink != "" {
		b.Sink(
			&v1beta1.Destination{
				Ref: &v1.ObjectReference{
					Kind:      "Service",
					Name:      sink,
					Namespace: "default",
				},
			})
	}
	return b.Build()
}
