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

type PollingWatcher struct {
	c        rest.Interface
	ns       string
	resource string
	name     string
	timeout  time.Duration
	done     chan bool
	result   chan watch.Event
	wg       *sync.WaitGroup
}

func NewWatcher(c rest.Interface, ns string, resource string, name string, timeout time.Duration) (watch.Interface, error) {
	native, err := nativeWatch(c, ns, resource, name, timeout)
	if err == nil {
		return native, nil
	}
	fmt.Println("falling back to polling")
	polling := &PollingWatcher{c, ns, resource, name, timeout, make(chan bool), make(chan watch.Event), &sync.WaitGroup{}}
	err = polling.start()
	if err != nil {
		return nil, err
	}
	return polling, nil
}

func (w *PollingWatcher) poll() (runtime.Object, error) {
	return w.c.Get().
		Namespace(w.ns).
		Resource(w.resource).
		Name(w.name).
		Do().
		Get()
}

func nativeWatch(c rest.Interface, ns string, resource string, name string, timeout time.Duration) (watch.Interface, error) {
	opts := v1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("metadata.name", name).String(),
	}
	opts.Watch = true
	addWatchTimeout(&opts, timeout)

	return c.Get().
		Namespace(ns).
		Resource(resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

func (w *PollingWatcher) start() error {
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		var err error
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var obj, newobj runtime.Object
		gotNotFound := false
		done := false
		for !done {
			obj = newobj
			select {
			case <-ticker.C:
				newobj, err = w.poll()
				if err != nil {
					if api_errors.IsNotFound(err) {
						// This is ok. It's either a delete or not created yet.
						if obj != nil {
							// a delete
							w.result <- watch.Event{
								Type:   watch.Deleted,
								Object: obj,
							}
						}
						gotNotFound = true
						continue
					} else {
						// Send an error and then break
						w.result <- watch.Event{
							Type: watch.Error,
						}
						break
					}
				}
				if gotNotFound {
					// Created
					w.result <- watch.Event{
						Type:   watch.Added,
						Object: newobj,
					}
					gotNotFound = false
					continue
				}
				gotNotFound = false
				// This could still be a create. Are the uids the same?
				newObj, ok1 := newobj.(v1.Object)
				oldObj, ok2 := obj.(v1.Object)
				if ok1 && ok2 && newObj.GetUID() != oldObj.GetUID() {
					// It's a delete and recreate
					w.result <- watch.Event{
						Type:   watch.Deleted,
						Object: obj,
					}
					w.result <- watch.Event{
						Type:   watch.Added,
						Object: newobj,
					}
					continue
				}
				if ok1 && ok2 && newObj.GetResourceVersion() != oldObj.GetResourceVersion() {
					w.result <- watch.Event{
						Type:   watch.Modified,
						Object: newobj,
					}
					continue
				}
			case done = <-w.done:
				break
			}
		}
	}()
	return nil
}

func (w *PollingWatcher) ResultChan() <-chan watch.Event {
	return w.result
}

func (w *PollingWatcher) Stop() {
	w.done <- true
	w.wg.Wait()
	close(w.result)
}
