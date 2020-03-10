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
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1/fake"
)

var testNamespace = "test-ns"

func setup() (fakeSvr fake.FakeEventingV1alpha1, client KnEventingClient) {
	fakeE := fake.FakeEventingV1alpha1{Fake: &client_testing.Fake{}}
	cli := NewKnEventingClient(&fakeE, testNamespace)
	return fakeE, cli
}

func TestDeleteTrigger(t *testing.T) {
	var name = "new-trigger"
	server, client := setup()

	server.AddReactor("delete", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorTrigger" {
				return true, nil, fmt.Errorf("error while deleting trigger %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteTrigger(name)
	assert.NilError(t, err)

	err = client.DeleteTrigger("errorTrigger")
	assert.ErrorContains(t, err, "errorTrigger")
}

func TestCreateTrigger(t *testing.T) {
	var name = "new-trigger"
	server, client := setup()

	objNew := newTrigger(name)

	server.AddReactor("create", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
			if name == objNew.Name {
				objNew.Generation = 2
				return true, objNew, nil
			}
			return true, nil, fmt.Errorf("error while creating trigger %s", name)
		})

	t.Run("create trigger without error", func(t *testing.T) {
		err := client.CreateTrigger(objNew)
		assert.NilError(t, err)
	})

	t.Run("create trigger with an error returns an error object", func(t *testing.T) {
		err := client.CreateTrigger(newTrigger("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestGetTrigger(t *testing.T) {
	var name = "mytrigger"
	server, client := setup()

	server.AddReactor("get", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "errorTrigger" {
				return true, nil, fmt.Errorf("error while getting trigger %s", name)
			}
			return true, newTrigger(name), nil
		})

	trigger, err := client.GetTrigger(name)
	assert.NilError(t, err)
	assert.Equal(t, trigger.Name, name)
	assert.Equal(t, trigger.Spec.Broker, "default")

	_, err = client.GetTrigger("errorTrigger")
	assert.ErrorContains(t, err, "errorTrigger")
}

func TestListTrigger(t *testing.T) {
	serving, client := setup()

	t.Run("list trigger returns a list of triggers", func(t *testing.T) {
		trigger1 := newTrigger("trigger-1")
		trigger2 := newTrigger("trigger-2")

		serving.AddReactor("list", "triggers",
			func(a client_testing.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &v1alpha1.TriggerList{Items: []v1alpha1.Trigger{*trigger1, *trigger2}}, nil
			})

		listTriggers, err := client.ListTriggers()
		assert.NilError(t, err)
		assert.Assert(t, len(listTriggers.Items) == 2)
		assert.Equal(t, listTriggers.Items[0].Name, "trigger-1")
		assert.Equal(t, listTriggers.Items[1].Name, "trigger-2")
	})
}

func TestTriggerBuilder(t *testing.T) {
	a := NewTriggerBuilder("testtrigger")
	a.Filters(map[string]string{"type": "foo"})

	t.Run("update filters", func(t *testing.T) {
		b := NewTriggerBuilderFromExisting(a.Build())
		assert.DeepEqual(t, b.Build(), a.Build())
		b.Filters(map[string]string{"type": "new"})
		expected := &v1alpha1.TriggerFilter{
			Attributes: &v1alpha1.TriggerFilterAttributes{
				"type": "new",
			},
		}
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)
	})

	t.Run("update filters with both new entry and updated entry", func(t *testing.T) {
		b := NewTriggerBuilderFromExisting(a.Build())
		assert.DeepEqual(t, b.Build(), a.Build())
		b.Filters(map[string]string{"type": "new", "source": "bar"})
		expected := &v1alpha1.TriggerFilter{
			Attributes: &v1alpha1.TriggerFilterAttributes{
				"type":   "new",
				"source": "bar",
			},
		}
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)
	})

	t.Run("add and remove inject annotation", func(t *testing.T) {
		b := NewTriggerBuilder("broker-trigger")
		b.InjectBroker(true)
		expected := &metav1.ObjectMeta{
			Annotations: map[string]string{
				v1alpha1.InjectionAnnotation: "enabled",
			},
		}
		assert.DeepEqual(t, expected.Annotations, b.Build().ObjectMeta.Annotations)

		b = NewTriggerBuilderFromExisting(b.Build())
		b.InjectBroker(false)
		assert.DeepEqual(t, make(map[string]string), b.Build().ObjectMeta.Annotations)

	})

}

func newTrigger(name string) *v1alpha1.Trigger {
	return NewTriggerBuilder(name).
		Namespace(testNamespace).
		Broker("default").
		Filters(map[string]string{"type": "foo"}).
		Build()
}
