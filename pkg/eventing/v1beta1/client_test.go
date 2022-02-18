// Copyright © 2022 The Knative Authors
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

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/eventing/pkg/apis/eventing/v1beta1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1beta1/fake"
	"knative.dev/pkg/apis"
)

const (
	testNamespace = "test-ns"
	testBroker    = "test-broker"
	testSource    = "test.source"
	testType      = "test-type"
	testName      = "test-eventtype"
	errName       = "error-eventtype"
)

func setup(ns string) (fakeSvr fake.FakeEventingV1beta1, client KnEventingV1Beta1Client) {
	fakeE := fake.FakeEventingV1beta1{Fake: &client_testing.Fake{}}
	cli := NewKnEventingV1Beta1Client(&fakeE, ns)
	return fakeE, cli
}

func TestNamespace(t *testing.T) {
	_, client := setup(testNamespace)
	assert.Equal(t, testNamespace, client.Namespace())
}

func TestBuilder(t *testing.T) {
	et := newEventtypeWithSourceBroker(testName, testSource, testBroker)
	assert.Equal(t, et.Name, testName)
	assert.Equal(t, et.Spec.Broker, testBroker)
	source := et.Spec.Source
	assert.Assert(t, source != nil)
	assert.Equal(t, source.String(), testSource)
}

func TestKnEventingV1Beta1Client_CreateEventtype(t *testing.T) {
	server, client := setup(testNamespace)

	server.AddReactor("create", "eventtypes",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())

			name := a.(client_testing.CreateAction).GetObject().(metav1.Object).GetName()
			if name == errName {
				return true, nil, fmt.Errorf("error while creating eventtype %s", name)
			}
			return true, nil, nil
		})
	ctx := context.Background()

	t.Run("create eventtype successfully", func(t *testing.T) {
		objNew := newEventtype(testName)
		err := client.CreateEventtype(ctx, objNew)
		assert.NilError(t, err)
	})
	t.Run("create eventtype with source and broker successfully", func(t *testing.T) {
		objNew := newEventtypeWithSourceBroker(testName, testSource, testBroker)
		err := client.CreateEventtype(ctx, objNew)
		assert.NilError(t, err)
	})
	t.Run("create eventtype with error", func(t *testing.T) {
		objNew := newEventtype(errName)
		err := client.CreateEventtype(ctx, objNew)
		assert.ErrorContains(t, err, "error while creating eventtype")
	})
}

func TestKnEventingV1Beta1Client_DeleteEventtype(t *testing.T) {
	server, client := setup(testNamespace)

	server.AddReactor("delete", "eventtypes",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())

			name := a.(client_testing.DeleteAction).GetName()
			if name == errName {
				return true, nil, fmt.Errorf("error while deleting eventtype %s", name)
			}
			return true, nil, nil
		})
	ctx := context.Background()

	t.Run("delete eventtype successfully", func(t *testing.T) {
		err := client.DeleteEventtype(ctx, testName)
		assert.NilError(t, err)
	})
	t.Run("delete eventtype with error", func(t *testing.T) {
		err := client.DeleteEventtype(ctx, errName)
		assert.ErrorContains(t, err, "error while deleting eventtype")
	})
}

func TestKnEventingV1Beta1Client_GetEventtype(t *testing.T) {
	server, client := setup(testNamespace)

	server.AddReactor("get", "eventtypes",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())

			name := a.(client_testing.GetAction).GetName()
			if name == errName {
				return true, nil, fmt.Errorf("error while getting eventtype %s", name)
			}
			return true, newEventtype(testName), nil
		})
	ctx := context.Background()

	t.Run("get eventtype successfully", func(t *testing.T) {
		et, err := client.GetEventtype(ctx, testName)
		assert.NilError(t, err)
		assert.Equal(t, et.Name, testName)
	})
	t.Run("get eventtype with error", func(t *testing.T) {
		_, err := client.GetEventtype(ctx, errName)
		assert.ErrorContains(t, err, "error while getting eventtype")
	})
}

func TestKnEventingV1Beta1Client_ListEventtypes(t *testing.T) {
	server, client := setup(testNamespace)

	server.AddReactor("list", "eventtypes",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			assert.Equal(t, testNamespace, a.GetNamespace())

			return true, &v1beta1.EventTypeList{Items: []v1beta1.EventType{
				*newEventtype("eventtype-1"),
				*newEventtype("eventtype-2")}}, nil
		})
	ctx := context.Background()

	list, err := client.ListEventtypes(ctx)
	assert.NilError(t, err)
	assert.Assert(t, list != nil)
	assert.Equal(t, len(list.Items), 2)
	assert.Equal(t, list.Items[0].Name, "eventtype-1")
	assert.Equal(t, list.Items[1].Name, "eventtype-2")
}

func newEventtypeWithSourceBroker(name string, source string, broker string) *v1beta1.EventType {
	url, _ := apis.ParseURL(source)
	return NewEventtypeBuilder(name).
		Namespace(testNamespace).
		WithGvk().
		Type(testType).
		Source(url).
		Broker(broker).
		Build()
}

func newEventtype(name string) *v1beta1.EventType {
	return NewEventtypeBuilder(name).
		Namespace(testNamespace).
		Type(testType).
		Build()
}
