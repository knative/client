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

package apiserver

import (
	"errors"
	"fmt"
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
)

var (
	testApiServerSrcName = "foo"
)

func fakeApiServerSourceCreate(args []string, withExistingService bool, sync bool) (
	action client_testing.Action,
	src *v1alpha1.ApiServerSource,
	output string,
	err error) {
	knParams := &commands.KnParams{}
	cmd, fakeSource, buf := commands.CreateSourcesTestKnCommand(NewApiServerCommand(knParams), knParams)
	fakeSource.AddReactor("create", "apiserversources",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			createAction, ok := a.(client_testing.CreateAction)
			action = createAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", a)
			}
			src, ok = createAction.GetObject().(*v1alpha1.ApiServerSource)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, src, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		output = err.Error()
		return
	}
	output = buf.String()
	return
}

func TestApiServerSourceCreate(t *testing.T) {
	action, created, output, err := fakeApiServerSourceCreate([]string{
		"apiserver", "create", testApiServerSrcName, "--resource", "Event:v1:true", "--service-account", "myaccountname", "--sink", "svc:mysvc"}, true, false)
	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("create", "apiserversources") {
		t.Fatalf("Bad action %v", action)
	}

	//construct a wanted instance
	wanted := &v1alpha1.ApiServerSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testApiServerSrcName,
			Namespace: commands.FakeNamespace,
		},
		Spec: v1alpha1.ApiServerSourceSpec{
			Resources: []v1alpha1.ApiServerResource{{
				APIVersion: "v1",
				Kind:       "Event",
				Controller: true,
			}},
			ServiceAccountName: "myaccountname",
			Mode:               "Ref",
			Sink: &apisv1alpha1.Destination{
				Ref: &v1.ObjectReference{
					Kind:       "Service",
					APIVersion: "serving.knative.dev/v1alpha1",
				},
			},
		},
	}

	//assert equal
	assert.DeepEqual(t, wanted, created)
	assert.Check(t, util.ContainsAll(output, "ApiServerSource", testApiServerSrcName, "created", "namespace", commands.FakeNamespace))
}
