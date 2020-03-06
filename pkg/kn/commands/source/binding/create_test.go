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

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/sources/v1alpha2"

	"knative.dev/client/pkg/util"
)

func TestSimpleCreateBinding(t *testing.T) {
	mysvc := createService("mysvc")
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc)

	bindingClient := v1alpha2.NewMockKnSinkBindingClient(t)
	bindingRecorder := bindingClient.Recorder()
	bindingRecorder.CreateSinkBinding(createSinkBinding("testbinding", "mysvc", deploymentGvk, "mydeploy", map[string]string{"bla": "blub", "foo": "bar"}), nil)

	out, err := executeSinkBindingCommand(bindingClient, dynamicClient, "create", "testbinding", "--sink", "svc:mysvc", "--subject", "deployment:apps/v1:mydeploy", "--ce-override", "bla=blub", "--ce-override", "foo=bar")
	assert.NilError(t, err, "Source should have been created")
	util.ContainsAll(out, "created", "default", "testbinding")

	bindingRecorder.Validate()
}

func TestNoSinkError(t *testing.T) {
	bindingClient := v1alpha2.NewMockKnSinkBindingClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	_, err := executeSinkBindingCommand(bindingClient, dynamicClient, "create", "testbinding", "--sink", "svc:mysvc", "--subject", "deployment:apps/v1:app=myapp")
	assert.ErrorContains(t, err, "mysvc")
	assert.ErrorContains(t, err, "not found")
}

func TestNoSinkGivenError(t *testing.T) {
	out, err := executeSinkBindingCommand(nil, nil, "create", "testbinding", "--subject", "deployment:apps/v1:app=myapp")
	assert.ErrorContains(t, err, "sink")
	assert.ErrorContains(t, err, "required")
	assert.Assert(t, util.ContainsAll(out, "not set", "required"))
}
