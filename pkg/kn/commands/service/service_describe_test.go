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

package service

import (
	"regexp"
	"strings"
	"testing"

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"

	"github.com/knative/client/pkg/kn/commands"
)

func fakeServiceDescribe(args []string, response *v1alpha1.Service) (action clienttesting.Action, output string, err error) {
	knParams := &commands.KnParams{}
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
	fakeServing.AddReactor("*", "*",
		func(a clienttesting.Action) (bool, runtime.Object, error) {
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
	if err == nil ||
		!strings.Contains(err.Error(), "no") ||
		!strings.Contains(err.Error(), "service") ||
		!strings.Contains(err.Error(), "provided") {
		t.Fatalf("expect to fail with missing service name (got: %v)", err)
	}
}

func TestServiceDescribeDefaultOutput(t *testing.T) {
	expectedService := v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Status: v1alpha1.ServiceStatus{
			RouteStatusFields: v1alpha1.RouteStatusFields{
				DeprecatedDomain: "foo.default.example.com",
				Address:          &duckv1alpha1.Addressable{Hostname: "foo.default.svc.cluster.local"},
			},
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

	assertMatches(t, output, "Name:\\s+foo")
	assertMatches(t, output, "Namespace:\\s+default")
	assertMatches(t, output, "Address:\\s+foo.default.svc.cluster.local")
	assertMatches(t, output, "URL:\\s+foo.default.example.com")
	assertMatches(t, output, "Age:")
}

func assertMatches(t *testing.T, value, expr string) {
	ok, err := regexp.MatchString(expr, value)
	if err != nil {
		t.Fatalf("invalid pattern %q. %v", expr, err)
	}
	if !ok {
		t.Errorf("got %s which does not match %s\n", value, expr)
	}
}
