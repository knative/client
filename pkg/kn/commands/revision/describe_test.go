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

package revision

import (
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
	"sigs.k8s.io/yaml"
)

func fakeRevision(args []string, response *v1alpha1.Revision) (action client_testing.Action, output string, err error) {
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
	output = buf.String()
	return
}

func TestDescribeRevisionWithNoName(t *testing.T) {
	_, _, err := fakeRevision([]string{"revision", "describe"}, &v1alpha1.Revision{})
	expectedError := "requires the revision name."
	if err == nil || err.Error() != expectedError {
		t.Fatal("expect to fail with missing revision name")
	}
}

func TestDescribeRevisionYaml(t *testing.T) {
	expectedRevision := v1alpha1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: v1alpha1.RevisionSpec{
			DeprecatedContainer: &corev1.Container{
				Name:  "some-container",
				Image: "knative/test:latest",
			},
		},
		Status: v1alpha1.RevisionStatus{
			ServiceName: "foo-service",
		},
	}

	action, data, err := fakeRevision([]string{"revision", "describe", "test-rev", "-o", "yaml"}, &expectedRevision)
	if err != nil {
		t.Fatal(err)
	}

	if action == nil {
		t.Fatal("No action")
	} else if !action.Matches("get", "revisions") {
		t.Fatalf("Bad action %v", action)
	}

	jsonData, err := yaml.YAMLToJSON([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	var returnedRevision v1alpha1.Revision
	err = json.Unmarshal(jsonData, &returnedRevision)
	if err != nil {
		t.Fatal(err)
	}

	if !equality.Semantic.DeepEqual(expectedRevision, returnedRevision) {
		t.Fatal("mismatched objects")
	}
}
