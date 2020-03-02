// Copyright 2019 The Knative Authors

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

	"knative.dev/client/pkg/util"
)

func TestSourcePing(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("Creating a testservice")
	test.serviceCreate(t, r, "testsvc0")

	t.Log("create Ping sources with a sink to a service")

	test.pingSourceCreate(t, r, "testpingsource0", "* * * * */1", "ping", "svc:testsvc0")

	t.Log("delete Ping sources")
	test.pingSourceDelete(t, r, "testpingsource0")

	t.Log("create Ping source with a missing sink service")
	test.pingSourceCreateMissingSink(t, r, "testpingsource1", "* * * * */1", "ping", "svc:unknown")

	t.Log("update Ping source sink service")
	test.pingSourceCreate(t, r, "testpingsource2", "* * * * */1", "ping", "svc:testsvc0")
	test.serviceCreate(t, r, "testsvc1")
	test.pingSourceUpdateSink(t, r, "testpingsource2", "svc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.getResourceFieldsWithJSONPath("pingsource", "testpingsource2", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")

	t.Log("verify Ping source description")
	mymsg := "This is a message from Ping."
	test.pingSourceCreate(t, r, "testpingsource3", "*/1 * * * *", mymsg, "svc:testsvc1")
	test.verifyPingSourceDescribe(t, r, "testpingsource3", "*/1 * * * *", mymsg, "testsvc1")
}

func (test *e2eTest) pingSourceCreate(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := test.kn.Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "created", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) pingSourceDelete(t *testing.T, r *KnRunResultCollector, sourceName string) {
	out := test.kn.Run("source", "ping", "delete", sourceName)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "deleted", "namespace", test.kn.namespace))
	r.AssertNoError(out)

}

func (test *e2eTest) pingSourceCreateMissingSink(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := test.kn.Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(t, util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
	r.AssertError(out)
}

func (test *e2eTest) pingSourceUpdateSink(t *testing.T, r *KnRunResultCollector, sourceName string, sink string) {
	out := test.kn.Run("source", "ping", "update", sourceName, "--sink", sink)
	assert.Check(t, util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) pingSourceCreateWithResources(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string, sa string, requestcpu string, requestmm string, limitcpu string, limitmm string) {
	out := test.kn.Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink, "--service-account", sa,
		"--requests-cpu", requestcpu, "--requests-memory", requestmm, "--limits-cpu", limitcpu, "--limits-memory", limitmm)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "created", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) pingSourceUpdateResources(t *testing.T, r *KnRunResultCollector, sourceName string, requestcpu string, requestmm string, limitcpu string, limitmm string) {
	out := test.kn.Run("source", "ping", "update", sourceName,
		"--requests-cpu", requestcpu, "--requests-memory", requestmm, "--limits-cpu", limitcpu, "--limits-memory", limitmm)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, sourceName, "updated", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) verifyPingSourceDescribe(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := test.kn.Run("source", "ping", "describe", sourceName)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, sourceName, schedule, data, sink))
	r.AssertNoError(out)
}
