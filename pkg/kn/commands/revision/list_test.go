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

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

var revisionListHeader = []string{"NAME", "SERVICE", "TRAFFIC", "TAGS", "GENERATION", "AGE", "CONDITIONS", "READY", "REASON"}

func fakeRevisionList(args []string, response *servingv1.RevisionList) (action clienttesting.Action, output []string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewRevisionCommand(knParams), knParams)
	fakeServing.AddReactor("list", "*",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
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
	action, output, err := fakeRevisionList([]string{"revision", "list"}, &servingv1.RevisionList{})
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	} else if output[0] != "No revisions found." {
		t.Errorf("Bad output %s", output[0])
	}
}

func TestRevisionListEmptyByName(t *testing.T) {
	action, _, err := fakeRevisionList([]string{"revision", "list", "name"}, &servingv1.RevisionList{})
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
}

func TestRevisionListDefaultOutput(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "foo", "1", "100", "")
	revision2 := createMockRevisionWithParams("bar-abcd", "bar", "1", "100", "")
	revision3 := createMockRevisionWithParams("foo-wxyz", "foo", "2", "50", "tag10")
	revision4 := createMockRevisionWithParams("bar-wxyz", "bar", "2", "0", "tag2")
	// Validate edge case for catching the sorting issue caused by string comparison
	revision5 := createMockRevisionWithParams("foo-wxyz", "foo", "10", "tag1", "tagx")
	revision6 := createMockRevisionWithParams("bar-wxyz", "bar", "10", "50", "")

	RevisionList := &servingv1.RevisionList{Items: []servingv1.Revision{
		*revision1, *revision2, *revision3, *revision4, *revision5, *revision6}}
	action, output, err := fakeRevisionList([]string{"revision", "list"}, RevisionList)
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	assert.Check(t, util.ContainsAll(output[0], revisionListHeader...))
	var expectedOutput [][]string = [][]string{
		{"bar-wxyz", "bar", "10"},
		{"bar-wxyz", "bar", "2"},
		{"bar-abcd", "bar", "1"},
		{"foo-wxyz", "foo", "10"},
		{"foo-wxyz", "foo", "2"},
		{"foo-abcd", "foo", "1"},
	}
	for i, content := range expectedOutput {
		assert.Check(t, util.ContainsAll(output[i+1], content...))
	}
}

func TestRevisionListDefaultOutputNoHeaders(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "foo", "2", "100", "")
	revision2 := createMockRevisionWithParams("bar-wxyz", "bar", "1", "100", "")
	RevisionList := &servingv1.RevisionList{Items: []servingv1.Revision{*revision1, *revision2}}
	action, output, err := fakeRevisionList([]string{"revision", "list", "--no-headers"}, RevisionList)
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}

	assert.Check(t, util.ContainsNone(output[0], "NAME", "URL", "GENERATION", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(output[0], "bar-wxyz", "bar", "1"))
	assert.Check(t, util.ContainsAll(output[1], "foo-abcd", "foo", "2"))

}

func TestRevisionListForService(t *testing.T) {
	revision1 := createMockRevisionWithParams("foo-abcd", "svc1", "1", "50", "")
	revision2 := createMockRevisionWithParams("bar-wxyz", "svc1", "2", "50", "")
	revision3 := createMockRevisionWithParams("foo-abcd", "svc2", "1", "0", "")
	revision4 := createMockRevisionWithParams("bar-wxyz", "svc2", "2", "100", "")
	RevisionList := &servingv1.RevisionList{Items: []servingv1.Revision{*revision1, *revision2, *revision3, *revision4}}
	action, output, err := fakeRevisionList([]string{"revision", "list", "-s", "svc1"}, RevisionList)
	assert.NilError(t, err)
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
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	}
	if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}
	assert.Assert(t, util.ContainsAll(output[0], "No", "revisions", "found"), "no revisions")
}

func TestRevisionListOneOutput(t *testing.T) {
	revision := createMockRevisionWithParams("foo-abcd", "foo", "1", "100", "")
	RevisionList := &servingv1.RevisionList{Items: []servingv1.Revision{*revision}}
	action, output, err := fakeRevisionList([]string{"revision", "list", "foo-abcd"}, RevisionList)
	assert.NilError(t, err)
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "revisions") {
		t.Errorf("Bad action %v", action)
	}

	assert.Assert(t, util.ContainsAll(output[0], revisionListHeader...))
	assert.Assert(t, util.ContainsAll(output[1], "foo", "foo-abcd"))
}

func TestRevisionListOutputWithTwoRevName(t *testing.T) {
	RevisionList := &servingv1.RevisionList{Items: []servingv1.Revision{}}
	_, _, err := fakeRevisionList([]string{"revision", "list", "foo-abcd", "bar-abcd"}, RevisionList)
	assert.ErrorContains(t, err, "'kn revision list' accepts maximum 1 argument")
}

func createMockRevisionWithParams(name, svcName, generation, traffic, tags string) *servingv1.Revision {
	revision := &servingv1.Revision{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Revision",
			APIVersion: "serving.knative.dev/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				serving.ServiceLabelKey:                 svcName,
				serving.ConfigurationGenerationLabelKey: generation,
			},
			Annotations: map[string]string{
				"client.knative.dev/traffic": traffic,
				"client.knative.dev/tags":    tags,
			},
		},
	}
	return revision
}
