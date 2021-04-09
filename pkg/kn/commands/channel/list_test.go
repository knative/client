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

	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"

	clientmessagingv1 "knative.dev/client/pkg/messaging/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"

	"knative.dev/client/pkg/util"
)

func TestChannelListNoChannelsFound(t *testing.T) {
	cClient := clientmessagingv1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.ListChannel(nil, nil)
	out, err := executeChannelCommand(cClient, "list")
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "No channels found"))
	cRecorder.Validate()
}

func TestChannelListNoChannelsFoundWithOutputSet(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.ListChannel(nil, nil)
	out, err := executeChannelCommand(cClient, "list", "-o", "json")
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "\"apiVersion\": \"messaging.knative.dev/v1beta1\"", "\"kind\": \"ChannelList\"", "\"items\": []"))
	cRecorder.Validate()
}

func TestChannelListEmptyWithOutputSet(t *testing.T) {
	cClient := v1beta1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	channelList := &messagingv1beta1.ChannelList{}
	err := util.UpdateGroupVersionKindWithScheme(channelList, messagingv1beta1.SchemeGroupVersion, scheme.Scheme)
	assert.NilError(t, err)
	cRecorder.ListChannel(channelList, nil)
	out, err := executeChannelCommand(cClient, "list", "-o", "json")
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "\"apiVersion\": \"messaging.knative.dev/v1beta1\"", "\"kind\": \"ChannelList\"", "\"items\": []"))
	cRecorder.Validate()
}

func TestChannelList(t *testing.T) {
	cClient := clientmessagingv1.NewMockKnChannelsClient(t)
	cRecorder := cClient.Recorder()
	clist := &messagingv1.ChannelList{}
	clist.Items = []messagingv1.Channel{
		*createChannel("c0", "default", &schema.GroupVersionKind{Group: "messaging.knative.dev", Version: "v1", Kind: "InMemoryChannel"}),
		*createChannel("c1", "default", &schema.GroupVersionKind{Group: "messaging.knative.dev", Version: "v1", Kind: "InMemoryChannel"}),
	}
	cRecorder.ListChannel(clist, nil)
	out, err := executeChannelCommand(cClient, "list")
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "c0", "c1"))
	cRecorder.Validate()
}
