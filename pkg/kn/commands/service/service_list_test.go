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

package service

import (
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	v1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
)

func fakeServiceList(args []string, response *v1alpha1.ServiceList) (action client_testing.Action, output []string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
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
	action, output, err := fakeServiceList([]string{"service", "list"}, &v1alpha1.ServiceList{})
	if err != nil {
		t.Error(err)
		return
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "services") {
		t.Errorf("Bad action %v", action)
	} else if output[0] != "No resources found." {
		t.Errorf("Bad output %s", output[0])
	}
}

func TestServiceListDefaultOutput(t *testing.T) {
	service1 := createMockServiceWithParams("foo", "foo.default.example.com", 1)
	service2 := createMockServiceWithParams("bar", "bar.default.example.com", 2)
	serviceList := &v1alpha1.ServiceList{Items: []v1alpha1.Service{*service1, *service2}}
	action, output, err := fakeServiceList([]string{"service", "list"}, serviceList)
	if err != nil {
		t.Fatal(err)
	}
	if action == nil {
		t.Errorf("No action")
	} else if !action.Matches("list", "services") {
		t.Errorf("Bad action %v", action)
	}
	testContains(t, output[0], []string{"NAME", "DOMAIN", "GENERATION", "AGE", "CONDITIONS", "READY", "REASON"}, "column header")
	testContains(t, output[1], []string{"foo", "foo.default.example.com", "1"}, "value")
	testContains(t, output[2], []string{"bar", "bar.default.example.com", "2"}, "value")
}

func testContains(t *testing.T, output string, sub []string, element string) {
	for _, each := range sub {
		if !strings.Contains(output, each) {
			t.Errorf("Missing %s: %s", element, each)
		}
	}
}

func createMockServiceWithParams(name, domain string, generation int64) *v1alpha1.Service {
	service := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{
			DeprecatedRunLatest: &v1alpha1.RunLatestType{},
		},
		Status: v1alpha1.ServiceStatus{
			Status: duckv1beta1.Status{
				ObservedGeneration: generation},
			RouteStatusFields: v1alpha1.RouteStatusFields{
				DeprecatedDomain: domain,
			},
		},
	}
	return service
}
