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
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/client/pkg/messaging/v1beta1"

	"knative.dev/client/pkg/util"
)

func TestCreateChannelErrorCase(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	_, err := executeChannelCommand(cClient, "create")
	assert.Error(t, err, "'kn channel create' requires the channel name given as single argument")
	cRecorder.Validate()
}

func TestCreateChannelErrorCaseTypeFormat(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	_, err := executeChannelCommand(cClient, "create", "pipe", "--type", "foo::bar")
	assert.Error(t, err, "Error: incorrect value 'foo::bar' for '--type', must be in the format 'Group:Version:Kind' or configure an alias in kn config")
	cRecorder.Validate()
}

func TestCreateChannelDefaultChannel(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.CreateChannel(createChannel("pipe", nil), nil)
	out, err := executeChannelCommand(cClient, "create", "pipe")
	assert.NilError(t, err, "channel should be created")
	assert.Assert(t, util.ContainsAll(out, "created", "pipe", "default"))
	cRecorder.Validate()
}

func TestCreateChannelWithTypeFlagInMemoryChannel(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.CreateChannel(createChannel("pipe", &schema.GroupVersionKind{Group: "messaging.knative.dev", Version: "v1beta1", Kind: "InMemoryChannel"}), nil)
	out, err := executeChannelCommand(cClient, "create", "pipe", "--type", "imcv1beta1")
	assert.NilError(t, err, "channel should be created")
	assert.Assert(t, util.ContainsAll(out, "created", "pipe", "default"))
	cRecorder.Validate()
}
