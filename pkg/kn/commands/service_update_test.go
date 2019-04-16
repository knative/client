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
	"errors"
	"fmt"
	"testing"

	servinglib "github.com/knative/client/pkg/serving"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	client_testing "k8s.io/client-go/testing"
)

func fakeServiceUpdate(original *v1alpha1.Service, args []string) (
	action client_testing.Action,
	updated *v1alpha1.Service,
	output string,
	err error) {

	buf := new(bytes.Buffer)
	fakeServing := &fake.FakeServingV1alpha1{&client_testing.Fake{}}
	cmd := NewKnCommand(KnParams{
		Output:         buf,
		ServingFactory: func() (serving.ServingV1alpha1Interface, error) { return fakeServing, nil },
	})
	fakeServing.AddReactor("update", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			updateAction, ok := a.(client_testing.UpdateAction)
			action = updateAction
			if !ok {
				return true, nil, fmt.Errorf("wrong kind of action %v", action)
			}
			updated, ok = updateAction.GetObject().(*v1alpha1.Service)
			if !ok {
				return true, nil, errors.New("was passed the wrong object")
			}
			return true, updated, nil
		})
	fakeServing.AddReactor("get", "*",
		func(a client_testing.Action) (bool, runtime.Object, error) {
			return true, original, nil
		})
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}
	output = buf.String()
	return
}

func TestServiceUpdateImage(t *testing.T) {

	orig := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: v1alpha1.ServiceSpec{
			RunLatest: &v1alpha1.RunLatestType{},
		},
	}

	config, err := servinglib.GetConfiguration(orig)

	if err != nil {
		t.Fatal(err)
	}

	servinglib.UpdateImage(config, "gcr.io/foo/bar:baz")

	action, updated, _, err := fakeServiceUpdate(orig, []string{
		"service", "update", "foo", "--image", "gcr.io/foo/quux:xyzzy"})

	if err != nil {
		t.Fatal(err)
	} else if !action.Matches("update", "services") {
		t.Fatalf("Bad action %v", action)
	}
	conf, err := servinglib.GetConfiguration(updated)
	if err != nil {
		t.Fatal(err)
	} else if conf.RevisionTemplate.Spec.Container.Image != "gcr.io/foo/quux:xyzzy" {
		t.Fatalf("wrong image set: %v", conf.RevisionTemplate.Spec.Container.Image)
	}
}
