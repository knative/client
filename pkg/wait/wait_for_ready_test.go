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
	"bytes"
	"strings"
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
	events         []watch.Event
	timeout        time.Duration
	errorExpected  bool
	messageContent []string
}

func TestAddWaitForReady(t *testing.T) {

	for i, tc := range prepareTestCases() {
		fakeWatchApi := NewFakeWatch(tc.events)
		outBuffer := new(bytes.Buffer)

		waitForReady := NewWaitForReady(
			"blub",
			func(opts v1.ListOptions) (watch.Interface, error) {
				return fakeWatchApi, nil
			},
			func(obj runtime.Object) (apis.Conditions, error) {
				return apis.Conditions(obj.(*v1alpha1.Service).Status.Conditions), nil
			})
		fakeWatchApi.Start()
		err := waitForReady.Wait("foobar", tc.timeout, outBuffer)
		close(fakeWatchApi.eventChan)

		if !tc.errorExpected && err != nil {
			t.Errorf("%d: Error received %v", i, err)
			continue
		}
		if tc.errorExpected && err == nil {
			t.Errorf("%d: No error but expected one", i)
		}
		txtToCheck := outBuffer.String()
		if err != nil {
			txtToCheck = err.Error()
		}

		for _, msg := range tc.messageContent {
			if !strings.Contains(txtToCheck, msg) {
				t.Errorf("%d: '%s' does not contain expected part %s", i, txtToCheck, msg)
			}
		}

		if fakeWatchApi.StopCalled != 1 {
			t.Errorf("%d: Exactly one 'stop' should be called, but got %d", i, fakeWatchApi.StopCalled)
		}

	}
}

// Test cases which consists of a series of events to send and the expected behaviour.
func prepareTestCases() []waitForReadyTestCase {
	return []waitForReadyTestCase{
		{peNormal(), time.Second, false, []string{"OK", "foobar", "blub"}},
		{peError(), time.Second, true, []string{"FakeError"}},
		{peTimeout(), time.Second, true, []string{"timeout"}},
		{peWrongGeneration(), time.Second, true, []string{"timeout"}},
	}
}

// =============================================================================

func peNormal() []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, CreateTestServiceWithConditions(corev1.ConditionUnknown, corev1.ConditionTrue, "")},
		{watch.Modified, CreateTestServiceWithConditions(corev1.ConditionTrue, corev1.ConditionTrue, "")},
	}
}

func peError() []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, CreateTestServiceWithConditions(corev1.ConditionFalse, corev1.ConditionTrue, "FakeError")},
	}
}

func peTimeout() []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
	}
}

func peWrongGeneration() []watch.Event {
	return []watch.Event{
		{watch.Added, CreateTestServiceWithConditions(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, CreateTestServiceWithConditions(corev1.ConditionTrue, corev1.ConditionTrue, "", 1, 2)},
	}
}
