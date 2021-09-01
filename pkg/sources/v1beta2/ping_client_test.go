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

package v1beta2

import (
	"context"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	sourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1beta2/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func setupPingSourcesClient(t *testing.T) (sources fake.FakeSourcesV1beta2, client KnPingSourcesClient) {
	sources = fake.FakeSourcesV1beta2{Fake: &clienttesting.Fake{}}
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

	err := client.CreatePingSource(context.Background(), newPingSource("testsource", "mysvc"))
	assert.NilError(t, err)

	err = client.CreatePingSource(context.Background(), newPingSource("testsource", ""))
	assert.ErrorContains(t, err, "sink")
	assert.ErrorContains(t, err, "required")

	err = client.CreatePingSource(context.Background(), newPingSource("errorSource", "mysvc"))
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
			return true, NewPingSourceBuilderFromExisting(newSource.(*sourcesv1beta2.PingSource)).Build(), nil
		})

	err := client.UpdatePingSource(context.Background(), newPingSource("testsource", ""))
	assert.NilError(t, err)

	err = client.UpdatePingSource(context.Background(), newPingSource("errorSource", ""))
	assert.ErrorContains(t, err, "errorSource")
}

func TestUpdatePingSourceWithRetry(t *testing.T) {
	sourcesServer, client := setupPingSourcesClient(t)

	var attemptCount, maxAttempts = 0, 5
	var newData = "newData"
	sourcesServer.AddReactor("get", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			if name == "deletedSource" {
				source := newPingSource(name, "mysvc")
				now := metav1.Now()
				source.DeletionTimestamp = &now
				return true, source, nil
			}
			if name == "getErrorSource" {
				return true, nil, errors.NewInternalError(fmt.Errorf("mock internal error"))
			}
			return true, newPingSource(name, "mysvc"), nil
		})

	sourcesServer.AddReactor("update", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			newSource := a.(clienttesting.UpdateAction).GetObject()
			name := newSource.(metav1.Object).GetName()

			if name == "testSource" && attemptCount > 0 {
				attemptCount--
				return true, nil, errors.NewConflict(sourcesv1beta2.Resource("pingsource"), "errorSource", fmt.Errorf("error updating because of conflict"))
			}
			if name == "errorSource" {
				return true, nil, errors.NewInternalError(fmt.Errorf("mock internal error"))
			}
			return true, NewPingSourceBuilderFromExisting(newSource.(*sourcesv1beta2.PingSource)).Build(), nil
		})

	err := client.UpdatePingSourceWithRetry(context.Background(), "testSource", func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error) {
		origSource.Spec.Data = newData
		return origSource, nil
	}, maxAttempts)
	assert.NilError(t, err, "No retries required as no conflict error occurred")

	attemptCount = maxAttempts - 1
	err = client.UpdatePingSourceWithRetry(context.Background(), "testSource", func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error) {
		origSource.Spec.Data = newData
		return origSource, nil
	}, maxAttempts)
	assert.NilError(t, err, "Update retried %d times and succeeded", maxAttempts)
	assert.Equal(t, attemptCount, 0)

	attemptCount = maxAttempts
	err = client.UpdatePingSourceWithRetry(context.Background(), "testSource", func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error) {
		origSource.Spec.Data = newData
		return origSource, nil
	}, maxAttempts)
	assert.ErrorType(t, err, errors.IsConflict, "Update retried %d times and failed", maxAttempts)
	assert.Equal(t, attemptCount, 0)

	err = client.UpdatePingSourceWithRetry(context.Background(), "errorSource", func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error) {
		origSource.Spec.Data = newData
		return origSource, nil
	}, maxAttempts)
	assert.ErrorType(t, err, errors.IsInternalError)

	err = client.UpdatePingSourceWithRetry(context.Background(), "deletedSource", func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error) {
		origSource.Spec.Data = newData
		return origSource, nil
	}, maxAttempts)
	assert.ErrorContains(t, err, "marked for deletion")

	err = client.UpdatePingSourceWithRetry(context.Background(), "testSource", func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error) {
		origSource.Spec.Data = newData
		return origSource, fmt.Errorf("error updating object")
	}, maxAttempts)
	assert.ErrorContains(t, err, "error updating object")

	err = client.UpdatePingSourceWithRetry(context.Background(), "getErrorSource", func(origSource *sourcesv1beta2.PingSource) (*sourcesv1beta2.PingSource, error) {
		origSource.Spec.Data = newData
		return origSource, nil
	}, maxAttempts)
	assert.ErrorType(t, err, errors.IsInternalError)
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

	err := client.DeletePingSource(context.Background(), "testsource")
	assert.NilError(t, err)

	err = client.DeletePingSource(context.Background(), "errorSource")
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

	source, err := client.GetPingSource(context.Background(), "testsource")
	assert.NilError(t, err)
	assert.Equal(t, source.Name, "testsource")
	assert.Equal(t, source.Spec.Sink.Ref.Name, "mysvc")

	_, err = client.GetPingSource(context.Background(), "errorSource")
	assert.ErrorContains(t, err, "errorSource")
}

func TestListPingSource(t *testing.T) {
	sourcesServer, client := setupPingSourcesClient(t)

	sourcesServer.AddReactor("list", "pingsources",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			cJSource := newPingSource("testsource", "mysvc")
			return true, &sourcesv1beta2.PingSourceList{Items: []sourcesv1beta2.PingSource{*cJSource}}, nil
		})

	sourceList, err := client.ListPingSource(context.Background())
	assert.NilError(t, err)
	assert.Equal(t, len(sourceList.Items), 1)
}

func newPingSource(name string, sink string) *sourcesv1beta2.PingSource {
	b := NewPingSourceBuilder(name).
		Schedule("* * * * *").
		Data("mydata").
		CloudEventOverrides(map[string]string{"type": "foo"}, []string{})

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
