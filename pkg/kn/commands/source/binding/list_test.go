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
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"

	clientv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

func TestListBindingSimple(t *testing.T) {
	bindingClient := clientv1alpha2.NewMockKnSinkBindingClient(t)

	bindingRecorder := bindingClient.Recorder()
	binding := createSinkBinding("testbinding", "mysvc", deploymentGvk, "mydeploy", nil)
	bindingList := v1alpha2.SinkBindingList{
		Items: []v1alpha2.SinkBinding{
			*binding,
		},
	}
	bindingRecorder.ListSinkBindings(&bindingList, nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	util.ContainsAll(out, "NAME", "SUBJECT", "SINK", "AGE", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "testbinding", "deployment:apps/v1:mydeploy", "mysvc")

	bindingRecorder.Validate()
}

func TestListBindingEmpty(t *testing.T) {
	bindingClient := clientv1alpha2.NewMockKnSinkBindingClient(t)

	bindingRecorder := bindingClient.Recorder()
	bindingList := v1alpha2.SinkBindingList{}
	bindingRecorder.ListSinkBindings(&bindingList, nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "list")
	assert.NilError(t, err, "Sources should be listed")
	util.ContainsNone(out, "NAME", "SUBJECT", "SINK", "AGE", "CONDITIONS", "READY", "REASON")
	util.ContainsAll(out, "No", "sink binding", "found")

	bindingRecorder.Validate()
}
