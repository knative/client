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

	"github.com/knative/pkg/apis"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type waitForReadyTestCase struct {
	events        []watch.Event
	timeout       time.Duration
	errorExpected bool
}

func TestAddWaitForReady(t *testing.T) {

	for i, tc := range prepareTestCases("test-service") {
		fakeWatchApi := NewFakeWatch(tc.events)

		waitForReady := NewWaitForReady(
			"blub",
			func(opts v1.ListOptions) (watch.Interface, error) {
				return fakeWatchApi, nil
			},
			func(obj runtime.Object) (apis.Conditions, error) {
				return apis.Conditions(obj.(*v1alpha1.Service).Status.Conditions), nil
			})
		fakeWatchApi.Start()
		err := waitForReady.Wait("foobar", tc.timeout)
		close(fakeWatchApi.eventChan)

		if !tc.errorExpected && err != nil {
			t.Errorf("%d: Error received %v", i, err)
			continue
		}
		if tc.errorExpected && err == nil {
			t.Errorf("%d: No error but expected one", i)
		}

		if fakeWatchApi.StopCalled != 1 {
			t.Errorf("%d: Exactly one 'stop' should be called, but got %d", i, fakeWatchApi.StopCalled)
		}

	}
}

// Test cases which consists of a series of events to send and the expected behaviour.
func prepareTestCases(name string) []waitForReadyTestCase {
	return []waitForReadyTestCase{
		{peNormal(name), time.Second, false},
		{peError(name), time.Second, true},
		{peTimeout(name), time.Second, true},
		{peWrongGeneration(name), time.Second, true},
	}
}

// =============================================================================

func peNormal(name string) []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "")},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "")},
	}
}

func peError(name string) []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionFalse, corev1.ConditionTrue, "FakeError")},
	}
}

func peTimeout(name string) []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
	}
}

func peWrongGeneration(name string) []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", 1, 2)},
	}
}
