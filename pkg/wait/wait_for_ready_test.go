package wait

import (
	"bytes"
	"github.com/knative/pkg/apis"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"strings"
	"testing"
)

type fakeWatch struct {
	eventChan  chan watch.Event
	events     []watch.Event
	stopCalled int
}

func newFakeWatch(events []watch.Event) *fakeWatch {
	return &fakeWatch{
		eventChan: make(chan watch.Event),
		events:    events,
	}
}

func (f *fakeWatch) Stop() {
	f.stopCalled++
}

func (f *fakeWatch) start() {
	go f.pumpEvents()
}

func (f *fakeWatch) ResultChan() <-chan watch.Event {
	return f.eventChan
}

func (f *fakeWatch) pumpEvents() {
	for _, ev := range f.events {
		f.eventChan <- ev
	}
}

type waitForReadyTestCase struct {
	events         []watch.Event
	timeout        int
	errorExpected  bool
	messageContent []string
}

func TestAddWaitForReady(t *testing.T) {

	for i, tc := range prepareTestCases() {
		fakeWatchApi := newFakeWatch(tc.events)
		outBuffer := new(bytes.Buffer)

		waitForReady := NewWaitForReady(
			"blub",
			func(opts v1.ListOptions) (watch.Interface, error) {
				return fakeWatchApi, nil
			},
			func(obj runtime.Object) (apis.Conditions, error) {
				println("Extract called")
				return apis.Conditions(obj.(*v1alpha1.Service).Status.Conditions), nil
			})
		fakeWatchApi.start()
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

		if fakeWatchApi.stopCalled != 1 {
			t.Errorf("%d: Exactly one 'stop' should be called, but got %d", i, fakeWatchApi.stopCalled)
		}

	}
}

// Test cases which consists of a series of events to send and the expected behaviour.
func prepareTestCases() []waitForReadyTestCase {
	return []waitForReadyTestCase{
		{peNormal(), 1, false, []string{"OK", "foobar", "blub"}},
		{peError(), 1, true, []string{"FakeError"}},
		{peTimeout(), 1, true, []string{"timeout"}},
		{peWrongGeneration(), 1, true, []string{"timeout"}},
	}
}

// =============================================================================

func peNormal() []watch.Event {
	return []watch.Event{
		{watch.Added, createThinService(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, createThinService(corev1.ConditionUnknown, corev1.ConditionTrue, "")},
		{watch.Modified, createThinService(corev1.ConditionTrue, corev1.ConditionTrue, "")},
	}
}

func peError() []watch.Event {
	return []watch.Event{
		{watch.Added, createThinService(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, createThinService(corev1.ConditionFalse, corev1.ConditionTrue, "FakeError")},
	}
}

func peTimeout() []watch.Event {
	return []watch.Event{
		{watch.Added, createThinService(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
	}
}

func peWrongGeneration() []watch.Event {
	return []watch.Event{
		{watch.Added, createThinService(corev1.ConditionUnknown, corev1.ConditionUnknown, "")},
		{watch.Modified, createThinService(corev1.ConditionTrue, corev1.ConditionTrue, "", 1, 2)},
	}
}

func createThinService(readyStatus corev1.ConditionStatus, otherReadyStatus corev1.ConditionStatus, reason string, generations ...int64) runtime.Object {
	service := v1alpha1.Service{}
	if len(generations) == 2 {
		service.Generation = generations[0]
		service.Status.ObservedGeneration = generations[1]
	} else {
		service.Generation = 1
		service.Status.ObservedGeneration = 1
	}
	service.Status.Conditions = []apis.Condition{
		{Type: "RoutesReady", Status: otherReadyStatus},
		{Type: apis.ConditionReady, Status: readyStatus, Reason: reason},
		{Type: "ConfigurationsReady", Status: otherReadyStatus},
	}
	return &service
}
