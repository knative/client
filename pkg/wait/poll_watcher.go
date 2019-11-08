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
	"sync"
	"time"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"knative.dev/serving/pkg/client/clientset/versioned/scheme"
)

type pollingWatcher struct {
	c        rest.Interface
	ns       string
	resource string
	name     string
	timeout  time.Duration
	done     chan bool
	result   chan watch.Event
	wg       *sync.WaitGroup
}

type WatchFunc func(v1.ListOptions) (watch.Interface, error)

// NewWatcher makes a watch.Interface on the given resource in the client,
// falling back to polling if the server does not support Watch.
func NewWatcher(watchFunc WatchFunc, c rest.Interface, ns string, resource string, name string, timeout time.Duration) (watch.Interface, error) {
	native, err := nativeWatch(watchFunc, c, ns, resource, name, timeout)
	if err == nil {
		return native, nil
	}
	polling := &pollingWatcher{c, ns, resource, name, timeout, make(chan bool), make(chan watch.Event), &sync.WaitGroup{}}
	err = polling.start()
	if err != nil {
		return nil, err
	}
	return polling, nil
}

func (w *pollingWatcher) poll() (runtime.Object, error) {
	return w.c.Get().
		Namespace(w.ns).
		Resource(w.resource).
		Name(w.name).
		Do().
		Get()
}

func (w *pollingWatcher) start() error {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		var err error
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var old, new runtime.Object
		done := false
		for !done {
			old = new

			select {
			case <-ticker.C:
				new, err = w.poll()
				newObj, ok1 := new.(v1.Object)
				oldObj, ok2 := old.(v1.Object)

				if old != nil && err != nil && api_errors.IsNotFound(err) {
					// Deleted
					w.result <- watch.Event{
						Type:   watch.Deleted,
						Object: old,
					}
				} else if err != nil {
					// Just an error
					w.result <- watch.Event{
						Type: watch.Error,
					}
				} else if old == nil && new != nil {
					// Added
					w.result <- watch.Event{
						Type:   watch.Added,
						Object: new,
					}
				} else if !(ok1 && ok2) {
					// Error wrong types
					w.result <- watch.Event{
						Type: watch.Error,
					}
				} else if newObj.GetUID() != oldObj.GetUID() {
					// Deleted and readded.
					w.result <- watch.Event{
						Type:   watch.Deleted,
						Object: old,
					}
					w.result <- watch.Event{
						Type:   watch.Added,
						Object: new,
					}
				} else if newObj.GetResourceVersion() != oldObj.GetResourceVersion() {
					// Modified.
					w.result <- watch.Event{
						Type:   watch.Modified,
						Object: new,
					}
				}
			case done = <-w.done:
				break
			}
		}
	}()
	return nil
}

func (w *pollingWatcher) ResultChan() <-chan watch.Event {
	return w.result
}

func (w *pollingWatcher) Stop() {
	w.done <- true
	w.wg.Wait()
	close(w.result)
	close(w.done)
}

func nativeWatch(watchFunc WatchFunc, c rest.Interface, ns string, resource string, name string, timeout time.Duration) (watch.Interface, error) {
	opts := v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
	}
	opts.Watch = true
	addWatchTimeout(&opts, timeout)

	if watchFunc != nil {
		return watchFunc(opts)
	}
	return c.Get().
		Namespace(ns).
		Resource(resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
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
