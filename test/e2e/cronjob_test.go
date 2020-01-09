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
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	test.serviceCreate(t, "testsvc0")

	t.Run("create cronJob sources with a sink to a service", func(t *testing.T) {
		test.cronJobSourceCreate(t, "testcronjobsource0", "* * * * */1", "ping", "svc:testsvc0")
	})

	t.Run("delete cronJob sources", func(t *testing.T) {
		test.cronJobSourceDelete(t, "testcronjobsource0")
	})

	t.Run("create cronJob source with a missing sink service", func(t *testing.T) {
		test.cronJobSourceCreateMissingSink(t, "testcronjobsource1", "* * * * */1", "ping", "svc:unknown")
	})

	t.Run("update cronJob source sink service", func(t *testing.T) {
		test.cronJobSourceCreate(t, "testcronjobsource2", "* * * * */1", "ping", "svc:testsvc0")
		test.serviceCreate(t, "testsvc1")
		test.cronJobSourceUpdateSink(t, "testcronjobsource2", "svc:testsvc1")
		jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
		out, err := test.getResourceFieldsWithJSONPath(t, "cronjobsource", "testcronjobsource2", jpSinkRefNameInSpec)
		assert.NilError(t, err)
		assert.Equal(t, out, "testsvc1")
	})
}

func (test *e2eTest) cronJobSourceCreate(t *testing.T, sourceName string, schedule string, data string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"source", "cronjob", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "cronjob", "source", sourceName, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) cronJobSourceDelete(t *testing.T, sourceName string) {
	out, err := test.kn.RunWithOpts([]string{"source", "cronjob", "delete", sourceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "cronjob", "source", sourceName, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) cronJobSourceCreateMissingSink(t *testing.T, sourceName string, schedule string, data string, sink string) {
	_, err := test.kn.RunWithOpts([]string{"source", "cronjob", "create", sourceName,
		"--schedule", schedule, "--data", data, "--sink", sink}, runOpts{NoNamespace: false, AllowError: true})
	assert.ErrorContains(t, err, "services.serving.knative.dev", "not found")
}

func (test *e2eTest) cronJobSourceUpdateSink(t *testing.T, sourceName string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"source", "cronjob", "update", sourceName, "--sink", sink}, runOpts{})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, sourceName, "updated", "namespace", test.kn.namespace))
}
