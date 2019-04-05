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
	"fmt"
	"strings"
	"testing"

	printers "github.com/knative/client/pkg/util/printers"
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
	expected := []string{"No resources found.", ""}
	for i, s := range output {
		if s != expected[i] {
			t.Errorf("%d Bad output line %v", i, s)
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

// tabbedOutput takes a list of strings and returns the tabwriter formatted output
func tabbedOutput(s []string) string {
	buf := new(bytes.Buffer)
	printer := printers.GetNewTabWriter(buf)

	for _, line := range s {
		line_items := strings.Split(line, ",")
		fmt.Fprintf(printer, "%s\n", strings.Join(line_items, "\t"))
	}
	printer.Flush()
	return buf.String()

}
func TestListDefaultOutput(t *testing.T) {
	action, output, err := fakeList([]string{"service", "list"}, &v1alpha1.ServiceList{
		Items: []v1alpha1.Service{
			v1alpha1.Service{
				TypeMeta: serviceType,
				ObjectMeta: metav1.ObjectMeta{
					Name: "foo",
				},
				Status: v1alpha1.ServiceStatus{
					Domain:                    "foo.default.example.com",
					LatestCreatedRevisionName: "foo-abcde",
					LatestReadyRevisionName:   "foo-abcde",
				},
			},
			v1alpha1.Service{
				TypeMeta: serviceType,
				ObjectMeta: metav1.ObjectMeta{
					Name: "bar",
				},
				Status: v1alpha1.ServiceStatus{
					Domain:                    "bar.default.example.com",
					LatestCreatedRevisionName: "bar-abcde",
					LatestReadyRevisionName:   "bar-abcde",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	// each line's tab/spaces are replaced by comma
	expected := []string{"NAME,DOMAIN,LATESTCREATED,LATESTREADY,AGE",
		"foo,foo.default.example.com,foo-abcde,foo-abcde,",
		"bar,bar.default.example.com,bar-abcde,bar-abcde,"}
	expected_lines := strings.Split(tabbedOutput(expected), "\n")

	for i, s := range output {
		if s != expected_lines[i] {
			t.Errorf("Bad output line %v expected %v", s, expected_lines[i])
		}
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "services") {
		t.Errorf("Bad action %v", action)
	}
}
