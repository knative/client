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

package channel

import (
	"bytes"

	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"

	"knative.dev/client/pkg/commands"
	clientv1beta1 "knative.dev/client/pkg/messaging/v1"
	eventingduck "knative.dev/eventing/pkg/apis/duck/v1"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

// TODO: Remove that blankConfig hack for tests in favor of overwriting GetConfig()
func init() {
	var err error
	blankConfig, err = clientcmd.NewClientConfigFromBytes([]byte(`kind: Config
version: v1
users:
- name: u
clusters:
- name: c
  cluster:
    server: example.com
contexts:
- name: x
  context:
    user: u
    cluster: c
current-context: x
`))
	if err != nil {
		panic(err)
	}
}

func executeChannelCommand(channelClient clientv1beta1.KnChannelsClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output

	cmd := NewChannelCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	channelClientFactory = func(config clientcmd.ClientConfig, namespace string) (clientv1beta1.KnChannelsClient, error) {
		return channelClient, nil
	}
	defer cleanupChannelMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupChannelMockClient() {
	channelClientFactory = nil
}

func createChannel(name, namespace string, gvk *schema.GroupVersionKind) *messagingv1.Channel {
	return clientv1beta1.NewChannelBuilder(name, namespace).Type(gvk).Build()
}

func createChannelWithStatus(name string, namespace string, gvk *schema.GroupVersionKind) *messagingv1.Channel {
	channel := clientv1beta1.NewChannelBuilder(name, namespace).Type(gvk).Build()
	channel.Status = messagingv1.ChannelStatus{
		ChannelableStatus: eventingduck.ChannelableStatus{
			AddressStatus: duckv1.AddressStatus{
				Address: &duckv1.Addressable{
					URL: &apis.URL{Scheme: "http", Host: "pipe-channel.test"},
				},
			},
		},
	}
	return channel
}
