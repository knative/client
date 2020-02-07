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
	"testing"
	"time"

	"gotest.tools/assert"
	api_errors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

type fakePollInterval struct {
	c chan time.Time
}

func (f *fakePollInterval) PollChan() <-chan time.Time {
	return f.c
}

func (f *fakePollInterval) Stop() {}

func newFakePollInterval(n int) PollInterval {
	c := make(chan time.Time, n)
	t := time.Now()
	for i := 0; i < n; i++ {
		c <- t.Add(time.Duration(i) * time.Second)
	}
	return &fakePollInterval{c}
}

func newWatcherForTest(pollResults []runtime.Object) watch.Interface {
	i := 0
	poll := func() (runtime.Object, error) {
		defer func() { i += 1 }()
		if pollResults[i] == nil {
			// 404
			return nil, api_errors.NewNotFound(schema.GroupResource{"thing", "stuff"}, "eggs")
		}
		return pollResults[i], nil
	}
	ret := &pollingWatcher{nil, "", "", "", time.Minute, make(chan bool), make(chan watch.Event), &sync.WaitGroup{},
		newFakePollInterval(len(pollResults)), poll}
	ret.start()
	return ret
}

var a, aa, b, bb, c, cc, z, zz runtime.Object

func init() {
	a = &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", ResourceVersion: "a", UID: "one"}}
	aa = a.DeepCopyObject()
	b = &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", ResourceVersion: "b", UID: "one"}}
	bb = b.DeepCopyObject()
	c = &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", ResourceVersion: "c", UID: "one"}}
	cc = c.DeepCopyObject()
	z = &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", ResourceVersion: "z", UID: "two"}}
	zz = z.DeepCopyObject()
}

type testCase struct {
	pollResults  []runtime.Object
	watchResults []watch.Event
}

func TestPollWatcher(t *testing.T) {
	cases := []testCase{
		// Doesn't exist for a while, then does for a while.
		{[]runtime.Object{nil, nil, a, aa, nil}, []watch.Event{{watch.Added, a}, {watch.Deleted, a}}},
		// Changes.
		{[]runtime.Object{a, b}, []watch.Event{{watch.Added, a}, {watch.Modified, b}}},
		// Changes but stays the same a couple times too.
		{[]runtime.Object{a, aa, b, bb, c, cc, nil},
			[]watch.Event{{watch.Added, a}, {watch.Modified, b}, {watch.Modified, c}, {watch.Deleted, c}}},
		// Deleted and recreated between polls.
		{[]runtime.Object{a, z}, []watch.Event{{watch.Added, a}, {watch.Deleted, a}, {watch.Added, z}}},
	}
	for _, c := range cases {
		w := newWatcherForTest(c.pollResults)
		for _, expected := range c.watchResults {
			actual := <-w.ResultChan()
			assert.Equal(t, actual.Type, expected.Type)
			if actual.Type == watch.Added || actual.Type == watch.Modified || actual.Type == watch.Deleted {
				fmt.Printf("expected, %v, actual %v\n", expected, actual)
				assert.Equal(t, actual.Object.(metav1.Object).GetResourceVersion(), expected.Object.(metav1.Object).GetResourceVersion())
				assert.Equal(t, actual.Object.(metav1.Object).GetUID(), expected.Object.(metav1.Object).GetUID())
			}
		}
		w.Stop()
	}
}
