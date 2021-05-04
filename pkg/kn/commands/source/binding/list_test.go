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

	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"gotest.tools/v3/assert"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"

	clientv1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
)

func TestListBindingSimple(t *testing.T) {
	bindingClient := clientv1.NewMockKnSinkBindingClient(t)

	bindingRecorder := bindingClient.Recorder()
	binding := createSinkBinding("testbinding", "mysvc", deploymentGvk, "mydeploy", "default", nil)
	bindingList := v1.SinkBindingList{
		Items: []v1.SinkBinding{
			*binding,
		},
	}
	bindingRecorder.ListSinkBindings(&bindingList, nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "NAME", "SUBJECT", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "testbinding", "deployment:apps/v1:mydeploy", "mysvc"))

	bindingRecorder.Validate()
}

func TestListBindingEmpty(t *testing.T) {
	bindingClient := clientv1.NewMockKnSinkBindingClient(t)

	bindingRecorder := bindingClient.Recorder()
	bindingList := v1.SinkBindingList{}
	bindingRecorder.ListSinkBindings(&bindingList, nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsNone(out, "NAME", "SUBJECT", "SINK", "AGE", "CONDITIONS", "READY", "REASON"))
	assert.Assert(t, util.ContainsAll(out, "No", "sink binding", "found"))

	bindingRecorder.Validate()
}

func TestListBindingEmptyWithJsonOutput(t *testing.T) {
	bindingClient := clientv1.NewMockKnSinkBindingClient(t)

	bindingRecorder := bindingClient.Recorder()
	bindingList := v1.SinkBindingList{}
	_ = util.UpdateGroupVersionKindWithScheme(&bindingList, v1.SchemeGroupVersion, scheme.Scheme)
	bindingRecorder.ListSinkBindings(&bindingList, nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "list", "-o", "json")
	assert.NilError(t, err, "Sources should be listed")
	assert.Assert(t, util.ContainsAll(out, "\"apiVersion\": \"sources.knative.dev/v1\"", "\"items\": []", "\"kind\": \"SinkBindingList\""))

	bindingRecorder.Validate()
}
