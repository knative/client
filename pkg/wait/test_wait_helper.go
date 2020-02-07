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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"knative.dev/pkg/apis"
	duck "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// Helper for testing watch functionality
type FakeWatch struct {
	eventChan chan watch.Event
	events    []watch.Event

	// Record how often stop was called
	StopCalled int
}

// Create a new fake watch with the given events which will be send when
// on start
func NewFakeWatch(events []watch.Event) *FakeWatch {
	return &FakeWatch{
		eventChan: make(chan watch.Event),
		events:    events,
	}
}

// Stop the watch challend
func (f *FakeWatch) Stop() {
	f.StopCalled++
}

// Start and fire events
func (f *FakeWatch) Start() {
	go f.fireEvents()
}

// Channel for getting the events
func (f *FakeWatch) ResultChan() <-chan watch.Event {
	return f.eventChan
}

func (f *FakeWatch) fireEvents() {
	for _, ev := range f.events {
		f.eventChan <- ev
	}
}

// Create a service skeleton with a given ConditionReady status and all other statuses set to otherReadyStatus. Optionally a single generation can be added.
func CreateTestServiceWithConditions(name string, readyStatus corev1.ConditionStatus, otherReadyStatus corev1.ConditionStatus, reason string, message string, generations ...int64) runtime.Object {
	service := servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if len(generations) == 2 {
		service.Generation = generations[0]
		service.Status.ObservedGeneration = generations[1]
	} else {
		service.Generation = 1
		service.Status.ObservedGeneration = 1
	}
	service.Status.Conditions = duck.Conditions([]apis.Condition{
		{Type: "RoutesReady", Status: otherReadyStatus},
		{Type: apis.ConditionReady, Status: readyStatus, Reason: reason, Message: message},
		{Type: "ConfigurationsReady", Status: otherReadyStatus},
	})
	return &service
}
