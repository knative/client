// Copyright © 2018 The Knative Authors
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

package revision

import (
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	v1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
)

func fakeRevision(args []string, response *v1alpha1.ServiceList) (action client_testing.Action, output []string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewRevisionCommand(knParams), knParams)
	fakeServing.AddReactor("*", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			action = a
			return true, response, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = strings.Split(buf.String(), "\n")
	return
}

func TestUnknownSubcommand(t *testing.T) {
	_, _, err := fakeRevision([]string{"revision", "unknown"}, &v1alpha1.ServiceList{})
	if err == nil {
		t.Error(err)
		return
	}

	if err.Error() != "unknown command \"unknown\" for \"kn revision\"" {
		t.Errorf("Bad error message '%s'", err.Error())
	}
}

func TestEmptySubcommand(t *testing.T) {
	_, _, err := fakeRevision([]string{"revision"}, &v1alpha1.ServiceList{})
	if err == nil {
		t.Error(err)
		return
	}

	if err.Error() != "please provide a valid sub-command for \"kn revision\"" {
		t.Errorf("Bad error message '%s'", err.Error())
	}
}
