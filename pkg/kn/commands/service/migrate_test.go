//// Copyright © 2019 The Knative Authors
////
//// Licensed under the Apache License, Version 2.0 (the "License");
//// you may not use this file except in compliance with the License.
//// You may obtain a copy of the License at
////
////     http://www.apache.org/licenses/LICENSE-2.0
////
//// Unless required by applicable law or agreed to in writing, software
//// distributed under the License is distributed on an "AS IS" BASIS,
//// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//// See the License for the specific language governing permissions and
//// limitations under the License.
//
package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	servinglib "knative.dev/client/pkg/serving"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/wait"

	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/serving/pkg/apis/serving/v1alpha1"
)

func fakeMigrate(args []string, withExistingService bool, sync bool) (
	action client_testing.Action,
	service *v1alpha1.Service,
	output string,
	err error) {
	knParams := &commands.KnParams{}
	nrGetCalled := 0
	cmd, fakeServing, buf := commands.CreateTestKnCommand(NewServiceCommand(knParams), knParams)
	fakeServing.AddReactor("get", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			nrGetCalled++
			if withExistingService || (sync && nrGetCalled > 1) {
				return true, &v1alpha1.Service{}, nil
			}
			return true, nil, api_errors.NewNotFound(schema.GroupResource{}, "")
		})
	fakeServing.AddReactor("create", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			createAction, ok := a.(client_testing.CreateAction)
			action = createAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", a)
			}
			service, ok = createAction.GetObject().(*v1alpha1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, service, nil
		})
	fakeServing.AddReactor("update", "services",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			updateAction, ok := a.(client_testing.UpdateAction)
			action = updateAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", a)
			}
			service, ok = updateAction.GetObject().(*v1alpha1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, service, nil
		})
	if sync {
		fakeServing.AddWatchReactor("services",
			func(a client_testing.Action) (bool, watch.Interface, error) {
				watchAction := a.(client_testing.WatchAction)
				_, found := watchAction.GetWatchRestrictions().Fields.RequiresExactMatch("metadata.name")
				if !found {
					return true, nil, errors.New("no field selector on metadata.name found")
				}
				w := wait.NewFakeWatch(getServiceEvents("test-service"))
				w.Start()
				return true, w, nil
			})
		fakeServing.AddReactor("get", "services",
			func(a client_testing.Action) (bool, runtime.Object, error) {
				return true, &v1alpha1.Service{}, nil
			})
	}
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		output = err.Error()
		return
	}
	output = buf.String()
	return
}

func TestMigrateWithoutNamespaceParameter(t *testing.T) {
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz", "--async"}, false, false)
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}
	template, err := servinglib.RevisionTemplateOfService(created)
	if err != nil {
		t.Fatal(err)
	} else if template.Spec.Containers[0].Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", template.Spec.Containers[0].Image)
	} else if !strings.Contains(output, "foo") || !strings.Contains(output, "created") ||
		!strings.Contains(output, commands.FakeNamespace) {
		t.Fatalf("wrong stdout message: %v", output)
	}

	_, _, output, err = fakeMigrate([]string{
		"service", "migrate"}, false, false)
	if err != nil {
		if !strings.Contains(output, "--source-namespace") {
			t.Fatalf("wrong stdout message: %v", output)
		}
	}
}

func TestMigrateWithoutKubeconfigParameter(t *testing.T) {
	action, created, output, err := fakeServiceCreate([]string{
		"service", "create", "foo", "--image", "gcr.io/foo/bar:baz"}, false, true)
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "services") {
		t.Fatalf("Bad action %v", action)
	}
	template, err := servinglib.RevisionTemplateOfService(created)
	if err != nil {
		t.Fatal(err)
	}
	if template.Spec.Containers[0].Image != "gcr.io/foo/bar:baz" {
		t.Fatalf("wrong image set: %v", template.Spec.Containers[0].Image)
	}
	if !strings.Contains(output, "foo") || !strings.Contains(output, "created") ||
		!strings.Contains(output, commands.FakeNamespace) {
		t.Fatalf("wrong stdout message: %v", output)
	}
	if !strings.Contains(output, "OK") || !strings.Contains(output, "Waiting") {
		t.Fatalf("not running in sync mode")
	}
	_, _, output, err = fakeMigrate([]string{
		"service", "migrate", "--source-namespace", "default", "--destination-namespace", "default"}, false, false)
	if err != nil {
		if !strings.Contains(output, "--source-kubeconfig") {
			t.Fatalf("wrong stdout message: %v", output)
		}
	}
}
