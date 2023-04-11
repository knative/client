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
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/wait"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	client_testing "k8s.io/client-go/testing"
	clienttesting "k8s.io/client-go/testing"
)

func fakeServiceWait(args []string) (action client_testing.Action, name string, output string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
	fakeServing.AddReactor("get", "services",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
			return true, &servingv1.Service{}, nil
		})
	fakeServing.AddWatchReactor("services",
		func(a clienttesting.Action) (bool, watch.Interface, error) {
			watchAction := a.(clienttesting.WatchAction)
			value, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
			if !found {
				return true, nil, errors.New("no field selector on metadata.name found")
			}
			action = watchAction
			name = value
			w := wait.NewFakeWatch(getServiceEvents("test-service"))
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

func TestServiceWaitNoFlags(t *testing.T) {
	sevName := "sev-12345"
	action, name, output, err := fakeServiceWait([]string{"service", "wait", sevName})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("watch", "services") {
		t.Errorf("Bad action %v", action)
	} else if name != sevName {
		t.Errorf("Bad service name returned after wait.")
	}
	assert.Check(t, util.ContainsAll(output, "Service", sevName, "ready", "namespace", commands.FakeNamespace))
}

func TestServiceWaitWithFlags(t *testing.T) {
	sevName := "sev-08567"
	action, name, output, err := fakeServiceWait([]string{"service", "wait", sevName, "--wait-timeout", "2", "--wait-window", "1"})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("watch", "services") {
		t.Errorf("Bad action %v", action)
	} else if name != sevName {
		t.Errorf("Bad service name returned after wait.")
	}
	assert.Check(t, util.ContainsAll(output, "Service", sevName, "ready", "namespace", commands.FakeNamespace))
}
