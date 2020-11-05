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

// +build e2e
// +build !serving

package e2e

import (
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test"
)

func TestSink(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	// create broker
	err = test.LabelNamespaceForDefaultBroker(r)
	assert.NilError(t, err)
	defer test.UnlabelNamespaceForDefaultBroker(r)

	t.Log("Create Ping source with a sink to the default broker")
	pingSourceCreate(r, "testpingsource0", "* * * * */1", "ping", "broker:default")

	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.GetResourceFieldsWithJSONPath(t, it, "pingsource", "testpingsource0", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "default")

	// create a channel
	err = test.ChannelCreate(r, "pipe")

	t.Log("Update Ping source with a sink to the channel")
	pingSourceUpdateSink(r, "testpingsource0", "channel:pipe")
	out, err := test.GetResourceFieldsWithJSONPath(t, it, "pingsource", "testpingsource0", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "pipe")

	t.Log("delete Ping sources")
	pingSourceDelete(r, "testpingsource0")
}
