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
	"fmt"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"knative.dev/pkg/apis"
)

// Callbacks and configuration used while waiting
type waitForReadyConfig struct {
	conditionsExtractor ConditionsExtractor
	kind                string
}

// Callbacks and configuration used while waiting for event
type waitForEvent struct {
	eventDone EventDone
	kind      string
}

// EventDone is a marker to stop actual waiting on given event state
type EventDone func(ev *watch.Event) bool

// Interface used for waiting of a resource of a given name to reach a definitive
// state in its "Ready" condition.
type Wait interface {

	// Wait on resource the resource with this name
	// and write event messages for unknown event to the status writer.
	// Returns an error (if any) and the overall time it took to wait
	Wait(ctx context.Context, watcher watch.Interface, name string, options Options, msgCallback MessageCallback) (error, time.Duration)
}

type Options struct {
	// Window for how long a ReadyCondition == false has to stay
	// for being considered as an error (useful for flaky reconciliation
	ErrorWindow *time.Duration

	// Timeout for how long to wait at maximum
	Timeout *time.Duration
}

// Create watch which is used when waiting for Ready condition
type WatchMaker func(name string, timeout time.Duration) (watch.Interface, error)

// Extract conditions from a runtime object
type ConditionsExtractor func(obj runtime.Object) (apis.Conditions, error)

// Callback for event messages
type MessageCallback func(durationSinceState time.Duration, message string)

// NewWaitForReady waits until the condition is set to Ready == True
func NewWaitForReady(kind string, extractor ConditionsExtractor) Wait {
	return &waitForReadyConfig{
		kind:                kind,
		conditionsExtractor: extractor,
	}
}

// NewWaitForEvent creates a Wait object which waits until a specific event (i.e. when
// the EventDone function returns true)
func NewWaitForEvent(kind string, eventDone EventDone) Wait {
	return &waitForEvent{
		kind:      kind,
		eventDone: eventDone,
	}
}

// SimpleMessageCallback returns a callback which prints out a simple event message to a given writer
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

// NoopMessageCallback is callback which does nothing
func NoopMessageCallback() MessageCallback {
	return func(durationSinceState time.Duration, message string) {}
}

// Wait until a resource enters condition of type "Ready" to "False" or "True".
// `watchFunc` creates the actual watch, `kind` is the type what your are watching for
// (e.g. "service"), `timeout` is a timeout after which the watch should be cancelled if no
// target state has been entered yet and `out` is used for printing out status messages
// msgCallback gets called for every event with an 'Ready' condition == UNKNOWN with the event's message.
func (w *waitForReadyConfig) Wait(ctx context.Context, watcher watch.Interface, name string, options Options, msgCallback MessageCallback) (error, time.Duration) {

	timeout := options.timeoutWithDefault()
	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()
	for {
		start := time.Now()
		retry, timeoutReached, err := w.waitForReadyCondition(ctx, watcher, start, timeoutTimer, options.errorWindowWithDefault(), msgCallback)

		if err != nil {
			return err, time.Since(start)
		}
		if timeoutReached {
			return fmt.Errorf("timeout: %s '%s' not ready after %d seconds", w.kind, name, int(timeout/time.Second)), time.Since(start)
		}

		if retry {
			// sleep to prevent CPU pegging and restart the loop
			time.Sleep(pollInterval)
			continue
		}
		return nil, time.Since(start)
	}
}

