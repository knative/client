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
	"knative.dev/client/pkg/util"
)

func TestInjectBrokerTrigger(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	assert.NilError(t, err)

	test.serviceCreate(t, r, "sinksvc0")
	test.serviceCreate(t, r, "sinksvc1")

	t.Log("create triggers and list them")
	test.triggerCreateWithInject(t, r, "trigger1", "sinksvc0", []string{"a=b"})
	test.triggerCreateWithInject(t, r, "trigger2", "sinksvc1", []string{"type=knative.dev.bar", "source=ping"})
	test.verifyTriggerList(t, r, "trigger1", "trigger2")
	test.triggerDelete(t, r, "trigger1")
	test.triggerDelete(t, r, "trigger2")

	t.Log("create trigger with error")
	out := test.kn.Run("trigger", "create", "errorTrigger", "--broker", "mybroker", "--inject-broker",
		"--sink", "svc:sinksvc0", "--filter", "a=b")
	r.AssertError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stderr, "broker", "name", "'default'", "--inject-broker", "flag"))
}

func (test *e2eTest) triggerCreateWithInject(t *testing.T, r *KnRunResultCollector, name string, sinksvc string, filters []string) {
	args := []string{"trigger", "create", name, "--broker", "default", "--inject-broker", "--sink", "svc:" + sinksvc}
	for _, v := range filters {
		args = append(args, "--filter", v)
	}
	out := test.kn.Run(args...)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "created", "namespace", test.kn.namespace))
}
