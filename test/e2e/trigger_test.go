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
	"strings"
	"testing"
	"time"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/client/pkg/util"
)

func TestBrokerTrigger(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	err = test.lableNamespaceForDefaultBroker(t)
	assert.NilError(t, err)
	defer test.unlableNamespaceForDefaultBroker(t)

	test.serviceCreate(t, r, "sinksvc0")
	test.serviceCreate(t, r, "sinksvc1")

	t.Log("create triggers and list them")
	test.triggerCreate(t, r, "trigger1", "sinksvc0", []string{"a=b"})
	test.triggerCreate(t, r, "trigger2", "sinksvc1", []string{"type=knative.dev.bar", "source=ping"})
	test.verifyTriggerList(t, r, "trigger1", "trigger2")
	test.triggerDelete(t, r, "trigger1")
	test.triggerDelete(t, r, "trigger2")

	t.Log("create a trigger and delete it")
	test.triggerCreate(t, r, "deltrigger", "sinksvc0", []string{"a=b"})
	test.triggerDelete(t, r, "deltrigger")
	test.verifyTriggerNotfound(t, r, "deltrigger")

	t.Log("create a trigger, describe and update it")
	test.triggerCreate(t, r, "updtrigger", "sinksvc0", []string{"a=b"})
	test.verifyTriggerDescribe(t, r, "updtrigger", "default", "sinksvc0", []string{"a", "b"})
	test.triggerUpdate(t, r, "updtrigger", "type=knative.dev.bar", "sinksvc1")
	test.verifyTriggerDescribe(t, r, "updtrigger", "default", "sinksvc1", []string{"a", "b", "type", "knative.dev.bar"})
	test.triggerDelete(t, r, "updtrigger")

	t.Log("create trigger with error return")
	test.triggerCreateMissingSink(t, r, "errtrigger", "notfound")
}

func (test *e2eTest) unlableNamespaceForDefaultBroker(t *testing.T) {
	_, err := kubectl{}.Run("label", "namespace", test.kn.namespace, "knative-eventing-injection-")
	if err != nil {
		t.Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection-'. Error: %s", test.kn.namespace, err.Error())
	}
}

func (test *e2eTest) lableNamespaceForDefaultBroker(t *testing.T) error {
	_, err := kubectl{}.Run("label", "namespace", test.kn.namespace, "knative-eventing-injection=enabled")
	if err != nil {
		t.Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection=enabled'. Error: %s", test.kn.namespace, err.Error())
	}

	return wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
		out, err := kubectl{test.kn.namespace}.Run("get", "broker", "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Ready\")].status}'")
		if err != nil {
			return false, nil
		} else {
			return strings.Contains(out, "True"), nil
		}
	})
}

func (test *e2eTest) triggerCreate(t *testing.T, r *KnRunResultCollector, name string, sinksvc string, filters []string) {
	args := []string{"trigger", "create", name, "--broker", "default", "--sink", "svc:" + sinksvc}
	for _, v := range filters {
		args = append(args, "--filter", v)
	}
	out := test.kn.Run(args...)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) triggerCreateMissingSink(t *testing.T, r *KnRunResultCollector, name string, sinksvc string) {
	out := test.kn.Run("trigger", "create", name, "--broker", "default", "--sink", "svc:"+sinksvc)
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
}

func (test *e2eTest) triggerDelete(t *testing.T, r *KnRunResultCollector, name string) {
	out := test.kn.Run("trigger", "delete", name)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) triggerUpdate(t *testing.T, r *KnRunResultCollector, name string, filter string, sinksvc string) {
	out := test.kn.Run("trigger", "update", name, "--filter", filter, "--sink", "svc:"+sinksvc)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "updated", "namespace", test.kn.namespace))
}

func (test *e2eTest) verifyTriggerList(t *testing.T, r *KnRunResultCollector, triggers ...string) {
	out := test.kn.Run("trigger", "list")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, triggers...))
}

func (test *e2eTest) verifyTriggerDescribe(t *testing.T, r *KnRunResultCollector, name string, broker string, sink string, filters []string) {
	out := test.kn.Run("trigger", "describe", name)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, filters...))
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, name, broker, sink))
}

func (test *e2eTest) verifyTriggerNotfound(t *testing.T, r *KnRunResultCollector, name string) {
	out := test.kn.Run("trigger", "describe", name)
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, name, "not found"))
}
