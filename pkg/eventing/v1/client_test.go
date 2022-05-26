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
	"time"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	v1 "knative.dev/eventing/pkg/apis/duck/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/client/pkg/wait"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1/fake"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var (
	testNamespace = "test-ns"
	testClass     = "test-class"
)

func setup() (fakeSvr fake.FakeEventingV1, client KnEventingClient) {
	fakeE := fake.FakeEventingV1{Fake: &client_testing.Fake{}}
	cli := NewKnEventingClient(&fakeE, testNamespace)
	return fakeE, cli
}

func TestNamespace(t *testing.T) {
	_, client := setup()
	assert.Equal(t, testNamespace, client.Namespace())
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
				return true, &eventingv1.TriggerList{Items: []eventingv1.Trigger{*trigger1, *trigger2}}, nil
			})

		listTriggers, err := client.ListTriggers(context.Background())
		assert.NilError(t, err)
		assert.Assert(t, len(listTriggers.Items) == 2)
		assert.Equal(t, listTriggers.Items[0].Name, "trigger-1")
		assert.Equal(t, listTriggers.Items[1].Name, "trigger-2")
	})
}

func TestUpdateTrigger(t *testing.T) {
	serving, client := setup()
	t.Run("update trigger will update the trigger",
		func(t *testing.T) {
			trigger := newTrigger("trigger-1")
			errTrigger := newTrigger("errorTrigger")
			serving.AddReactor("update", "triggers",
				func(a client_testing.Action) (bool, runtime.Object, error) {
					newTrigger := a.(client_testing.UpdateAction).GetObject()
					name := newTrigger.(metav1.Object).GetName()
					if name == "errorTrigger" {
						return true, nil, errors.NewInternalError(fmt.Errorf("mock internal error"))
					}
					return true, trigger, nil
				})
			err := client.UpdateTrigger(context.Background(), trigger)
			assert.NilError(t, err)
			err = client.UpdateTrigger(context.Background(), errTrigger)
			assert.ErrorType(t, err, errors.IsInternalError)
		})
}

