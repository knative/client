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

	"github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/util"
)

var revisionListHeader = []string{"NAME", "SERVICE", "GENERATION", "AGE", "CONDITIONS", "READY", "REASON"}

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
	} else if output[0] != "No revisions found." {
		t.Errorf("Bad output %s", output[0])
	}
}

func TestRevisionListEmptyByName(t *testing.T) {
	action, _, err := fakeRevisionList([]string{"revision", "list", "name"}, &v1alpha1.RevisionList{})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
}

func TestRevisionListDefaultOutput(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "foo", "1")
	revision2 := createMockRevisionWithParams("bar-abcd", "bar", "1")
	revision3 := createMockRevisionWithParams("foo-wxyz", "foo", "2")
	revision4 := createMockRevisionWithParams("bar-wxyz", "bar", "2")
	// Validate edge case for catching the sorting issue caused by string comparison
	revision5 := createMockRevisionWithParams("foo-wxyz", "foo", "10")
	revision6 := createMockRevisionWithParams("bar-wxyz", "bar", "10")

	RevisionList := &v1alpha1.RevisionList{Items: []v1alpha1.Revision{
		*revision1, *revision2, *revision3, *revision4, *revision5, *revision6}}
	action, output, err := fakeRevisionList([]string{"revision", "list"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output[0], revisionListHeader...))
	assert.Check(t, util.ContainsAll(output[1], "bar-wxyz", "bar", "10"))
	assert.Check(t, util.ContainsAll(output[2], "foo-wxyz", "foo", "10"))
	assert.Check(t, util.ContainsAll(output[3], "bar-wxyz", "bar", "2"))
	assert.Check(t, util.ContainsAll(output[4], "foo-wxyz", "foo", "2"))
	assert.Check(t, util.ContainsAll(output[5], "bar-abcd", "bar", "1"))
	assert.Check(t, util.ContainsAll(output[6], "foo-abcd", "foo", "1"))
}

func TestRevisionListForService(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "svc1", "1")
	revision2 := createMockRevisionWithParams("bar-wxyz", "svc1", "2")
	revision3 := createMockRevisionWithParams("foo-abcd", "svc2", "1")
	revision4 := createMockRevisionWithParams("bar-wxyz", "svc2", "2")
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
	assert.Check(t, util.ContainsAll(output[0], revisionListHeader...))
	assert.Check(t, util.ContainsAll(output[1], "bar-wxyz", "svc1"))
	assert.Check(t, util.ContainsAll(output[2], "foo-abcd", "svc1"))
	action, output, err = fakeRevisionList([]string{"revision", "list", "-s", "svc2"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output[0], revisionListHeader...))
	assert.Check(t, util.ContainsAll(output[1], "bar-wxyz", "svc2"))
	assert.Check(t, util.ContainsAll(output[2], "foo-abcd", "svc2"))
	//test for non existent service
	action, output, err = fakeRevisionList([]string{"revision", "list", "-s", "svc3"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	}
	if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	assert.Assert(t, util.ContainsAll(output[0], "No", "revisions", "svc3"), "no revisions")
}

func TestRevisionListOneOutput(t *testing.T) {
	revision := createMockRevisionWithParams("foo-abcd", "foo", "1")
	RevisionList := &v1alpha1.RevisionList{Items: []v1alpha1.Revision{*revision}}
	action, output, err := fakeRevisionList([]string{"revision", "list", "foo-abcd"}, RevisionList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}

	assert.Assert(t, util.ContainsAll(output[0], revisionListHeader...))
	assert.Assert(t, util.ContainsAll(output[1], "foo", "foo-abcd"))
}

func TestRevisionListOutputWithTwoRevName(t *testing.T) {
	RevisionList := &v1alpha1.RevisionList{Items: []v1alpha1.Revision{}}
	_, _, err := fakeRevisionList([]string{"revision", "list", "foo-abcd", "bar-abcd"}, RevisionList)
	assert.ErrorContains(t, err, "'kn revision list' accepts maximum 1 argument")
}

func createMockRevisionWithParams(name, svcName, generation string) *v1alpha1.Revision {
	revision := &v1alpha1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				serving.ServiceLabelKey:                 svcName,
				serving.ConfigurationGenerationLabelKey: generation,
			},
		},
	}
	return revision
}
