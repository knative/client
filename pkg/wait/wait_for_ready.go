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
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"knative.dev/pkg/apis"
)

// Callbacks and configuration used while waiting
type waitForReadyConfig struct {
	watchMaker          WatchMaker
	conditionsExtractor ConditionsExtractor
	kind                string
}

// Interface used for waiting of a resource of a given name to reach a definitive
// state in its "Ready" condition.
type WaitForReady interface {

	// Wait on resource the resource with this name until a given timeout
	// and write event messages for unknown event to the status writer
	Wait(name string, timeout time.Duration, msgCallback MessageCallback) (error, time.Duration)
}

// Utility alias for "A function like Watch from the client"
type WatchFunc func(opts v1.ListOptions) (watch.Interface, error)

// Utility alias for "A function that takes the name and just gets the resource"
type GetFunc func(name string) (runtime.Object, error)

// Create watch which is used when waiting for Ready condition
type WatchMaker func(name string, timeout time.Duration) (watch.Interface, error)

// Extract conditions from a runtime object
type ConditionsExtractor func(obj runtime.Object) (apis.Conditions, error)

// Callback for event messages
type MessageCallback func(durationSinceState time.Duration, message string)

// Constructor with resource type specific configuration
func NewWaitForReady(kind string, watchMaker WatchMaker, extractor ConditionsExtractor) WaitForReady {
	return &waitForReadyConfig{
		kind:                kind,
		watchMaker:          watchMaker,
		conditionsExtractor: extractor,
	}
}

// A simple message callback which prints out messages line by line
func SimpleMessageCallback(out io.Writer) MessageCallback {
	oldMessage := ""
	return func(duration time.Duration, message string) {
		txt := message
		if message == oldMessage {
			txt = "..."
		}
		fmt.Fprintf(out, "%7.3fs %s\n", float64(duration.Round(time.Millisecond))/float64(time.Second), txt)
		oldMessage = message
	}
}

// Noop-callback
func NoopMessageCallback() MessageCallback {
	return func(durationSinceState time.Duration, message string) {}
}

// Wait until a resource enters condition of type "Ready" to "False" or "True".
// `watchFunc` creates the actual watch, `kind` is the type what your are watching for
// (e.g. "service"), `timeout` is a timeout after which the watch should be cancelled if no
// target state has been entered yet and `out` is used for printing out status messages
// msgCallback gets called for every event with an 'Ready' condition == UNKNOWN with the event's message.
func (w *waitForReadyConfig) Wait(name string, timeout time.Duration, msgCallback MessageCallback) (error, time.Duration) {

	floatingTimeout := timeout
	for {
		start := time.Now()
		retry, timeoutReached, err := w.waitForReadyCondition(start, name, floatingTimeout, msgCallback)
		if err != nil {
			return err, time.Since(start)
		}
		floatingTimeout = floatingTimeout - time.Since(start)
		if timeoutReached || floatingTimeout < 0 {
			return fmt.Errorf("timeout: %s '%s' not ready after %d seconds", w.kind, name, int(timeout/time.Second)), time.Since(start)
		}

		if retry {
			// restart loop
			continue
		}
		return nil, time.Since(start)
	}
}

func addWatchTimeout(opts *v1.ListOptions, timeout time.Duration) {
	if timeout == 0 {
		return
	}
	// Wait for service to enter 'Ready' state, with a timeout of which is slightly larger than
	// the provided timeout. We have our own timeout which fires after "timeout" seconds
	// and stops the watch
	timeOutWatchSeconds := int64((timeout + 30*time.Second) / time.Second)
	opts.TimeoutSeconds = &timeOutWatchSeconds
}

func (w *waitForReadyConfig) waitForReadyCondition(start time.Time, name string, timeout time.Duration, msgCallback MessageCallback) (retry bool, timeoutReached bool, err error) {

	watcher, err := w.watchMaker(name, timeout)
	if err != nil {
		return false, false, err
	}

	defer watcher.Stop()
	for {
		select {
		case <-time.After(timeout):
			return false, true, nil
		case event, ok := <-watcher.ResultChan():
			if !ok || event.Object == nil {
				return true, false, nil
			}

			// Skip event if generations has not yet been consolidated
			inSync, err := isGivenEqualsObservedGeneration(event.Object)
			if err != nil {
				return false, false, err
			}
			if !inSync {
				continue
			}

			conditions, err := w.conditionsExtractor(event.Object)
			if err != nil {
				return false, false, err
			}
			for _, cond := range conditions {
				if cond.Type == apis.ConditionReady {
					switch cond.Status {
					case corev1.ConditionTrue:
						return false, false, nil
					case corev1.ConditionFalse:
						return false, false, fmt.Errorf("%s: %s", cond.Reason, cond.Message)
					}
					if cond.Message != "" {
						msgCallback(time.Since(start), cond.Message)
					}
				}
			}
		}
	}
}

// Going over Unstructured to keep that function generally applicable.
// Alternative implemenentation: Add a func-field to waitForReadyConfig which has to be
// provided for every resource (like the conditions extractor)
func isGivenEqualsObservedGeneration(object runtime.Object) (bool, error) {
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return false, err
	}
	meta, ok := unstructured["metadata"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("cannot extract metadata from %v", object)
	}
	status, ok := unstructured["status"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("cannot extract status from %v", object)
	}
	observedGeneration, ok := status["observedGeneration"]
	if !ok {
		// Can be the case if not status has been attached yet
		return false, nil
	}
	givenGeneration, ok := meta["generation"]
	if !ok {
		return false, fmt.Errorf("no field 'generation' in metadata of %v", object)
	}
	return givenGeneration == observedGeneration, nil
}
