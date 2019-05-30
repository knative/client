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

package commands

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"sigs.k8s.io/yaml"
)

func fakeServiceDescribe(args []string, response *v1alpha1.Service) (action client_testing.Action, output string, err error) {
	buf := new(bytes.Buffer)
	fakeServing := &fake.FakeServingV1alpha1{&client_testing.Fake{}}
	cmd := NewKnCommand(KnParams{
		Output:         buf,
		ServingFactory: func() (serving.ServingV1alpha1Interface, error) { return fakeServing, nil },
	})
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

func TestEmptyServiceDescribe(t *testing.T) {
	_, _, err := fakeServiceDescribe([]string{"service", "describe"}, &v1alpha1.Service{})
	expectedError := "requires the service name."
	if err == nil || err.Error() != expectedError {
		t.Fatal("expect to fail with missing service name")
	}
}

func TestServiceDescribeDefaultOutput(t *testing.T) {
	expectedService := v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
	}
	action, output, err := fakeServiceDescribe([]string{"service", "describe", "test-foo"}, &expectedService)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Fatal("No action")
	} else if !action.Matches("get", "services") {
		t.Fatalf("Bad action %v", action)
	}

	jsonData, err := yaml.YAMLToJSON([]byte(output))
	if err != nil {
		t.Fatal(err)
	}
	var returnedService v1alpha1.Service
	err = json.Unmarshal(jsonData, &returnedService)
	if err != nil {
		t.Fatal(err)
	}

	if !equality.Semantic.DeepEqual(expectedService, returnedService) {
		t.Fatal("mismatched objects")
	}
}
