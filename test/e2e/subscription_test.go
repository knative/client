// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build e2e && !serving
// +build e2e,!serving

package e2e

import (
	"testing"

	"gotest.tools/v3/assert"

	"knative.dev/client-pkg/pkg/util"
	"knative.dev/client-pkg/pkg/util/test"
)

func TestSubscriptions(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("Create a subscription with all the flags")
	test.ChannelCreate(r, "c0")
	test.ServiceCreate(r, "svc0")
	test.ServiceCreate(r, "svc1")
	test.ServiceCreate(r, "svc2")
	test.SubscriptionCreate(r, "sub0", "--channel", "c0", "--sink", "ksvc:svc0", "--sink-reply", "ksvc:svc1", "--sink-dead-letter", "ksvc:svc2")

	t.Log("Update a subscription")
	test.ServiceCreate(r, "svc3")
	test.SubscriptionUpdate(r, "sub0", "--sink", "ksvc:svc3")

	t.Log("List subscriptions")
	slist := test.SubscriptionList(r)
	assert.Check(t, util.ContainsAll(slist, "NAME", "CHANNEL", "SUBSCRIBER", "REPLY", "DEAD LETTER SINK", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(slist, "sub0", "c0", "ksvc:svc3", "ksvc:svc1", "ksvc:svc2", "True"))

	t.Log("Describe subscription")
	sdesc := test.SubscriptionDescribe(r, "sub0")
	assert.Check(t, util.ContainsAll(sdesc, "sub0", "Age", "Channel", "Channel", "c0", "Subscriber", "svc3", "Resource", "Service", "serving.knative.dev/v1", "Reply", "svc1", "DeadLetterSink", "svc2", "Conditions"))

	t.Log("Delete subscription")
	test.SubscriptionDelete(r, "sub0")
	test.ServiceDelete(r, "svc0")
	test.ServiceDelete(r, "svc1")
	test.ServiceDelete(r, "svc2")
	test.ServiceDelete(r, "svc3")
	test.ChannelDelete(r, "c0")
}
