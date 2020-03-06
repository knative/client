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

package binding

import (
	"errors"
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	clientsourcesv1alpha1 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestSimpleBindingUpdate(t *testing.T) {
	sinkBindingClient := clientsourcesv1alpha1.NewMockKnSinkBindingClient(t)

	mysvc := createService("myscv")
	othersvc := createService("othersvc")

	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc, othersvc)

	bindingRecorder := sinkBindingClient.Recorder()
	ceOverrideMap := map[string]string{"bla": "blub", "foo": "bar"}
	bindingRecorder.GetSinkBinding("testbinding", createSinkBinding("testbinding", "mysvc", deploymentGvk, "mydeploy", ceOverrideMap), nil)
	bindingRecorder.UpdateSinkBinding(createSinkBinding("testbinding", "othersvc", deploymentGvk, "mydeploy", ceOverrideMap), nil)

	out, err := executeSinkBindingCommand(sinkBindingClient, dynamicClient, "update", "testbinding", "--sink", "svc:othersvc", "--ce-override", "bla=blub", "--ce-override", "foo=bar")
	assert.NilError(t, err)
	util.ContainsAll(out, "updated", "default", "testbinding", "foo", "bar")

	bindingRecorder.Validate()
}

func createService(name string) *servingv1.Service {
	mysvc := &servingv1.Service{
		TypeMeta:   v1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: "default"},
	}
	return mysvc
}

func TestUpdateError(t *testing.T) {
	sinkBindingClient := clientsourcesv1alpha1.NewMockKnSinkBindingClient(t)
	bindingRecorder := sinkBindingClient.Recorder()
	bindingRecorder.GetSinkBinding("testbinding", nil, errors.New("no such binding testbinding"))

	out, err := executeSinkBindingCommand(sinkBindingClient, nil, "update", "testbinding")
	assert.ErrorContains(t, err, "testbinding")
	util.ContainsAll(out, "testbinding", "name", "required")

	bindingRecorder.Validate()
}