func TestUpdateTriggerWithRetry(t *testing.T) {
	serving, client := setup()
	var attemptCount, maxAttempts = 0, 5
	serving.AddReactor("get", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "deletedTrigger" {
				trigger := newTrigger(name)
				now := metav1.Now()
				trigger.DeletionTimestamp = &now
				return true, trigger, nil
			}
			if name == "getErrorTrigger" {
				return true, nil, errors.NewInternalError(fmt.Errorf("mock internal error"))
			}
			return true, newTrigger(name), nil
		})

	serving.AddReactor("update", "triggers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			newTrigger := a.(client_testing.UpdateAction).GetObject()
			name := newTrigger.(metav1.Object).GetName()

			if name == "testTrigger" && attemptCount > 0 {
				attemptCount--
				return true, nil, errors.NewConflict(eventingv1.Resource("trigger"), "errorTrigger", fmt.Errorf("error updating because of conflict"))
			}
			if name == "errorTrigger" {
				return true, nil, errors.NewInternalError(fmt.Errorf("mock internal error"))
			}
			return true, NewTriggerBuilderFromExisting(newTrigger.(*eventingv1.Trigger)).Build(), nil
		})

	t.Run("Update trigger successfully without any retries", func(t *testing.T) {
		err := client.UpdateTriggerWithRetry(context.Background(), "testTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, nil
		}, maxAttempts)
		assert.NilError(t, err, "No retries required as no conflict error occurred")
	})

	t.Run("Update trigger with retry after max retries", func(t *testing.T) {
		attemptCount = maxAttempts - 1
		err := client.UpdateTriggerWithRetry(context.Background(), "testTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, nil
		}, maxAttempts)
		assert.NilError(t, err, "Update retried %d times and succeeded", maxAttempts)
		assert.Equal(t, attemptCount, 0)
	})

	t.Run("Update trigger with retry and fail with conflict after exhausting max retries", func(t *testing.T) {
		attemptCount = maxAttempts
		err := client.UpdateTriggerWithRetry(context.Background(), "testTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsConflict, "Update retried %d times and failed", maxAttempts)
		assert.Equal(t, attemptCount, 0)
	})

	t.Run("Update trigger with retry and fail with conflict after exhausting max retries", func(t *testing.T) {
		attemptCount = maxAttempts
		err := client.UpdateTriggerWithRetry(context.Background(), "testTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsConflict, "Update retried %d times and failed", maxAttempts)
		assert.Equal(t, attemptCount, 0)
	})

	t.Run("Update trigger with retry fails with a non conflict error", func(t *testing.T) {
		err := client.UpdateTriggerWithRetry(context.Background(), "errorTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsInternalError)
	})

	t.Run("Update trigger with retry fails with resource already deleted error", func(t *testing.T) {
		err := client.UpdateTriggerWithRetry(context.Background(), "deletedTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, nil
		}, maxAttempts)
		assert.ErrorContains(t, err, "marked for deletion")
	})

	t.Run("Update trigger with retry fails with error from updateFunc", func(t *testing.T) {
		err := client.UpdateTriggerWithRetry(context.Background(), "testTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, fmt.Errorf("error updating object")
		}, maxAttempts)
		assert.ErrorContains(t, err, "error updating object")
	})

	t.Run("Update trigger with retry fails with error from GetTrigger", func(t *testing.T) {
		err := client.UpdateTriggerWithRetry(context.Background(), "getErrorTrigger", func(trigger *eventingv1.Trigger) (*eventingv1.Trigger, error) {
			return trigger, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsInternalError)
	})
}

func TestTriggerBuilder(t *testing.T) {
	a := NewTriggerBuilder("testtrigger")
	a.Filters(map[string]string{"type": "foo"})

	t.Run("update filters", func(t *testing.T) {
		b := NewTriggerBuilderFromExisting(a.Build())
		assert.DeepEqual(t, b.Build(), a.Build())
		b.Filters(map[string]string{"type": "new"})
		expected := &eventingv1.TriggerFilter{
			Attributes: eventingv1.TriggerFilterAttributes{
				"type": "new",
			},
		}
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)
	})

	t.Run("update filters with both new entry and updated entry", func(t *testing.T) {
		b := NewTriggerBuilderFromExisting(a.Build())
		assert.DeepEqual(t, b.Build(), a.Build())
		b.Filters(map[string]string{"type": "new", "source": "bar"})
		expected := &eventingv1.TriggerFilter{
			Attributes: eventingv1.TriggerFilterAttributes{
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
		expected := &eventingv1.TriggerFilter{}
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)

		b.Filters((make(map[string]string)))
		assert.DeepEqual(t, expected, b.Build().Spec.Filter)
	})

	t.Run("add and remove inject annotation", func(t *testing.T) {
		b := NewTriggerBuilder("broker-trigger")
		b.InjectBroker(true)
		expected := &metav1.ObjectMeta{
			Annotations: map[string]string{
				eventingv1.InjectionAnnotation: "enabled",
			},
		}
		assert.DeepEqual(t, expected.Annotations, b.Build().ObjectMeta.Annotations)

		b = NewTriggerBuilderFromExisting(b.Build())
		b.InjectBroker(false)
		assert.DeepEqual(t, make(map[string]string), b.Build().ObjectMeta.Annotations)

	})
}

func TestWithGvk(t *testing.T) {
	t.Run("Broker withGvk", func(t *testing.T) {
		b := NewBrokerBuilder("test").WithGvk().Build()
		assert.Assert(t, !b.GroupVersionKind().Empty())
	})
	t.Run("Trigger withGvk", func(t *testing.T) {
		trigger := NewTriggerBuilder("test").WithGvk().Build()
		assert.Assert(t, !trigger.GroupVersionKind().Empty())
	})
}

func TestBrokerCreate(t *testing.T) {
	var name = "broker"
	server, client := setup()

	objNew := newBroker(name)
	brokerObjWithClass := newBrokerWithClass(name)
	brokerObjWithDeliveryOptions := newBrokerWithDeliveryOptions(name)
	brokerObjWithNilDeliveryOptions := newBrokerWithNilDeliveryOptions(name)

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

	t.Run("create broker with class without error", func(t *testing.T) {
		err := client.CreateBroker(context.Background(), brokerObjWithClass)
		assert.NilError(t, err)
	})

	t.Run("create broker with an error returns an error object", func(t *testing.T) {
		err := client.CreateBroker(context.Background(), newBroker("unknown"))
		assert.ErrorContains(t, err, "unknown")
	})

	t.Run("create broker with delivery options", func(t *testing.T) {
		err := client.CreateBroker(context.Background(), brokerObjWithDeliveryOptions)
		assert.NilError(t, err)
	})

	t.Run("create broker with nil delivery options", func(t *testing.T) {
		err := client.CreateBroker(context.Background(), brokerObjWithNilDeliveryOptions)
		assert.NilError(t, err)
	})

	t.Run("create broker with nil delivery spec", func(t *testing.T) {
		builderFuncs := []func(builder *BrokerBuilder) *BrokerBuilder{
			func(builder *BrokerBuilder) *BrokerBuilder {
				var sink = &duckv1.Destination{
					Ref: &duckv1.KReference{Name: "test-svc", Kind: "Service", APIVersion: "serving.knative.dev/v1", Namespace: "default"},
				}
				return builder.DlSink(sink)
			},
			func(builder *BrokerBuilder) *BrokerBuilder {
				var retry int32 = 5
				return builder.Retry(&retry)
			},
			func(builder *BrokerBuilder) *BrokerBuilder {
				var timeout = "PT5S"
				return builder.Timeout(&timeout)
			},
			func(builder *BrokerBuilder) *BrokerBuilder {
				var policy = v1.BackoffPolicyType("linear")
				return builder.BackoffPolicy(&policy)
			},
			func(builder *BrokerBuilder) *BrokerBuilder {
				var delay = "PT5S"
				return builder.BackoffDelay(&delay)
			},
			func(builder *BrokerBuilder) *BrokerBuilder {
				var max = "PT5S"
				return builder.RetryAfterMax(&max)
			},
		}
		for _, bf := range builderFuncs {
			brokerBuilder := NewBrokerBuilder(name)
			brokerBuilder.broker.Spec.Delivery = nil
			updatedBuilder := bf(brokerBuilder)

			broker := updatedBuilder.Build()
			err := client.CreateBroker(context.Background(), broker)
			assert.NilError(t, err)
		}
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

	server.AddReactor("get", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "notFound" {
				return true, nil, errors.NewNotFound(eventingv1.Resource("broker"), "notFound")
			}
			return false, nil, nil
		})
	server.AddReactor("delete", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorBroker" {
				return true, nil, fmt.Errorf("error while deleting broker %s", name)
			}
			return false, nil, nil
		})

	err := client.DeleteBroker(context.Background(), name, 0)
	assert.NilError(t, err)

	err = client.DeleteBroker(context.Background(), "errorBroker", 0)
	assert.ErrorContains(t, err, "errorBroker", 0)

	err = client.DeleteBroker(context.Background(), "notFound", 0)
	assert.ErrorContains(t, err, "not found", 0)
	assert.ErrorContains(t, err, "notFound", 0)
}

