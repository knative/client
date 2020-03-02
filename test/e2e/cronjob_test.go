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
	"fmt"
	"strings"
	"testing"
	"time"

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
	//assert.Equal(t, out, "testsvc1")
	assert.Check(t, util.ContainsAllIgnoreCase(out, "testsvc1"))

	t.Log("verify event messages from cronjob source")
	mymsg := "This is a message from cronjob."
	test.serviceCreateEventDisplay(t, r, "displaysvc")
	test.cronJobSourceCreate(t, r, "testcronjobsource3", "*/1 * * * *", mymsg, "svc:displaysvc")
	results := test.kn.Run("source", "cronjob", "describe", "testcronjobsource3")
	r.AssertNoError(results)
	out, err = kubectl{test.kn.namespace}.Run("get", "pod", "-l", "sources.eventing.knative.dev/cronJobSource=testcronjobsource3", "-o", "yaml")
	t.Log(out)
	err = test.verifyEventDisplayLogs(t, "displaysvc", "This is a message from cronjob")
	assert.NilError(t, err)
}

func (test *e2eTest) verifyEventDisplayLogs(t *testing.T, svcname string, message string) error {
	var (
		retries int
		err     error
		out     string
	)

	selectorStr := fmt.Sprintf("serving.knative.dev/service=%s", svcname)
	for retries < 5 {
		out, err = kubectl{test.kn.namespace}.Run("logs", "-l", selectorStr, "-c", "user-container")
		if err != nil {
			t.Logf("error happens at kubectl logs -l %s -c -n %s: %v", selectorStr, test.kn.namespace, err)
		} else if err == nil && strings.Contains(out, message) {
			break
		} else {
			t.Logf("return from kubectl logs -l %s -n %s: %s", selectorStr, test.kn.namespace, out)
			out, err = kubectl{test.kn.namespace}.Run("get", "pods")
			t.Log(out)
			out, err = kubectl{test.kn.namespace}.Run("logs", "-l", "sources.eventing.knative.dev/cronJobSource=testcronjobsource3")
			t.Log(out)
		}
		retries++
		time.Sleep(2 * time.Minute)
	}

	if retries == 5 {
		return fmt.Errorf("Expected log incorrect after retry 5 times. Expecting to include:\n%s\n Instead found:\n%s\n", message, out)
	} else {
		return nil
	}
}

func (test *e2eTest) serviceCreateEventDisplay(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "create", serviceName, "--image", EventDisplayImage)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", test.kn.namespace, "ready"))
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
