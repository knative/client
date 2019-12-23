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
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"knative.dev/pkg/apis"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

type waitForReadyTestCase struct {
	events           []watch.Event
	timeout          time.Duration
	errorExpected    bool
	messagesExpected []string
}

func TestAddWaitForReady(t *testing.T) {

	for i, tc := range prepareTestCases("test-service") {
		fakeWatchApi := NewFakeWatch(tc.events)

		waitForReady := NewWaitForReady(
			"blub",
			func(name string, timeout time.Duration) (watch.Interface, error) {
				return fakeWatchApi, nil
			},
			func(obj runtime.Object) (apis.Conditions, error) {
				return apis.Conditions(obj.(*v1alpha1.Service).Status.Conditions), nil
			})
		fakeWatchApi.Start()
		var msgs []string
		err, _ := waitForReady.Wait("foobar", tc.timeout, func(_ time.Duration, msg string) {
			msgs = append(msgs, msg)
		})
		close(fakeWatchApi.eventChan)

		if !tc.errorExpected && err != nil {
			t.Errorf("%d: Error received %v", i, err)
			continue
		}
		if tc.errorExpected && err == nil {
			t.Errorf("%d: No error but expected one", i)
		}

		// check messages
		assert.Assert(t, cmp.DeepEqual(tc.messagesExpected, msgs), "%d: Messages expected to be equal", i)

		if fakeWatchApi.StopCalled != 1 {
			t.Errorf("%d: Exactly one 'stop' should be called, but got %d", i, fakeWatchApi.StopCalled)
		}

	}
}

// Test cases which consists of a series of events to send and the expected behaviour.
func prepareTestCases(name string) []waitForReadyTestCase {
	return []waitForReadyTestCase{
		tc(peNormal, name, false),
		tc(peError, name, true),
		tc(peWrongGeneration, name, true),
		tc(peTimeout, name, true),
	}
}

func tc(f func(name string) (evts []watch.Event, nrMessages int), name string, isError bool) waitForReadyTestCase {
	events, nrMsgs := f(name)
	return waitForReadyTestCase{
		events,
		time.Second,
		isError,
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
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0])},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", messages[1])},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "")},
	}, len(messages)
}

func peError(name string) ([]watch.Event, int) {
	messages := pMessages(1)
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0])},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionFalse, corev1.ConditionTrue, "FakeError", "")},
	}, len(messages)
}

func peTimeout(name string) ([]watch.Event, int) {
	messages := pMessages(1)
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0])},
	}, len(messages)
}

func peWrongGeneration(name string) ([]watch.Event, int) {
	messages := pMessages(1)
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", messages[0])},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "", 1, 2)},
	}, len(messages)
}
