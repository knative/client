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

package v1beta1

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/client/pkg/wait"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1beta1/fake"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var testNamespace = "test-ns"

func setup() (fakeSvr fake.FakeEventingV1beta1, client KnEventingClient) {
	fakeE := fake.FakeEventingV1beta1{Fake: &client_testing.Fake{}}
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

	err := client.DeleteTrigger(context.Background(), name)
	assert.NilError(t, err)

	err = client.DeleteTrigger(context.Background(), "errorTrigger")
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
		err := client.CreateTrigger(context.Background(), objNew)
		assert.NilError(t, err)
	})

	t.Run("create trigger with an error returns an error object", func(t *testing.T) {
		err := client.CreateTrigger(context.Background(), newTrigger("unknown"))
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

	trigger, err := client.GetTrigger(context.Background(), name)
	assert.NilError(t, err)
	assert.Equal(t, trigger.Name, name)
	assert.Equal(t, trigger.Spec.Broker, "default")

	_, err = client.GetTrigger(context.Background(), "errorTrigger")
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
				return true, &v1beta1.TriggerList{Items: []v1beta1.Trigger{*trigger1, *trigger2}}, nil
			})

		listTriggers, err := client.ListTriggers(context.Background())
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
		expected := &v1beta1.TriggerFilter{
			Attributes: v1beta1.TriggerFilterAttributes{
				"type": "new",
			},
		}
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)
	})

	t.Run("update filters with both new entry and updated entry", func(t *testing.T) {
		b := NewTriggerBuilderFromExisting(a.Build())
		assert.DeepEqual(t, b.Build(), a.Build())
		b.Filters(map[string]string{"type": "new", "source": "bar"})
		expected := &v1beta1.TriggerFilter{
			Attributes: v1beta1.TriggerFilterAttributes{
				"type":   "new",
				"source": "bar",
			},
		}
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)
	})

	t.Run("update filters to remove filters", func(t *testing.T) {
		b := NewTriggerBuilderFromExisting(a.Build())
		assert.DeepEqual(t, b.Build(), a.Build())
		b.Filters(nil)
		expected := &v1beta1.TriggerFilter{}
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)

		b.Filters((make(map[string]string)))
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)
	})

	t.Run("add and remove inject annotation", func(t *testing.T) {
		b := NewTriggerBuilder("broker-trigger")
		b.InjectBroker(true)
		expected := &metav1.ObjectMeta{
			Annotations: map[string]string{
				v1beta1.InjectionAnnotation: "enabled",
			},
		}
		assert.DeepEqual(t, expected.Annotations, b.Build().ObjectMeta.Annotations)

		b = NewTriggerBuilderFromExisting(b.Build())
		b.InjectBroker(false)
		assert.DeepEqual(t, make(map[string]string), b.Build().ObjectMeta.Annotations)

	})

}

func TestBrokerCreate(t *testing.T) {
	var name = "broker"
	server, client := setup()

	objNew := newBroker(name)

	server.AddReactor("create", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
			if name == objNew.Name {
				objNew.Generation = 2
				return true, objNew, nil
			}
			return true, nil, fmt.Errorf("error while creating broker %s", name)
		})

	t.Run("create broker without error", func(t *testing.T) {
		err := client.CreateBroker(context.Background(), objNew)
		assert.NilError(t, err)
	})

	t.Run("create broker with an error returns an error object", func(t *testing.T) {
		err := client.CreateBroker(context.Background(), newBroker("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})
}

func TestBrokerGet(t *testing.T) {
	var name = "foo"
	server, client := setup()

	server.AddReactor("get", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "errorBroker" {
				return true, nil, fmt.Errorf("error while getting broker %s", name)
			}
			return true, newBroker(name), nil
		})

	broker, err := client.GetBroker(context.Background(), name)
	assert.NilError(t, err)
	assert.Equal(t, broker.Name, name)

	_, err = client.GetBroker(context.Background(), "errorBroker")
	assert.ErrorContains(t, err, "errorBroker")
}

func TestBrokerDelete(t *testing.T) {
	var name = "fooBroker"
	server, client := setup()

	server.AddReactor("delete", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorBroker" {
				return true, nil, fmt.Errorf("error while deleting broker %s", name)
			}
			return true, nil, nil
		})

	err := client.DeleteBroker(context.Background(), name, 0)
	assert.NilError(t, err)

	err = client.DeleteBroker(context.Background(), "errorBroker", 0)
	assert.ErrorContains(t, err, "errorBroker", 0)
}

func TestBrokerDeleteWithWait(t *testing.T) {
	var name = "fooBroker"
	server, client := setup()

	server.AddReactor("delete", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorBroker" {
				return true, nil, fmt.Errorf("error while deleting broker %s", name)
			}
			return true, nil, nil
		})

	server.AddWatchReactor("brokers",
		func(a client_testing.Action) (bool, watch.Interface, error) {
			watchAction := a.(client_testing.WatchAction)
			name, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
			if !found {
				return true, nil, errors.NewNotFound(v1beta1.Resource("broker"), name)
			}
			w := wait.NewFakeWatch(getBrokerDeleteEvents("fooBroker"))
			w.Start()
			return true, w, nil
		})

	err := client.DeleteBroker(context.Background(), name, time.Duration(10)*time.Second)
	assert.NilError(t, err)

	err = client.DeleteBroker(context.Background(), "errorBroker", time.Duration(10)*time.Second)
	assert.ErrorContains(t, err, "errorBroker", time.Duration(10)*time.Second)
}

func TestBrokerList(t *testing.T) {
	serving, client := setup()

	t.Run("broker list returns a list of brokers", func(t *testing.T) {
		broker1 := newBroker("foo1")
		broker2 := newBroker("foo2")

		serving.AddReactor("list", "brokers",
			func(a client_testing.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &v1beta1.BrokerList{Items: []v1beta1.Broker{*broker1, *broker2}}, nil
			})

		brokerList, err := client.ListBrokers(context.Background())
		assert.NilError(t, err)
		assert.Assert(t, len(brokerList.Items) == 2)
		assert.Equal(t, brokerList.Items[0].Name, "foo1")
		assert.Equal(t, brokerList.Items[1].Name, "foo2")
	})
}

func newTrigger(name string) *v1beta1.Trigger {
	return NewTriggerBuilder(name).
		Namespace(testNamespace).
		Broker("default").
		Filters(map[string]string{"type": "foo"}).
		Build()
}

func newBroker(name string) *v1beta1.Broker {
	return NewBrokerBuilder(name).
		Namespace(testNamespace).
		Build()
}

func getBrokerDeleteEvents(name string) []watch.Event {
	return []watch.Event{
		{Type: watch.Added, Object: createBrokerWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", "msg1")},
		{Type: watch.Modified, Object: createBrokerWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", "msg2")},
		{Type: watch.Deleted, Object: createBrokerWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "")},
	}
}

func createBrokerWithConditions(name string, readyStatus corev1.ConditionStatus, otherReadyStatus corev1.ConditionStatus, reason string, message string) runtime.Object {
	broker := newBroker(name)
	broker.Status.Conditions = duckv1.Conditions([]apis.Condition{
		{Type: "ChannelServiceReady", Status: otherReadyStatus},
		{Type: apis.ConditionReady, Status: readyStatus, Reason: reason, Message: message},
		{Type: "Addressable", Status: otherReadyStatus},
	})
	return broker
}
