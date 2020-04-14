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

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestBrokerTrigger(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	err = lableNamespaceForDefaultBroker(t, it)
	assert.NilError(t, err)
	defer unlableNamespaceForDefaultBroker(t, it)

	serviceCreate(r, "sinksvc0")
	serviceCreate(r, "sinksvc1")

	t.Log("create triggers and list them")
	triggerCreate(r, "trigger1", "sinksvc0", []string{"a=b"})
	triggerCreate(r, "trigger2", "sinksvc1", []string{"type=knative.dev.bar", "source=ping"})
	verifyTriggerList(r, "trigger1", "trigger2")
	verifyTriggerListOutputName(r, "trigger1", "trigger2")
	triggerDelete(r, "trigger1")
	triggerDelete(r, "trigger2")

	t.Log("create a trigger and delete it")
	triggerCreate(r, "deltrigger", "sinksvc0", []string{"a=b"})
	triggerDelete(r, "deltrigger")
	verifyTriggerNotfound(r, "deltrigger")

	t.Log("create a trigger with filters and remove them one by one")
	triggerCreate(r, "filtertrigger", "sinksvc0", []string{"foo=bar", "source=ping"})
	verifyTriggerDescribe(r, "filtertrigger", "default", "sinksvc0", []string{"foo", "bar", "source", "ping"})
	triggerUpdate(r, "filtertrigger", "foo-", "sinksvc0")
	verifyTriggerDescribe(r, "filtertrigger", "default", "sinksvc0", []string{"source", "ping"})
	triggerUpdate(r, "filtertrigger", "source-", "sinksvc0")
	verifyTriggerDescribe(r, "filtertrigger", "default", "sinksvc0", nil)
	triggerDelete(r, "filtertrigger")

	t.Log("create a trigger, describe and update it")
	triggerCreate(r, "updtrigger", "sinksvc0", []string{"a=b"})
	verifyTriggerDescribe(r, "updtrigger", "default", "sinksvc0", []string{"a", "b"})
	triggerUpdate(r, "updtrigger", "type=knative.dev.bar", "sinksvc1")
	verifyTriggerDescribe(r, "updtrigger", "default", "sinksvc1", []string{"a", "b", "type", "knative.dev.bar"})
	triggerDelete(r, "updtrigger")

	t.Log("create trigger with error return")
	triggerCreateMissingSink(r, "errtrigger", "notfound")
}

// Private functions

func unlableNamespaceForDefaultBroker(t *testing.T, it *test.KnTest) {
	_, err := test.Kubectl{}.Run("label", "namespace", it.Kn().Namespace(), "knative-eventing-injection-")
	if err != nil {
		t.Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection-'. Error: %s", it.Kn().Namespace(), err.Error())
	}
}

func lableNamespaceForDefaultBroker(t *testing.T, it *test.KnTest) error {
	_, err := test.Kubectl{}.Run("label", "namespace", it.Kn().Namespace(), "knative-eventing-injection=enabled")

	if err != nil {
		t.Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection=enabled'. Error: %s", it.Kn().Namespace(), err.Error())
	}

	return wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
		out, err := test.NewKubectl(it.Kn().Namespace()).Run("get", "broker", "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Ready\")].status}'")
		if err != nil {
			return false, nil
		} else {
			return strings.Contains(out, "True"), nil
		}
	})
}

func triggerCreate(r *test.KnRunResultCollector, name string, sinksvc string, filters []string) {
	args := []string{"trigger", "create", name, "--broker", "default", "--sink", "svc:" + sinksvc}
	if len(filters) > 0 {
		for _, v := range filters {
			args = append(args, "--filter", v)
		}
	}
	out := r.KnTest().Kn().Run(args...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func triggerCreateMissingSink(r *test.KnRunResultCollector, name string, sinksvc string) {
	out := r.KnTest().Kn().Run("trigger", "create", name, "--broker", "default", "--sink", "svc:"+sinksvc)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
}

func triggerDelete(r *test.KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("trigger", "delete", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

func triggerUpdate(r *test.KnRunResultCollector, name string, filter string, sinksvc string) {
	out := r.KnTest().Kn().Run("trigger", "update", name, "--filter", filter, "--sink", "svc:"+sinksvc)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "updated", "namespace", r.KnTest().Kn().Namespace()))
}

func verifyTriggerList(r *test.KnRunResultCollector, triggers ...string) {
	out := r.KnTest().Kn().Run("trigger", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, triggers...))
}

func verifyTriggerListOutputName(r *test.KnRunResultCollector, triggers ...string) {
	out := r.KnTest().Kn().Run("trigger", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, triggers...))
}

func verifyTriggerDescribe(r *test.KnRunResultCollector, name string, broker string, sink string, filters []string) {
	out := r.KnTest().Kn().Run("trigger", "describe", name)
	r.AssertNoError(out)
	if len(filters) > 0 {
		assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, filters...))
	} else {
		assert.Check(r.T(), util.ContainsNone(out.Stdout, "Filter"))
	}
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, name, broker, sink))
}

func verifyTriggerNotfound(r *test.KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("trigger", "describe", name)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, name, "not found"))
}
