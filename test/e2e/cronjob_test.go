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

func TestSourceCronJob(t *testing.T) {
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

	t.Log("create cronJob sources with a sink to a service")

	test.cronJobSourceCreate(t, r, "testcronjobsource0", "* * * * */1", "ping", "svc:testsvc0")

	t.Log("delete cronJob sources")
	test.cronJobSourceDelete(t, r, "testcronjobsource0")

	t.Log("create cronJob source with a missing sink service")
	test.cronJobSourceCreateMissingSink(t, r, "testcronjobsource1", "* * * * */1", "ping", "svc:unknown")

	t.Log("update cronJob source sink service")
	test.cronJobSourceCreate(t, r, "testcronjobsource2", "* * * * */1", "ping", "svc:testsvc0")
	test.serviceCreate(t, r, "testsvc1")
	test.cronJobSourceUpdateSink(t, r, "testcronjobsource2", "svc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.getResourceFieldsWithJSONPath("cronjobsource", "testcronjobsource2", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")

	t.Log("create cronJob source with service account and resources")
	test.cronJobSourceCreateWithResources(t, r, "testcronjobsource3", "* * * * */1", "ping", "svc:testsvc0", "default", "100m", "128Mi", "200m", "256Mi")
	test.verifyCronJobSourceDescribe(t, r, "testcronjobsource3", "* * * * */1", "ping", "testsvc0", "default", "100m", "128Mi", "200m", "256Mi")
	test.cronJobSourceUpdateResources(t, r, "testcronjobsource3", "101m", "129Mi", "201m", "257Mi")
	test.verifyCronJobSourceDescribe(t, r, "testcronjobsource3", "* * * * */1", "ping", "testsvc0", "default", "101m", "129Mi", "201m", "257Mi")
}

func (test *e2eTest) cronJobSourceCreate(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := test.kn.Run("source", "cronjob", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "cronjob", "source", sourceName, "created", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) cronJobSourceDelete(t *testing.T, r *KnRunResultCollector, sourceName string) {
	out := test.kn.Run("source", "cronjob", "delete", sourceName)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "cronjob", "source", sourceName, "deleted", "namespace", test.kn.namespace))
	r.AssertNoError(out)

}

func (test *e2eTest) cronJobSourceCreateMissingSink(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string) {
	out := test.kn.Run("source", "cronjob", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink)
	assert.Check(t, util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
	r.AssertError(out)
}

func (test *e2eTest) cronJobSourceUpdateSink(t *testing.T, r *KnRunResultCollector, sourceName string, sink string) {
	out := test.kn.Run("source", "cronjob", "update", sourceName, "--sink", sink)
	assert.Check(t, util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) cronJobSourceCreateWithResources(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string, sa string, requestcpu string, requestmm string, limitcpu string, limitmm string) {
	out := test.kn.Run("source", "cronjob", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink, "--service-account", sa,
		"--requests-cpu", requestcpu, "--requests-memory", requestmm, "--limits-cpu", limitcpu, "--limits-memory", limitmm)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "cronjob", "source", sourceName, "created", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) cronJobSourceUpdateResources(t *testing.T, r *KnRunResultCollector, sourceName string, requestcpu string, requestmm string, limitcpu string, limitmm string) {
	out := test.kn.Run("source", "cronjob", "update", sourceName,
		"--requests-cpu", requestcpu, "--requests-memory", requestmm, "--limits-cpu", limitcpu, "--limits-memory", limitmm)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, sourceName, "updated", "namespace", test.kn.namespace))
	r.AssertNoError(out)
}

func (test *e2eTest) verifyCronJobSourceDescribe(t *testing.T, r *KnRunResultCollector, sourceName string, schedule string, data string, sink string, sa string, requestcpu string, requestmm string, limitcpu string, limitmm string) {
	out := test.kn.Run("source", "cronjob", "describe", sourceName)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, sourceName, schedule, data, sink, sa, requestcpu, requestmm, limitcpu, limitmm))
	r.AssertNoError(out)
}
