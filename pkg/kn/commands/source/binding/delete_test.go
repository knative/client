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

	"gotest.tools/v3/assert"

	"knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
)

func TestSimpleDelete(t *testing.T) {

	bindingClient := v1.NewMockKnSinkBindingClient(t, "mynamespace")

	bindingRecorder := bindingClient.Recorder()
	bindingRecorder.DeleteSinkBinding("mybinding", nil)

	out, err := executeSinkBindingCommand(bindingClient, nil, "delete", "mybinding")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "deleted", "mynamespace", "mybinding", "Sink binding"))

	bindingRecorder.Validate()
}

func TestDeleteWithError(t *testing.T) {

	bindingClient := v1.NewMockKnSinkBindingClient(t, "mynamespace")

	bindingRecorder := bindingClient.Recorder()
	bindingRecorder.DeleteSinkBinding("mybinding", errors.New("no such sink binding mybinding"))

	out, err := executeSinkBindingCommand(bindingClient, nil, "delete", "mybinding")
	assert.ErrorContains(t, err, "mybinding")
	assert.Assert(t, util.ContainsAll(out, "no such", "mybinding"))

	bindingRecorder.Validate()
}
