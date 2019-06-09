package wait

import (
	"fmt"
	"github.com/knative/pkg/apis"
	"io"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"time"
)

// Callbacks and configuration used while waiting
type waitForReadyConfig struct {
	watchFunc           WatchFunc
	conditionsExtractor ConditionsExtractor
	kind                string
}

// Interface used for waiting of a resource of a given name to reach a definitive
// state in its "Ready" condition.
type WaitForReady interface {

	// Wait on resource the resource with this name until a given timeout
	// and write status out on writer
	Wait(name string, timeout int, out io.Writer) error
}

// Create watch which is used when waiting for Ready condition
type WatchFunc func(opts v1.ListOptions) (watch.Interface, error)

// Extract conditions from a runtime object
type ConditionsExtractor func(obj runtime.Object) (apis.Conditions, error)

// Constructor with resource type specific configuration
func NewWaitForReady(kind string, watchFunc WatchFunc, extractor ConditionsExtractor) WaitForReady {
	return &waitForReadyConfig{
		kind:                kind,
		watchFunc:           watchFunc,
		conditionsExtractor: extractor,
	}
}

// Wait until a resource enters condition of type "Ready" to "False" or "True".
// `watchFunc` creates the actual watch, `kind` is the type what your are watching for
// (e.g. "service"), `timeout` is a timeout after which the watch should be cancelled if no
// target state has been entered yet and `out` is used for printing out status messages
func (w *waitForReadyConfig) Wait(name string, timeout int, out io.Writer) error {
	opts := v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
	}
	addWatchTimeout(&opts, timeout)

	watcher, err := w.watchFunc(opts)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Waiting for %s '%s' to become ready ... ", w.kind, name)
	flush(out)

	if err := w.waitForReadyCondition(watcher, name, timeout); err != nil {
		fmt.Fprintln(out)
		return err
	}
	fmt.Fprintln(out, "OK")
	return nil
}

func addWatchTimeout(opts *v1.ListOptions, timeout int) {
	if timeout == 0 {
		return
	}
	// Wait for service to enter 'Ready' state, with a timeout of which is slightly larger than
	// the provided timeout. We have our own timeout which fires after "timeout" seconds
	// and stops the watch
	timeOutWatch := int64(timeout + 30)
	opts.TimeoutSeconds = &timeOutWatch
}

// Duck type for writers having a flush
type flusher interface {
	Flush() error
}

func flush(out io.Writer) {
	if flusher, ok := out.(flusher); ok {
		flusher.Flush()
	}
}

func (w *waitForReadyConfig) waitForReadyCondition(watcher watch.Interface, name string, timeout int) error {
	defer watcher.Stop()
	for {
		select {
		case <-time.After(time.Duration(timeout) * time.Second):
			return fmt.Errorf("timeout: %s '%s' not ready after %d seconds", w.kind, name, timeout)
		case event, ok := <-watcher.ResultChan():
			if !ok || event.Object == nil {
				return fmt.Errorf("timeout while waiting for %s '%s' to become ready", w.kind, name)
			}

			// Skip event if generations has not yet been consolidated
			inSync, err := isGivenEqualsObservedGeneration(event.Object)
			if err != nil {
				return err
			}
			if !inSync {
				continue
			}

			conditions, err := w.conditionsExtractor(event.Object)
			if err != nil {
				return err
			}
			for _, cond := range conditions {
				fmt.Printf("%v\n", cond)
				if cond.Type == apis.ConditionReady {
					switch cond.Status {
					case corev1.ConditionTrue:
						return nil
					case corev1.ConditionFalse:
						return fmt.Errorf("%s: %s", cond.Reason, cond.Message)
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
