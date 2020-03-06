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
	"k8s.io/apimachinery/pkg/runtime/schema"
	clienttesting "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/apis/sources/v1alpha2"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2/fake"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"
)

var testNamespace = "test-ns"

func setup() (fakeSvr fake.FakeSourcesV1alpha2, client KnSinkBindingClient) {
	fakeE := fake.FakeSourcesV1alpha2{Fake: &clienttesting.Fake{}}
	cli := NewKnSourcesClient(&fakeE, "test-ns").SinkBindingClient()
	return fakeE, cli
}

func TestDeleteSinkBinding(t *testing.T) {
	var name = "new-binding"
	server, client := setup()

	server.AddReactor("delete", "sinkbindings",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.DeleteAction).GetName()
			if name == "errorSinkBinding" {
				return true, nil, fmt.Errorf("error while deleting binding %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteSinkBinding(name)
	assert.NilError(t, err)

	err = client.DeleteSinkBinding("errorSinkBinding")
	assert.ErrorContains(t, err, "errorSinkBinding")
}

func TestCreateSinkBinding(t *testing.T) {
	var name = "new-binding"
	server, client := setup()

	objNew := newSinkBinding(name, "mysvc", "myping")

	server.AddReactor("create", "sinkbindings",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(clienttesting.CreateAction).GetObject().(metav1.Object).GetName()
			if name == objNew.Name {
				objNew.Generation = 2
				return true, objNew, nil
			}
			return true, nil, fmt.Errorf("error while creating binding %s", name)
		})

	t.Run("create binding without error", func(t *testing.T) {
		err := client.CreateSinkBinding(objNew)
		assert.NilError(t, err)
	})

	t.Run("create binding with an error returns an error object", func(t *testing.T) {
		err := client.CreateSinkBinding(newSinkBinding("unknown", "mysvc", "mypings"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestGetSinkBinding(t *testing.T) {
	var name = "mysinkbinding"
	server, client := setup()

	server.AddReactor("get", "sinkbindings",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			name := a.(clienttesting.GetAction).GetName()
			if name == "errorSinkBinding" {
				return true, nil, fmt.Errorf("error while getting binding %s", name)
			}
			return true, newSinkBinding(name, "mysvc", "myping"), nil
		})

	binding, err := client.GetSinkBinding(name)
	assert.NilError(t, err)
	assert.Equal(t, binding.Name, name)
	assert.Equal(t, binding.Spec.Sink.Ref.Name, "mysvc")
	assert.Equal(t, binding.Spec.Subject.Name, "myping")

	_, err = client.GetSinkBinding("errorSinkBinding")
	assert.ErrorContains(t, err, "errorSinkBinding")
}

func TestListSinkBinding(t *testing.T) {
	serving, client := setup()

	t.Run("list binding returns a list of sink-bindings", func(t *testing.T) {
		binding1 := newSinkBinding("binding-1", "mysvc-1", "myping")
		binding2 := newSinkBinding("binding-2", "mysvc-2", "myping")

		serving.AddReactor("list", "sinkbindings",
			func(a clienttesting.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &v1alpha2.SinkBindingList{Items: []v1alpha2.SinkBinding{*binding1, *binding2}}, nil
			})

		listSinkBindings, err := client.ListSinkBindings()
		assert.NilError(t, err)
		assert.Assert(t, len(listSinkBindings.Items) == 2)
		assert.Equal(t, listSinkBindings.Items[0].Name, "binding-1")
		assert.Equal(t, listSinkBindings.Items[0].Spec.Sink.Ref.Name, "mysvc-1")
		assert.Equal(t, listSinkBindings.Items[0].Spec.Subject.Name, "myping")
		assert.Equal(t, listSinkBindings.Items[1].Name, "binding-2")
		assert.Equal(t, listSinkBindings.Items[1].Spec.Sink.Ref.Name, "mysvc-2")
		assert.Equal(t, listSinkBindings.Items[1].Spec.Subject.Name, "myping")

	})
}

func TestSinkBindingBuilderAddCloudEventOverrides(t *testing.T) {
	aBuilder := NewSinkBindingBuilder("testsinkbinding")
	aBuilder.AddCloudEventOverrides(map[string]string{"type": "foo"})
	a, err := aBuilder.Build()
	assert.NilError(t, err)

	t.Run("update bindings", func(t *testing.T) {
		bBuilder := NewSinkBindingBuilderFromExisting(a)
		b, err := bBuilder.Build()
		assert.NilError(t, err)
		assert.DeepEqual(t, b, a)
		bBuilder.AddCloudEventOverrides(map[string]string{"type": "new"})
		expected := &duckv1.CloudEventOverrides{
			Extensions: map[string]string{
				"type": "new",
			},
		}

		assert.DeepEqual(t, expected, b.Spec.CloudEventOverrides)
	})

	t.Run("update bindings with both new entry and old entry", func(t *testing.T) {
		bBuilder := NewSinkBindingBuilderFromExisting(a)
		b, err := bBuilder.Build()
		assert.NilError(t, err)
		assert.DeepEqual(t, b, a)
		bBuilder.AddCloudEventOverrides(map[string]string{"source": "bar"})
		expected := &duckv1.CloudEventOverrides{
			Extensions: map[string]string{
				"type":   "foo",
				"source": "bar",
			},
		}
		assert.DeepEqual(t, expected, b.Spec.CloudEventOverrides)
	})
}

func TestSinkBindingBuilderForSubjectError(t *testing.T) {

	b, err := NewSinkBindingBuilder("test").SubjectName("bla").Build()
	assert.Assert(t, b == nil)
	assert.ErrorContains(t, err, "group")
	assert.ErrorContains(t, err, "version")
	assert.ErrorContains(t, err, "kind")

	b, err = NewSinkBindingBuilder("test").
		SubjectGVK(&schema.GroupVersionKind{"apps", "v1", "Deployment"}).
		SubjectName("foo").
		AddSubjectMatchLabel("bla", "blub").
		Build()
	assert.ErrorContains(t, err, "label selector")
	assert.ErrorContains(t, err, "name")
	assert.ErrorContains(t, err, "subject")
	assert.Assert(t, b == nil)

}

func TestSinkBindingBuilderForSubject(t *testing.T) {
	gvk := schema.GroupVersionKind{"apps", "v1", "Deployment"}

	b, err := NewSinkBindingBuilder("test").
		SubjectGVK(&gvk).
		SubjectName("foo").
		Build()
	assert.NilError(t, err)
	subject := b.Spec.Subject
	assert.Equal(t, subject.Name, "foo")
	assert.Assert(t, subject.Selector == nil)
	assert.DeepEqual(t, subject.GroupVersionKind(), gvk)

	b, err = NewSinkBindingBuilder("test").
		SubjectGVK(&schema.GroupVersionKind{"apps", "v1", "Deployment"}).
		AddSubjectMatchLabel("bla", "blub").
		AddSubjectMatchLabel("foo", "bar").
		Build()

	assert.NilError(t, err)
	subject = b.Spec.Subject
	assert.Equal(t, subject.Name, "")
	assert.DeepEqual(t, subject.GroupVersionKind(), gvk)
	selector := map[string]string{
		"bla": "blub",
		"foo": "bar",
	}
	assert.DeepEqual(t, subject.Selector.MatchLabels, selector)
}

func TestSinkBindingBuilderForSubjectDirect(t *testing.T) {

	subject := tracker.Reference{
		Name: "direct",
	}
	b, err := NewSinkBindingBuilder("test").
		Subject(&subject).
		SubjectName("nope").                 // should be ignored
		AddSubjectMatchLabel("bla", "blub"). // should be ignored
		Build()
	assert.NilError(t, err)
	subject = b.Spec.Subject
	assert.Equal(t, subject.Name, "direct")
	assert.Assert(t, subject.Selector == nil)
}

func newSinkBinding(name, sinkService, pingName string) *v1alpha2.SinkBinding {
	sink := &duckv1.Destination{
		Ref: &duckv1.KReference{Name: sinkService, Kind: "Service", Namespace: "default", APIVersion: "serving.knative.dev/v1"},
	}
	b, _ := NewSinkBindingBuilder(name).
		Namespace(testNamespace).
		Sink(sink).
		SubjectGVK(&schema.GroupVersionKind{"batch", "v1beta1", "CronJob"}).
		SubjectName(pingName).
		AddCloudEventOverrides(map[string]string{"type": "foo"}).
		Build()
	return b
}