func TestBrokerDeleteWithWait(t *testing.T) {
	var brokerName = "fooBroker"
	var deleted = "deletedBroker"
	server, client := setup()

	server.AddReactor("get", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == deleted {
				deletedBroker := newBroker(deleted)
				deletedBroker.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				return true, deletedBroker, nil
			}
			return false, nil, nil
		})
	server.AddReactor("delete", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.DeleteAction).GetName()
			if name == "errorBroker" {
				return true, nil, fmt.Errorf("error while deleting broker %s", name)
			}
			if name == deleted {
				deletedBroker := newBroker(deleted)
				deletedBroker.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				return true, deletedBroker, nil
			}
			return true, nil, nil
		})
	server.AddWatchReactor("brokers",
		func(a client_testing.Action) (bool, watch.Interface, error) {
			watchAction := a.(client_testing.WatchAction)
			name, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
			if !found {
				return true, nil, errors.NewNotFound(eventingv1.Resource("broker"), name)
			}
			w := wait.NewFakeWatch(getBrokerDeleteEvents("fooBroker"))
			w.Start()
			return true, w, nil
		})

	err := client.DeleteBroker(context.Background(), brokerName, time.Duration(10)*time.Second)
	assert.NilError(t, err)

	err = client.DeleteBroker(context.Background(), "errorBroker", time.Duration(10)*time.Second)
	assert.ErrorContains(t, err, "errorBroker", time.Duration(10)*time.Second)

	err = client.DeleteBroker(context.Background(), deleted, time.Duration(10)*time.Second)
	assert.ErrorContains(t, err, "marked for deletion")
	assert.ErrorContains(t, err, deleted, time.Duration(10)*time.Second)
}

