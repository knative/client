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

func fakeRevisionGet(args []string, revisions *v1alpha1.RevisionList, routes *v1alpha1.RouteList) (
	action client_testing.Action, output []string, err error) {
	buf := new(bytes.Buffer)
	fakeServing := &fake.FakeServingV1alpha1{&client_testing.Fake{}}
	cmd := NewKnCommand(KnParams{
		Output:         buf,
		ServingFactory: func() (serving.ServingV1alpha1Interface, error) { return fakeServing, nil },
	})
	fakeServing.AddReactor("*", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			action = a
			if action.Matches("list", "routes") {
				return true, routes, nil
			}
			return true, revisions, nil
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
	action, output, err := fakeRevisionGet(
		[]string{"revision", "get"},
		&v1alpha1.RevisionList{},
		&v1alpha1.RouteList{})

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
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	}
}

var revisionType = metav1.TypeMeta{
	Kind:       "revision",
	APIVersion: "serving.knative.dev/v1alpha1",
}
var routeType = metav1.TypeMeta{
	Kind:       "route",
	APIVersion: "serving.knative.dev/v1alpha1",
}

func TestRevisionGetDefaultOutput(t *testing.T) {
	fooLabel := make(map[string]string)
	barLabel := make(map[string]string)
	fooLabel["serving.knative.dev/service"] = "f1"
	barLabel["serving.knative.dev/service"] = "b1"

	// sample RevisionList
	rev_list := &v1alpha1.RevisionList{
		Items: []v1alpha1.Revision{
			v1alpha1.Revision{
				TypeMeta: revisionType,
				ObjectMeta: metav1.ObjectMeta{
					Name:   "foo",
					Labels: fooLabel,
				},
			},
			v1alpha1.Revision{
				TypeMeta: revisionType,
				ObjectMeta: metav1.ObjectMeta{
					Name:   "bar",
					Labels: barLabel,
				},
			},
		},
	}
	// sample RouteList
	route_list := &v1alpha1.RouteList{
		Items: []v1alpha1.Route{
			v1alpha1.Route{
				TypeMeta: routeType,
				Status: v1alpha1.RouteStatus{
					Domain: "foo.default.example.com",
					Traffic: []v1alpha1.TrafficTarget{
						v1alpha1.TrafficTarget{
							RevisionName: "foo",
							Percent:      100,
						},
					},
				},
			},
			v1alpha1.Route{
				TypeMeta: routeType,
				Status: v1alpha1.RouteStatus{
					Domain: "bar.default.example.com",
					Traffic: []v1alpha1.TrafficTarget{
						v1alpha1.TrafficTarget{
							RevisionName: "bar",
							Percent:      100,
						},
					},
				},
			},
		},
	}

	action, output, err := fakeRevisionGet(
		[]string{"revision", "get"},
		rev_list,
		route_list)
	if err != nil {
		t.Fatal(err)
	}
	// each line's tab/spaces are replaced by comma
	expected := []string{"NAME,SERVICE,AGE,TRAFFIC",
		"foo,f1,,100% -> foo.default.example.com",
		"bar,b1,,100% -> bar.default.example.com"}
	expected_lines := strings.Split(tabbedOutput(expected), "\n")

	for i, s := range output {
		if s != expected_lines[i] {
			t.Errorf("Bad output line %v expected %v", s, expected_lines[i])
		}
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "routes") {
		t.Errorf("Bad action %v", action)
	}
}
