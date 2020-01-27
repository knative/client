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
	"testing"

	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	dynamic_fake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/sources/v1alpha1"

	serving_v1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	"knative.dev/client/pkg/util"
)

var testGvk = schema.GroupVersionKind{"apps", "v1", "deployment"}

func TestSimpleCreateBinding(t *testing.T) {
	mysvc := &serving_v1alpha1.Service{
		TypeMeta:   v1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1alpha1"},
		ObjectMeta: v1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default", mysvc)

	bindingClient := v1alpha1.NewMockKnSinkBindingClient(t)
	bindingRecorder := bindingClient.Recorder()
	bindingRecorder.CreateSinkBinding(createSinkBinding("testbinding", "mysvc", testGvk, "mydeploy"), nil)

	out, err := executeSinkBindingCommand(bindingClient, dynamicClient, "testbinding", "--sink", "svc:mysvc", "--subject", "deployment:apps/v1:mydeploy")
	assert.NilError(t, err, "Source should have been created")
	util.ContainsAll(out, "created", "default", "testbinding")

	bindingRecorder.Validate()
}

func TestNoSinkError(t *testing.T) {
	bindingClient := v1alpha1.NewMockKnSinkBindingClient(t)
	dynamicClient := dynamic_fake.CreateFakeKnDynamicClient("default")

	_, err := executeSinkBindingCommand(bindingClient, dynamicClient, "testbinding", "--sink", "svc:mysvc", "--subject", "deployment:apps/v1:app=myapp")
	assert.ErrorContains(t, err, "mysvc")
	assert.ErrorContains(t, err, "not found")
}

func TestNoSinkGivenError(t *testing.T) {
	out, err := executeSinkBindingCommand(nil, nil, "testbinding", "--subject", "deployment:apps/v1:app=myapp")
	assert.ErrorContains(t, err, "sink")
	assert.ErrorContains(t, err, "required")
	assert.Assert(t, util.ContainsAll(out, "not set", "required"))
}
