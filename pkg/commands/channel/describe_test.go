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
	"encoding/json"
	"errors"
	"testing"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clientv1 "knative.dev/client/pkg/messaging/v1"
	"knative.dev/client/pkg/util"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
)

func TestDescribeChannelErrorCase(t *testing.T) {
	cClient := clientv1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	_, err := executeChannelCommand(cClient, "describe")
	assert.Error(t, err, "'kn channel describe' requires the channel name given as single argument")
	cRecorder.Validate()
}

func TestDescribeChannelErrorCaseNotFound(t *testing.T) {
	cClient := clientv1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.GetChannel("pipe", nil, errors.New("not found"))
	_, err := executeChannelCommand(cClient, "describe", "pipe")
	assert.Error(t, err, "not found")
	cRecorder.Validate()
}

func TestDescribeChannel(t *testing.T) {
	cClient := clientv1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()

	channel := createChannel("pipe", "default", &schema.GroupVersionKind{Group: "messaging.knative.dev", Version: "v1", Kind: "InMemoryChannel"})

	t.Run("default output", func(t *testing.T) {
		cRecorder.GetChannel("pipe", channel, nil)
		out, err := executeChannelCommand(cClient, "describe", "pipe")
		assert.NilError(t, err, "channel should be described")
		assert.Assert(t, util.ContainsAll(out, "messaging.knative.dev", "v1", "InMemoryChannel", "pipe"))
	})

	t.Run("json format output", func(t *testing.T) {
		cRecorder.GetChannel("pipe", channel, nil)
		out, err := executeChannelCommand(cClient, "describe", "pipe", "-o", "json")
		assert.NilError(t, err, "channel should be described")

		result := &messagingv1.Channel{}
		err = json.Unmarshal([]byte(out), result)
		assert.NilError(t, err, "channel should be in json format")
		assert.Assert(t, util.ContainsAll(out, "messaging.knative.dev", "v1", "InMemoryChannel", "pipe"))
		assert.DeepEqual(t, channel, result)
	})

	cRecorder.Validate()
}

func TestDescribeChannelURL(t *testing.T) {
	cClient := clientv1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.GetChannel("pipe", createChannelWithStatus("pipe", "default", &schema.GroupVersionKind{Group: "messaging.knative.dev", Version: "v1", Kind: "InMemoryChannel"}), nil)
	out, err := executeChannelCommand(cClient, "describe", "pipe", "-o", "url")
	assert.NilError(t, err, "channel should be described with url as output")
	assert.Assert(t, util.ContainsAll(out, "pipe-channel.test"))
	cRecorder.Validate()
}