func TestBrokerList(t *testing.T) {
	serving, client := setup()

	t.Run("broker list returns a list of brokers", func(t *testing.T) {
		broker1 := newBrokerWithGvk("foo1")
		broker2 := newBrokerWithGvk("foo2")

		serving.AddReactor("list", "brokers",
			func(a client_testing.Action) (bool, runtime.Object, error) {
				assert.Equal(t, testNamespace, a.GetNamespace())
				return true, &eventingv1.BrokerList{Items: []eventingv1.Broker{*broker1, *broker2}}, nil
			})

		brokerList, err := client.ListBrokers(context.Background())
		assert.NilError(t, err)
		assert.Assert(t, len(brokerList.Items) == 2)
		assert.Equal(t, brokerList.Items[0].Name, "foo1")
		assert.Equal(t, brokerList.Items[1].Name, "foo2")
		assert.Assert(t, !brokerList.GroupVersionKind().Empty())
		assert.Assert(t, !brokerList.Items[0].GroupVersionKind().Empty())
	})
}

func TestBrokerUpdate(t *testing.T) {
	var name = "broker"
	server, client := setup()

	obj := newBroker(name)
	errorObj := newBroker("error-obj")
	updatedObj := newBrokerWithDeliveryOptions(name)

	server.AddReactor("update", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			name := a.(client_testing.UpdateAction).GetObject().(metav1.Object).GetName()
			if name == "error-obj" {
				return true, nil, fmt.Errorf("error while creating broker %s", name)
			}
			return true, updatedObj, nil
		})
	server.AddReactor("get", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())
			return true, obj, nil
		})

	t.Run("update broker without error", func(t *testing.T) {
		err := client.UpdateBroker(context.Background(), updatedObj)
		assert.NilError(t, err)
	})

	t.Run("create broker with an error returns an error object", func(t *testing.T) {
		err := client.UpdateBroker(context.Background(), errorObj)
		assert.ErrorContains(t, err, "error while creating broker")
	})
}

