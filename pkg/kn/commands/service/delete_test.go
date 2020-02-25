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

package service

import (
	"testing"

	"github.com/pkg/errors"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	client_testing "k8s.io/client-go/testing"
	clienttesting "k8s.io/client-go/testing"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"
)

func fakeServiceDelete(args []string) (action client_testing.Action, name string, output string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
	fakeServing.AddReactor("delete", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			deleteAction, _ := a.(client_testing.DeleteAction)
			action = deleteAction
			name = deleteAction.GetName()
			return true, nil, nil
		})
	fakeServing.AddWatchReactor("services",
		func(a client_testing.Action) (bool, watch.Interface, error) {
			watchAction := a.(clienttesting.WatchAction)
			_, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
			if !found {
				return true, nil, errors.New("no field selector on metadata.name found")
			}
			w := wait.NewFakeWatch(getServiceDeleteEvents("test-service"))
			w.Start()
			return true, w, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = buf.String()
	return
}

func TestServiceDelete(t *testing.T) {
	sevName := "sev-12345"
	action, name, output, err := fakeServiceDelete([]string{"service", "delete", sevName})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("delete", "services") {
		t.Errorf("Bad action %v", action)
	} else if name != sevName {
		t.Errorf("Bad service name returned after delete.")
	}
	assert.Check(t, util.ContainsAll(output, "Service", sevName, "deleted", "namespace", commands.FakeNamespace))
}

func TestMultipleServiceDelete(t *testing.T) {
	sevName1 := "sev-12345"
	sevName2 := "sev-67890"
	sevName3 := "sev-abcde"
	action, _, output, err := fakeServiceDelete([]string{"service", "delete", sevName1, sevName2, sevName3})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("delete", "services") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output, "Service", sevName1, "deleted", "namespace", commands.FakeNamespace))
	assert.Check(t, util.ContainsAll(output, "Service", sevName2, "deleted", "namespace", commands.FakeNamespace))
	assert.Check(t, util.ContainsAll(output, "Service", sevName3, "deleted", "namespace", commands.FakeNamespace))
}

func getServiceDeleteEvents(name string) []watch.Event {
	return []watch.Event{
		{watch.Added, wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionUnknown, "", "msg1")},
		{watch.Modified, wait.CreateTestServiceWithConditions(name, corev1.ConditionUnknown, corev1.ConditionTrue, "", "msg2")},
		{watch.Deleted, wait.CreateTestServiceWithConditions(name, corev1.ConditionTrue, corev1.ConditionTrue, "", "")},
	}
}
