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
	serving "github.com/knative/serving/pkg/apis/serving"
	v1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
)

func fakeRevisionGet(args []string, response *v1alpha1.RevisionList) (action client_testing.Action, output []string, err error) {
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

func TestRevisionGetEmpty(t *testing.T) {
	action, output, err := fakeRevisionGet([]string{"revision", "get"}, &v1alpha1.RevisionList{})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	} else if output[0] != "No resources found." {
		t.Errorf("Bad output %s", output[0])
	}
}

func TestRevisionGetDefaultOutput(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "foo")
	revision2 := createMockRevisionWithParams("bar-wxyz", "bar")
	RevisionList := &v1alpha1.RevisionList{Items: []v1alpha1.Revision{*revision1, *revision2}}
	action, output, err := fakeRevisionGet([]string{"revision", "get"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	testContains(t, output[0], []string{"SERVICE", "NAME", "AGE", "CONDITIONS", "READY", "REASON"}, "column header")
	testContains(t, output[1], []string{"foo", "foo-abcd"}, "value")
	testContains(t, output[2], []string{"bar", "bar-wxyz"}, "value")
}

func testContains(t *testing.T, output string, sub []string, element string) {
	for _, each := range sub {
		if !strings.Contains(output, each) {
			t.Errorf("Missing %s: %s", element, each)
		}
	}
}

func createMockRevisionWithParams(name, svcName string) *v1alpha1.Revision {
	revision := &v1alpha1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels:    map[string]string{serving.ConfigurationLabelKey: svcName},
		},
	}
	return revision
}
