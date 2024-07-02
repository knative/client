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
	"encoding/json"
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	clientv1 "knative.dev/client/pkg/messaging/v1"
	"knative.dev/client/pkg/util"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
)

func TestDescribeSubscriptionErrorCase(t *testing.T) {
	cClient := clientv1.NewMockKnSubscriptionsClient(t)
	cRecorder := cClient.Recorder()
	_, err := executeSubscriptionCommand(cClient, nil, "describe")
	assert.Error(t, err, "'kn subscription describe' requires the subscription name given as single argument")
	cRecorder.Validate()
}

func TestDescribeSubscriptionErrorCaseNotFound(t *testing.T) {
	cClient := clientv1.NewMockKnSubscriptionsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.GetSubscription("sub0", nil, errors.New("not found"))
	_, err := executeSubscriptionCommand(cClient, nil, "describe", "sub0")
	assert.Error(t, err, "not found")
	cRecorder.Validate()
}

func TestDescribeSubscription(t *testing.T) {
	cClient := clientv1.NewMockKnSubscriptionsClient(t)
	cRecorder := cClient.Recorder()

	subscription := createSubscription("sub0", "imc0", "ksvc0", "b0", "b1")

	t.Run("default output", func(t *testing.T) {
		cRecorder.GetSubscription("sub0", subscription, nil)
		out, err := executeSubscriptionCommand(cClient, nil, "describe", "sub0")
		assert.NilError(t, err, "subscription should be described")
		assert.Assert(t, util.ContainsAll(out,
			"sub0",
			"Channel", "imc0", "messaging.knative.dev", "v1", "InMemoryChannel",
			"Subscriber", "ksvc0", "serving.knative.dev", "v1", "Service",
			"Reply", "b0", "eventing.knative.dev", "v1", "Broker",
			"DeadLetterSink", "b1"))
	})

	t.Run("json format output", func(t *testing.T) {
		cRecorder.GetSubscription("sub0", subscription, nil)
		out, err := executeSubscriptionCommand(cClient, nil, "describe", "sub0", "-o", "json")
		assert.NilError(t, err, "subscription should be described")

		result := &messagingv1.Subscription{}
		err = json.Unmarshal([]byte(out), result)
		assert.NilError(t, err, "subscription should be in json format")
		assert.DeepEqual(t, subscription, result)
	})

	cRecorder.Validate()
}
