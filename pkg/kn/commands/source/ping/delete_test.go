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

package ping

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"
	clientsourcesv1beta2 "knative.dev/client/pkg/sources/v1beta2"

	"knative.dev/client/pkg/util"
)

func TestSimpleDelete(t *testing.T) {
	pingClient := clientsourcesv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.DeletePingSource("testsource", nil)

	out, err := executePingSourceCommand(pingClient, nil, "delete", "testsource")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "deleted", "mynamespace", "testsource", "Ping"))

	pingRecorder.Validate()
}

func TestDeleteWithError(t *testing.T) {
	pingClient := clientsourcesv1beta2.NewMockKnPingSourceClient(t, "mynamespace")

	pingRecorder := pingClient.Recorder()
	pingRecorder.DeletePingSource("testsource", errors.New("no such Ping source testsource"))

	out, err := executePingSourceCommand(pingClient, nil, "delete", "testsource")
	assert.ErrorContains(t, err, "testsource")
	assert.Assert(t, util.ContainsAll(out, "Usage", "no such", "testsource"))

	pingRecorder.Validate()
}

func TestPingDeleteErrorForNoArgs(t *testing.T) {
	pingClient := clientsourcesv1beta2.NewMockKnPingSourceClient(t, "mynamespace")
	out, err := executePingSourceCommand(pingClient, nil, "delete")
	assert.ErrorContains(t, err, "single argument")
	assert.Assert(t, util.ContainsAll(out, "requires", "single argument"))
}
