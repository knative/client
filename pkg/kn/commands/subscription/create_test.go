/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package subscription

import (
	"testing"

	"gotest.tools/v3/assert"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	clientmessagingv1 "knative.dev/client/pkg/messaging/v1"
	"knative.dev/client/pkg/util"
)

func TestCreateSubscriptionErrorCase(t *testing.T) {
	cClient := clientmessagingv1.NewMockKnSubscriptionsClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	cRecorder := cClient.Recorder()
	_, err := executeSubscriptionCommand(cClient, dynamicClient, "create")
	assert.Error(t, err, "'kn subscription create' requires the subscription name given as single argument")
	cRecorder.Validate()
}

func TestCreateSubscriptionErrorCaseRequiredChannelFlag(t *testing.T) {
	cClient := clientmessagingv1.NewMockKnSubscriptionsClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	cRecorder := cClient.Recorder()
	_, err := executeSubscriptionCommand(cClient, dynamicClient, "create", "sub0")
	assert.Error(t, err, "'kn subscription create' requires the channel reference provided with --channel flag")
	cRecorder.Validate()
}

func TestCreateSubscriptionErrorCaseChannelFormat(t *testing.T) {
	cClient := clientmessagingv1.NewMockKnSubscriptionsClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	cRecorder := cClient.Recorder()
	_, err := executeSubscriptionCommand(cClient, dynamicClient, "create", "sub0", "--channel", "foo::bar")
	assert.Error(t, err, "Error: incorrect value 'foo::bar' for '--channel', must be in the format 'Group:Version:Kind:Name' or configure an alias in kn config and refer as: '--channel ALIAS:NAME'")
	cRecorder.Validate()
}

func TestCreateSubscription(t *testing.T) {
	cClient := clientmessagingv1.NewMockKnSubscriptionsClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default",
		createService("ksvc0"),
		createBroker("b0"),
		createBroker("b1"))

	cRecorder := cClient.Recorder()
	cRecorder.CreateSubscription(createSubscription("sub0",
		"imc0",
		"ksvc0",
		"b0",
		"b1"),
		nil)

	out, err := executeSubscriptionCommand(cClient, dynamicClient, "create", "sub0",
		"--channel", "imcv1beta1:imc0",
		"--sink", "ksvc0",
		"--sink-reply", "broker:b0",
		"--sink-dead-letter", "broker:b1")
	assert.NilError(t, err, "subscription should be created")
	assert.Assert(t, util.ContainsAll(out, "created", "sub0", "default"))
	cRecorder.Validate()
}