func TestUpdateBrokerWithRetry(t *testing.T) {
	serving, client := setup()
	var attemptCount, maxAttempts = 0, 5
	serving.AddReactor("get", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			name := a.(client_testing.GetAction).GetName()
			if name == "deletedBroker" {
				broker := newBroker(name)
				now := metav1.Now()
				broker.DeletionTimestamp = &now
				return true, broker, nil
			}
			if name == "getErrorBroker" {
				return true, nil, errors.NewInternalError(fmt.Errorf("mock internal error"))
			}
			return true, newBroker(name), nil
		})

	serving.AddReactor("update", "brokers",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			newBroker := a.(client_testing.UpdateAction).GetObject()
			name := newBroker.(metav1.Object).GetName()

			if name == "testBroker" && attemptCount > 0 {
				attemptCount--
				return true, nil, errors.NewConflict(eventingv1.Resource("broker"), "errorBroker", fmt.Errorf("error updating because of conflict"))
			}
			if name == "errorBroker" {
				return true, nil, errors.NewInternalError(fmt.Errorf("mock internal error"))
			}
			return true, NewBrokerBuilderFromExisting(newBroker.(*eventingv1.Broker)).Build(), nil
		})

	t.Run("Update broker successfully without any retries", func(t *testing.T) {
		err := client.UpdateBrokerWithRetry(context.Background(), "testBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, nil
		}, maxAttempts)
		assert.NilError(t, err, "No retries required as no conflict error occurred")
	})

	t.Run("Update broker with retry after max retries", func(t *testing.T) {
		attemptCount = maxAttempts - 1
		err := client.UpdateBrokerWithRetry(context.Background(), "testBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, nil
		}, maxAttempts)
		assert.NilError(t, err, "Update retried %d times and succeeded", maxAttempts)
		assert.Equal(t, attemptCount, 0)
	})

	t.Run("Update broker with retry and fail with conflict after exhausting max retries", func(t *testing.T) {
		attemptCount = maxAttempts
		err := client.UpdateBrokerWithRetry(context.Background(), "testBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsConflict, "Update retried %d times and failed", maxAttempts)
		assert.Equal(t, attemptCount, 0)
	})

	t.Run("Update broker with retry and fail with conflict after exhausting max retries", func(t *testing.T) {
		attemptCount = maxAttempts
		err := client.UpdateBrokerWithRetry(context.Background(), "testBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsConflict, "Update retried %d times and failed", maxAttempts)
		assert.Equal(t, attemptCount, 0)
	})

	t.Run("Update broker with retry fails with a non conflict error", func(t *testing.T) {
		err := client.UpdateBrokerWithRetry(context.Background(), "errorBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsInternalError)
	})

	t.Run("Update broker with retry fails with resource already deleted error", func(t *testing.T) {
		err := client.UpdateBrokerWithRetry(context.Background(), "deletedBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, nil
		}, maxAttempts)
		assert.ErrorContains(t, err, "marked for deletion")
	})

	t.Run("Update broker with retry fails with error from updateFunc", func(t *testing.T) {
		err := client.UpdateBrokerWithRetry(context.Background(), "testBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, fmt.Errorf("error updating object")
		}, maxAttempts)
		assert.ErrorContains(t, err, "error updating object")
	})

	t.Run("Update broker with retry fails with error from GetBroker", func(t *testing.T) {
		err := client.UpdateBrokerWithRetry(context.Background(), "getErrorBroker", func(broker *eventingv1.Broker) (*eventingv1.Broker, error) {
			return broker, nil
		}, maxAttempts)
		assert.ErrorType(t, err, errors.IsInternalError)
	})
}

func newTrigger(name string) *eventingv1.Trigger {
	return NewTriggerBuilder(name).
		Namespace(testNamespace).
		Broker("default").
		Filters(map[string]string{"type": "foo"}).
		Build()
}

func newBroker(name string) *eventingv1.Broker {
	return NewBrokerBuilder(name).
		Namespace(testNamespace).
		Class("").
		Build()
}

func newBrokerWithGvk(name string) *eventingv1.Broker {
	return NewBrokerBuilder(name).
		Namespace(testNamespace).
		WithGvk().
		Build()
}

func newBrokerWithClass(name string) *eventingv1.Broker {
	return NewBrokerBuilder(name).
		Namespace(testNamespace).
		Class(testClass).
		Build()
}

func newBrokerWithDeliveryOptions(name string) *eventingv1.Broker {
	sink := &duckv1.Destination{
		Ref: &duckv1.KReference{Name: "test-svc", Kind: "Service", APIVersion: "serving.knative.dev/v1", Namespace: "default"},
	}
	testTimeout := "PT10S"
	retry := int32(2)
	policy := v1.BackoffPolicyType("linear")
	return NewBrokerBuilder(name).
		Namespace(testNamespace).
		DlSink(sink).
		Timeout(&testTimeout).
		Retry(&retry).
		BackoffDelay(&testTimeout).
		BackoffPolicy(&policy).
		RetryAfterMax(&testTimeout).
		Build()
}

func newBrokerWithNilDeliveryOptions(name string) *eventingv1.Broker {
	return NewBrokerBuilder(name).
		Namespace(testNamespace).
		DlSink(nil).
		Timeout(nil).
		Retry(nil).
		BackoffDelay(nil).
		BackoffPolicy(nil).
		RetryAfterMax(nil).
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
