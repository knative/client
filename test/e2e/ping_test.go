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

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestSourcePing(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("Creating a testservice")
	serviceCreate(t, it, r, "testsvc0")

	t.Log("create Ping sources with a sink to a service")

	pingSourceCreate(t, it, r, "testpingsource0", "* * * * */1", "ping", "svc:testsvc0")
	pingSourceListOutputName(t, it, r, "testpingsource0")

	t.Log("delete Ping sources")
	pingSourceDelete(t, it, r, "testpingsource0")

	t.Log("create Ping source with a missing sink service")
	pingSourceCreateMissingSink(t, it, r, "testpingsource1", "* * * * */1", "ping", "svc:unknown")

	t.Log("update Ping source sink service")
	pingSourceCreate(t, it, r, "testpingsource2", "* * * * */1", "ping", "svc:testsvc0")
	serviceCreate(t, it, r, "testsvc1")
	pingSourceUpdateSink(t, it, r, "testpingsource2", "svc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := getResourceFieldsWithJSONPath(t, it, "pingsource", "testpingsource2", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")

	t.Log("verify Ping source description")
	mymsg := "This is a message from Ping."
	pingSourceCreate(t, it, r, "testpingsource3", "*/1 * * * *", mymsg, "svc:testsvc1")
	verifyPingSourceDescribe(t, it, r, "testpingsource3", "*/1 * * * *", mymsg, "testsvc1")
}

func pingSourceCreate(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := it.Kn().Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "created", "namespace", it.Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceDelete(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string) {
	out := it.Kn().Run("source", "ping", "delete", sourceName)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "deleted", "namespace", it.Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceListOutputName(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, pingSource string) {
	out := it.Kn().Run("source", "ping", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, pingSource))
}

func pingSourceCreateMissingSink(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := it.Kn().Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(t, util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
	r.AssertError(out)
}

func pingSourceUpdateSink(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, sink string) {
	out := it.Kn().Run("source", "ping", "update", sourceName, "--sink", sink)
	assert.Check(t, util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", it.Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceCreateWithResources(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string, sa string, requestcpu string, requestmm string, limitcpu string, limitmm string) {
	out := it.Kn().Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink, "--service-account", sa,
		"--requests-cpu", requestcpu, "--requests-memory", requestmm, "--limits-cpu", limitcpu, "--limits-memory", limitmm)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "created", "namespace", it.Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceUpdateResources(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, requestcpu string, requestmm string, limitcpu string, limitmm string) {
	out := it.Kn().Run("source", "ping", "update", sourceName,
		"--requests-cpu", requestcpu, "--requests-memory", requestmm, "--limits-cpu", limitcpu, "--limits-memory", limitmm)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, sourceName, "updated", "namespace", it.Kn().Namespace()))
	r.AssertNoError(out)
}

func verifyPingSourceDescribe(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := it.Kn().Run("source", "ping", "describe", sourceName)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, sourceName, schedule, data, sink))
	r.AssertNoError(out)
}
