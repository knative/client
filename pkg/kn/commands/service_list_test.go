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

package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
)

func fakeList(args []string, response *v1alpha1.ServiceList) (action client_testing.Action, output []string, err error) {
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
	output = strings.Split(buf.String(), "\n")
	return
}

func TestListEmpty(t *testing.T) {
	action, output, err := fakeList([]string{"service", "list"}, &v1alpha1.ServiceList{})
	if err != nil {
		t.Error(err)
		return
	}
	for _, s := range output {
		if s != "" {
			t.Errorf("Bad output line %v", s)
		}
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "services") {
		t.Errorf("Bad action %v", action)
	}
}

var serviceType = metav1.TypeMeta{
	Kind:       "service",
	APIVersion: "serving.knative.dev/v1alpha1",
}

func TestListDefaultOutput(t *testing.T) {
	action, output, err := fakeList([]string{"service", "list"}, &v1alpha1.ServiceList{
		Items: []v1alpha1.Service{
			v1alpha1.Service{
				TypeMeta: serviceType,
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
			},
			v1alpha1.Service{
				TypeMeta: serviceType,
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"foo", "bar", ""}
	for i, s := range output {
		if s != expected[i] {
			t.Errorf("Bad output line %v expected %v", s, expected[i])
		}
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "services") {
		t.Errorf("Bad action %v", action)
	}
}
