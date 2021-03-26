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

package wait

import (
	"context"
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type waitForReadyTestCase struct {
	events           []watch.Event
	timeout          time.Duration
	errorText        string
	messagesExpected []string
}

func TestWaitCancellation(t *testing.T) {
	fakeWatchApi := NewFakeWatch([]watch.Event{})
	fakeWatchApi.Start()
	wfe := NewWaitForEvent("foobar", func(e *watch.Event) bool {
		return false
	})

	timeout := time.Second * 5

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(time.Millisecond * 500)
		cancel()
	}()
	err, _ := wfe.Wait(ctx, fakeWatchApi, "foobar", Options{Timeout: &timeout}, NoopMessageCallback())
	assert.Assert(t, errors.Is(err, context.Canceled))

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(time.Millisecond * 500)
		cancel()
	}()
	wfr := NewWaitForReady(
		"blub",
		func(obj runtime.Object) (apis.Conditions, error) {
			return apis.Conditions(obj.(*servingv1.Service).Status.Conditions), nil
		})
	err, _ = wfr.Wait(ctx, fakeWatchApi, "foobar", Options{Timeout: &timeout}, NoopMessageCallback())
	assert.Assert(t, errors.Is(err, context.Canceled))
}

func TestAddWaitForReady(t *testing.T) {

	for i, tc := range prepareTestCases("test-service") {
		fakeWatchApi := NewFakeWatch(tc.events)

		waitForReady := NewWaitForReady(
			"blub",
			func(obj runtime.Object) (apis.Conditions, error) {
				return apis.Conditions(obj.(*servingv1.Service).Status.Conditions), nil
			})
		fakeWatchApi.Start()
		var msgs []string
		err, _ := waitForReady.Wait(context.Background(), fakeWatchApi, "foobar", Options{Timeout: &tc.timeout}, func(_ time.Duration, msg string) {
			msgs = append(msgs, msg)
		})
		close(fakeWatchApi.eventChan)

		if tc.errorText == "" && err != nil {
			t.Errorf("%d: Error received %v", i, err)
			continue
		}
		if tc.errorText != "" {
			if err == nil {
				t.Errorf("%d: No error but expected one", i)
			} else {
				assert.ErrorContains(t, err, tc.errorText)
			}
		}

		// check messages
		assert.Assert(t, cmp.DeepEqual(tc.messagesExpected, msgs), "%d: Messages expected to be equal", i)

		if fakeWatchApi.StopCalled != 0 {
			t.Errorf("%d: Exactly zero 'stop' should be called, but got %d", i, fakeWatchApi.StopCalled)
		}

	}
}

func TestAddWaitForDelete(t *testing.T) {
	for i, tc := range prepareDeleteTestCases("test-service") {
		fakeWatchAPI := NewFakeWatch(tc.events)

		waitForEvent := NewWaitForEvent(
			"blub",
			func(evt *watch.Event) bool { return evt.Type == watch.Deleted })
		fakeWatchAPI.Start()

		err, _ := waitForEvent.Wait(context.Background(), fakeWatchAPI, "foobar", Options{Timeout: &tc.timeout}, NoopMessageCallback())
		close(fakeWatchAPI.eventChan)

		if tc.errorText == "" && err != nil {
			t.Errorf("%d: Error received %v", i, err)
			continue
		}
		if tc.errorText != "" {
			if err == nil {
				t.Errorf("%d: No error but expected one", i)
			} else {
				assert.ErrorContains(t, err, tc.errorText)
			}
		}

		if fakeWatchAPI.StopCalled != 0 {
			t.Errorf("%d: Exactly zero 'stop' should be called, but got %d", i, fakeWatchAPI.StopCalled)
		}
	}
}

// Test cases which consists of a series of events to send and the expected behaviour.
func prepareTestCases(name string) []waitForReadyTestCase {
	return []waitForReadyTestCase{
		errorTest(name),
		tc(peNormal, name, time.Second, ""),
		tc(peWrongGeneration, name, 1*time.Second, "timeout"),
		tc(peTimeout, name, time.Second, "timeout"),
		tc(peReadyFalseWithinErrorWindow, name, time.Second, ""),
	}
}

func prepareDeleteTestCases(name string) []waitForReadyTestCase {
	return []waitForReadyTestCase{
		tc(deNormal, name, time.Second, ""),
		tc(peTimeout, name, 10*time.Second, "timeout"),
	}
}

func errorTest(name string) waitForReadyTestCase {
	events := []watch.Event{
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", "msg1")},
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionFalse, corev1.ConditionTrue, "FakeError", "Test Error")},
	}

	return waitForReadyTestCase{
		events:           events,
		timeout:          3 * time.Second,
		errorText:        "FakeError",
		messagesExpected: []string{"msg1", "Test Error"},
	}
}

func tc(f func(name string) (evts []watch.Event, nrMessages int), name string, timeout time.Duration, errorTxt string) waitForReadyTestCase {
	events, nrMsgs := f(name)
	return waitForReadyTestCase{
		events,
		timeout,
		errorTxt,
		pMessages(nrMsgs),
	}
}

func pMessages(max int) []string {
	return []string{
		"msg1", "msg2", "msg3", "msg4", "msg5", "msg6",
	}[:max]
}

// =============================================================================

func peNormal(name string) ([]watch.Event, int) {
	messages := pMessages(2)
	return []watch.Event{
		{Type: watch.Added, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0], 1, 2)},
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0], 2, 2)},
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", messages[1], 2, 2)},
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "", 2, 2)},
	}, len(messages)
}

func peTimeout(name string) ([]watch.Event, int) {
	messages := pMessages(1)
	return []watch.Event{
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0])},
	}, len(messages)
}

func peWrongGeneration(name string) ([]watch.Event, int) {
	messages := pMessages(1)
	return []watch.Event{
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0])},
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "", 1, 2)},
	}, len(messages)
}

func peReadyFalseWithinErrorWindow(name string) ([]watch.Event, int) {
	messages := pMessages(1)
	return []watch.Event{
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionFalse, corev1.ConditionFalse, "Route not ready", messages[0])},
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "Route ready", "")},
	}, len(messages)
}

func deNormal(name string) ([]watch.Event, int) {
	messages := pMessages(2)
	return []watch.Event{
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0])},
		{Type: watch.Modified, Object: CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", messages[1])},
		{Type: watch.Deleted, Object: CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "")},
	}, len(messages)
}
