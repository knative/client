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

package apiserver

import (
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

func fakeServiceDelete(args []string) (action client_testing.Action, name string, output string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeSource, buf := commands.CreateSourcesTestKnCommand(NewApiServerCommand(knParams), knParams)
	fakeSource.AddReactor("delete", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			deleteAction, _ := a.(client_testing.DeleteAction)
			action = deleteAction
			name = deleteAction.GetName()
			return true, nil, nil
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
	srcName := "src-12345"
	action, name, output, err := fakeServiceDelete([]string{"apiserver", "delete", srcName})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("delete", "apiserversources") {
		t.Errorf("Bad action %v", action)
	} else if name != srcName {
		t.Errorf("Bad service name returned after delete.")
	}
	assert.Check(t, util.ContainsAll(output, "ApiServerSource", srcName, "deleted", "namespace", commands.FakeNamespace))
}
