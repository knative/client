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

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/client/pkg/messaging/v1beta1"

	"knative.dev/client/pkg/util"
)

func TestDescribeChannelErrorCase(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	_, err := executeChannelCommand(cClient, "describe")
	assert.Error(t, err, "'kn channel describe' requires the channel name given as single argument")
	cRecorder.Validate()
}

func TestDescribeChannelErrorCaseNotFound(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.GetChannel("pipe", nil, errors.New("not found"))
	_, err := executeChannelCommand(cClient, "describe", "pipe")
	assert.Error(t, err, "not found")
	cRecorder.Validate()
}

func TestDescribeChannel(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.GetChannel("pipe", createChannel("pipe", &schema.GroupVersionKind{"messaging.knative.dev", "v1beta1", "InMemoryChannel"}), nil)
	out, err := executeChannelCommand(cClient, "describe", "pipe")
	assert.NilError(t, err, "channel should be described")
	assert.Assert(t, util.ContainsAll(out, "messaging.knative.dev", "v1beta1", "InMemoryChannel", "pipe"))
	cRecorder.Validate()
}
