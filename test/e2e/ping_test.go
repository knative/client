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

//go:build e2e && !serving
// +build e2e,!serving

package e2e

import (
	"testing"

	"gotest.tools/v3/assert"

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

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("Creating a testservice")
	test.ServiceCreate(r, "testsvc0")

	t.Log("create Ping sources with a sink to a service")

	pingSourceCreate(r, "testpingsource0", "* * * * */1", "ping", "ksvc:testsvc0")
	pingSourceListOutputName(r, "testpingsource0")

	t.Log("delete Ping sources")
	pingSourceDelete(r, "testpingsource0")

	t.Log("create Ping source with a missing sink service")
	pingSourceCreateMissingSink(r, "testpingsource1", "* * * * */1", "ping", "ksvc:unknown")

	t.Log("update Ping source sink service")
	pingSourceCreate(r, "testpingsource2", "* * * * */1", "ping", "ksvc:testsvc0")
	test.ServiceCreate(r, "testsvc1")
	pingSourceUpdateSink(r, "testpingsource2", "ksvc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.GetResourceFieldsWithJSONPath(t, it, "pingsource", "testpingsource2", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")

	t.Log("verify Ping source description")
	mymsg := "This is a message from Ping."
	pingSourceCreate(r, "testpingsource3", "*/1 * * * *", mymsg, "ksvc:testsvc1")
	verifyPingSourceDescribe(r, "testpingsource3", "*/1 * * * *", mymsg, "testsvc1")

	t.Log("create Ping source with base64 data without encoding flag")
	pingSourceCreate(r, "testpingsource4", "* * * * */1", "cGluZw==", "ksvc:testsvc1")
	verifyPingSourceDescribe(r, "testpingsource4", "* * * * */1", "cGluZw==", "testsvc1")

	t.Log("create Ping source with text data and encoding flag")
	pingSourceCreateWithEncoding(r, "testpingsource6", "* * * * */1", "cGluZw==", "ksvc:testsvc1", "text")
	verifyPingSourceDescribe(r, "testpingsource6", "* * * * */1", "cGluZw==", "testsvc1")

	t.Log("update Ping source with base64 data and encoding flag")
	pingSourceUpdateData(r, "testpingsource6", "cGluZw==", "base64")
	verifyPingSourceDescribe(r, "testpingsource6", "* * * * */1", "cGluZw==", "testsvc1")

	t.Log("update Ping source with base64 data and encoding flag with error")
	pingSourceUpdateData(r, "testpingsource6", "cGluZw==", "base65")
}

func pingSourceCreate(r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := r.KnTest().Kn().Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "created", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceCreateWithEncoding(r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string, encoding string) {
	out := r.KnTest().Kn().Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink, "--encoding", encoding)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "created", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceDelete(r *test.KnRunResultCollector, sourceName string) {
	out := r.KnTest().Kn().Run("source", "ping", "delete", sourceName)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "ping", "source", sourceName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceListOutputName(r *test.KnRunResultCollector, pingSource string) {
	out := r.KnTest().Kn().Run("source", "ping", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, pingSource))
}

func pingSourceCreateMissingSink(r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := r.KnTest().Kn().Run("source", "ping", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
	r.AssertError(out)
}

func pingSourceUpdateSink(r *test.KnRunResultCollector, sourceName string, sink string) {
	out := r.KnTest().Kn().Run("source", "ping", "update", sourceName, "--sink", sink)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}

func pingSourceUpdateData(r *test.KnRunResultCollector, sourceName string, data string, encoding string) {
	out := r.KnTest().Kn().Run("source", "ping", "update", sourceName, "--data", data, "--encoding", encoding)
	if encoding != "text" && encoding != "base64" {
		assert.Check(r.T(), util.ContainsAll(out.Stderr, "cannot update PingSource", "invalid value"))
		return
	}
	assert.Check(r.T(), util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}

func verifyPingSourceDescribe(r *test.KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := r.KnTest().Kn().Run("source", "ping", "describe", sourceName)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, sourceName, schedule, data, sink))
	r.AssertNoError(out)
}