// waitForReadyCondition waits until the status condition "Ready" is set to true (good path) or return an error
// when the "Ready" condition is set to false. An error is also returned when the given timeout is reached (plus the
// return value of timeoutReached is set to true in this case).
// An errorWindow can be specified which takes into account of intermediate "false" ready conditions. So before returning
// an error, this methods waits for the errorWindow duration and if an "True" or "Unknown" event arrives in the meantime
// for the "Ready" condition, then the method continues to wait.
func (w *waitForReadyConfig) waitForReadyCondition(
	ctx context.Context, watcher watch.Interface, start time.Time, timeoutTimer *time.Timer, errorWindow time.Duration, msgCallback MessageCallback,
) (retry bool, timeoutReached bool, err error) {

	// channel used to transport the error that has been received
	errChan := make(chan error)

	var errorTimer *time.Timer
	// Stop error timer if it has been started because of
	// a ConditionReady has been set to false
	defer (func() {
		if errorTimer != nil {
			errorTimer.Stop()
			errorTimer = nil
		}
	})()

	for {
		select {
		case <-ctx.Done():
			return false, false, ctx.Err()
		case <-timeoutTimer.C:
			// We reached a timeout without receiving a "Ready" == "True" event
			return false, true, nil
		case err = <-errChan:
			// The error timer fired and we have not received a recovery event ("True" / "Unknown") in the
			// meantime. So the error status is considered to be final.
			return false, false, err
		case event, ok := <-watcher.ResultChan():
			if !ok || event.Object == nil {
				// retry only if the channel is still open
				return ok, false, nil
			}

			// Check whether resource is in sync already (meta.generation == status.observedGeneration)
			inSync, err := generationCheck(event.Object)
			if err != nil {
				return false, false, err
			}

			// Skip events if generations has not yet been consolidated, regardless of type.
			// Wait for the next event to come in until the generations align
			if !inSync {
				continue
			}

			// Skip event if its not a MODIFIED event, as only MODIFIED events update the condition
			// we are looking for.
			// This will filter out all synthetic ADDED events that created bt the API server for
			// the initial state. See https://kubernetes.io/docs/reference/using-api/api-concepts/#the-resourceversion-parameter
			// for details:
			// "Get State and Start at Most Recent: Start a watch at the most recent resource version,
			//  which must be consistent (i.e. served from etcd via a quorum read). To establish initial state,
			//  the watch begins with synthetic “Added” events of all resources instances that exist at the starting
			//  resource version. All following watch events are for all changes that occurred after the resource
			//  version the watch started at."
			if event.Type != watch.Modified {
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
						// Any error timer running will be cancelled by the defer method that has been set above
						return false, false, nil
					case corev1.ConditionFalse:
						// Fire up a timer waiting for the error window duration to still allow to reconcile
						// to a true condition even after the condition went to false. If this is not the case within
						// this window, then an error is returned.
						// If there is already a timer running, we just log.
						if errorTimer == nil {
							err := fmt.Errorf("%s: %s", cond.Reason, cond.Message)
							errorTimer = time.AfterFunc(errorWindow, func() {
								errChan <- err
							})
						}
					case corev1.ConditionUnknown:
						// If an errorTimer is triggered because of a previous "False" event, but now
						// we received an "Unknown" event during the error window, cancel the error timer
						// to avoid to receive an error signal.
						if errorTimer != nil {
							errorTimer.Stop()
							errorTimer = nil
						}
					}
					if cond.Message != "" {
						msgCallback(time.Since(start), cond.Message)
					}
				}
			}
		}
	}
}

// Wait until the expected EventDone is satisfied
func (w *waitForEvent) Wait(ctx context.Context, watcher watch.Interface, name string, options Options, msgCallback MessageCallback) (error, time.Duration) {
	timeout := options.timeoutWithDefault()
	start := time.Now()
	// channel used to transport the error
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err(), time.Since(start)
		case <-timer.C:
			return fmt.Errorf("timeout: %s '%s' not ready after %d seconds", w.kind, name, int(timeout/time.Second)), time.Since(start)
		case event := <-watcher.ResultChan():
			if w.eventDone(&event) {
				return nil, time.Since(start)
			}
		}
	}
}

func generationCheck(object runtime.Object) (bool, error) {
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

func (o Options) timeoutWithDefault() time.Duration {
	if o.Timeout != nil {
		return *o.Timeout
	}
	return 60 * time.Second
}

func (o Options) errorWindowWithDefault() time.Duration {
	if o.ErrorWindow != nil {
		return *o.ErrorWindow
	}
	return 2 * time.Second
}
