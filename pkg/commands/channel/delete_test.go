// Copyright © 2020 The Knative Authors
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

package channel

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	v1beta1 "knative.dev/client/pkg/messaging/v1"
	"knative.dev/client/pkg/util"
)

func TestDeleteChannelErrorCase(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t, "test")
	cRecorder := cClient.Recorder()
	_, err := executeChannelCommand(cClient, "delete")
	assert.Error(t, err, "'kn channel delete' requires the channel name as single argument")
	cRecorder.Validate()
}

func TestDeleteWithError(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t, "test")
	cRecorder := cClient.Recorder()
	cRecorder.DeleteChannel("pipe", errors.New("not found"))
	_, err := executeChannelCommand(cClient, "delete", "pipe")
	assert.ErrorContains(t, err, "not found")
	cRecorder.Validate()
}

func TestChannelDelete(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t, "test")
	cRecorder := cClient.Recorder()
	cRecorder.DeleteChannel("pipe", nil)
	out, err := executeChannelCommand(cClient, "delete", "pipe")
	assert.NilError(t, err)
	assert.Assert(t, util.ContainsAll(out, "deleted", "pipe", "test"))
	cRecorder.Validate()
}
