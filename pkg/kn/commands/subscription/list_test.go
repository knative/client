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
	"strings"
	"testing"

	"knative.dev/eventing/pkg/client/clientset/versioned/scheme"

	"gotest.tools/v3/assert"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"

	v1beta1 "knative.dev/client/pkg/messaging/v1"
	"knative.dev/client/pkg/util"
)

func TestSubscriptionListNoSubscriptionsFound(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t)
	cRecorder := cClient.Recorder()
	cRecorder.ListSubscription(nil, nil)
	out, err := executeSubscriptionCommand(cClient, nil, "list")
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "No subscriptions found"))
	cRecorder.Validate()
}

func TestSubscriptionListNoSubscriptionsWithJsonOutput(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t)
	cRecorder := cClient.Recorder()
	clist := &messagingv1.SubscriptionList{}
	_ = util.UpdateGroupVersionKindWithScheme(clist, messagingv1.SchemeGroupVersion, scheme.Scheme)
	cRecorder.ListSubscription(clist, nil)
	out, err := executeSubscriptionCommand(cClient, nil, "list", "-o", "json")
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "\"apiVersion\": \"messaging.knative.dev/v1\"", "\"items\": []", "\"kind\": \"SubscriptionList\""))
	cRecorder.Validate()
}

func TestSubscriptionList(t *testing.T) {
	cClient := v1beta1.NewMockKnSubscriptionsClient(t)
	cRecorder := cClient.Recorder()
	clist := &messagingv1.SubscriptionList{}
	clist.Items = []messagingv1.Subscription{
		*createSubscription("s0", "imc0", "ksvc0", "b00", "b01"),
		*createSubscription("s1", "imc1", "ksvc1", "b10", "b11"),
		*createSubscription("s2", "imc2", "ksvc2", "b20", "b21"),
	}

	t.Run("default list output", func(t *testing.T) {
		cRecorder.ListSubscription(clist, nil)
		out, err := executeSubscriptionCommand(cClient, nil, "list")
		assert.NilError(t, err)
		ol := strings.Split(out, "\n")
		assert.Check(t, util.ContainsAll(ol[0], "NAME", "CHANNEL", "SUBSCRIBER", "REPLY", "DEAD LETTER SINK", "READY", "REASON"))
		assert.Check(t, util.ContainsAll(ol[1], "s0", "InMemoryChannel:imc0", "ksvc:ksvc0", "broker:b00", "broker:b01"))
		assert.Check(t, util.ContainsAll(ol[2], "s1", "imc1", "ksvc1", "b10", "b11"))
		assert.Check(t, util.ContainsAll(ol[3], "s2", "imc2", "ksvc2", "b20", "b21"))
	})

	t.Run("no headers list output", func(t *testing.T) {
		cRecorder.ListSubscription(clist, nil)
		out, err := executeSubscriptionCommand(cClient, nil, "list", "--no-headers")
		assert.NilError(t, err)
		ol := strings.Split(out, "\n")
		assert.Check(t, util.ContainsNone(ol[0], "NAME", "CHANNEL", "SUBSCRIBER", "REPLY", "DEAD LETTER SINK", "READY", "REASON"))
		assert.Check(t, util.ContainsAll(ol[0], "s0", "imc0", "ksvc0", "b00", "b01"))
	})

	cRecorder.Validate()
}
