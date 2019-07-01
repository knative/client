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

func fakeRevisionList(args []string, response *v1alpha1.RevisionList) (action client_testing.Action, output []string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewRevisionCommand(knParams), knParams)
	fakeServing.AddReactor("list", "*",
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

func TestRevisionListEmpty(t *testing.T) {
	action, output, err := fakeRevisionList([]string{"revision", "list"}, &v1alpha1.RevisionList{})
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

func TestRevisionListDefaultOutput(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "foo")
	revision2 := createMockRevisionWithParams("bar-wxyz", "bar")
	RevisionList := &v1alpha1.RevisionList{Items: []v1alpha1.Revision{*revision1, *revision2}}
	action, output, err := fakeRevisionList([]string{"revision", "list"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	testContains(t, output[0], []string{"NAME", "SERVICE", "AGE", "CONDITIONS", "READY", "REASON"}, "column header")
	testContains(t, output[1], []string{"foo-abcd", "foo"}, "value")
	testContains(t, output[2], []string{"bar-wxyz", "bar"}, "value")
}

func TestRevisionListForService(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "svc1")
	revision2 := createMockRevisionWithParams("bar-wxyz", "svc1")
	revision3 := createMockRevisionWithParams("foo-abcd", "svc2")
	revision4 := createMockRevisionWithParams("bar-wxyz", "svc2")
	RevisionList := &v1alpha1.RevisionList{Items: []v1alpha1.Revision{*revision1, *revision2, *revision3, *revision4}}
	action, output, err := fakeRevisionList([]string{"revision", "list", "-s", "svc1"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	testContains(t, output[0], []string{"NAME", "SERVICE", "AGE", "CONDITIONS", "READY", "REASON"}, "column header")
	testContains(t, output[1], []string{"foo-abcd", "svc1"}, "value")
	testContains(t, output[2], []string{"bar-wxyz", "svc1"}, "value")
	action, output, err = fakeRevisionList([]string{"revision", "list", "-s", "svc2"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	testContains(t, output[0], []string{"NAME", "SERVICE", "AGE", "CONDITIONS", "READY", "REASON"}, "column header")
	testContains(t, output[1], []string{"foo-abcd", "svc2"}, "value")
	testContains(t, output[2], []string{"bar-wxyz", "svc2"}, "value")
	//test for non existent service
	action, output, err = fakeRevisionList([]string{"revision", "list", "-s", "svc3"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	} else if !strings.Contains(output[0], "No resources found.") {
		t.Errorf("Bad output %s", output[0])
	}
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
			Labels:    map[string]string{serving.ServiceLabelKey: svcName},
		},
	}
	return revision
}
