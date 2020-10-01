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

	"gotest.tools/assert"
	"knative.dev/client/pkg/messaging/v1beta1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	"knative.dev/client/pkg/util"
)

func TestUpdateSubscriptionErrorCase(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	cRecorder := cClient.Recorder()
	_, err := executeSubscriptionCommand(cClient, dynamicClient, "update")
	assert.Error(t, err, "'kn subscription update' requires the subscription name given as single argument")
	cRecorder.Validate()
}

func TestUpdateSubscriptionErrorCaseUnknownChannelFlag(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t)
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default")

	cRecorder := cClient.Recorder()
	_, err := executeSubscriptionCommand(cClient, dynamicClient, "update", "sub0", "--channel", "imc:i1")
	assert.Error(t, err, "unknown flag: --channel")
	cRecorder.Validate()
}

func TestUpdateSubscription(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t)
	sub0 := createSubscription("sub0", "imc0", "ksvc0", "", "")
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default",
		sub0,
		createService("ksvc1"),
		createBroker("b0"),
		createBroker("b1"))

	cRecorder := cClient.Recorder()
	cRecorder.GetSubscription("sub0", sub0, nil)
	cRecorder.UpdateSubscription(createSubscription("sub0",
		"imc0",
		"ksvc1",
		"b0",
		"b1"),
		nil)

	out, err := executeSubscriptionCommand(cClient, dynamicClient, "update", "sub0",
		"--sink", "ksvc1",
		"--sink-reply", "broker:b0",
		"--sink-dead-letter", "broker:b1")
	assert.NilError(t, err, "subscription should be updated")
	assert.Assert(t, util.ContainsAll(out, "updated", "sub0", "default"))
	cRecorder.Validate()
}
