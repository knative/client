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

	"knative.dev/client/lib/test/integration"
	"knative.dev/client/pkg/util"
)

func TestBrokerTrigger(t *testing.T) {
	t.Parallel()
	it, err := integration.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := integration.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	err = lableNamespaceForDefaultBroker(t, it)
	assert.NilError(t, err)
	defer unlableNamespaceForDefaultBroker(t, it)

	serviceCreate(t, it, r, "sinksvc0")
	serviceCreate(t, it, r, "sinksvc1")

	t.Log("create triggers and list them")
	triggerCreate(t, it, r, "trigger1", "sinksvc0", []string{"a=b"})
	triggerCreate(t, it, r, "trigger2", "sinksvc1", []string{"type=knative.dev.bar", "source=ping"})
	verifyTriggerList(t, it, r, "trigger1", "trigger2")
	triggerDelete(t, it, r, "trigger1")
	triggerDelete(t, it, r, "trigger2")

	t.Log("create a trigger and delete it")
	triggerCreate(t, it, r, "deltrigger", "sinksvc0", []string{"a=b"})
	triggerDelete(t, it, r, "deltrigger")
	verifyTriggerNotfound(t, it, r, "deltrigger")

	t.Log("create a trigger with filters and remove them one by one")
	triggerCreate(t, it, r, "filtertrigger", "sinksvc0", []string{"foo=bar", "source=ping"})
	verifyTriggerDescribe(t, it, r, "filtertrigger", "default", "sinksvc0", []string{"foo", "bar", "source", "ping"})
	triggerUpdate(t, it, r, "filtertrigger", "foo-", "sinksvc0")
	verifyTriggerDescribe(t, it, r, "filtertrigger", "default", "sinksvc0", []string{"source", "ping"})
	triggerUpdate(t, it, r, "filtertrigger", "source-", "sinksvc0")
	verifyTriggerDescribe(t, it, r, "filtertrigger", "default", "sinksvc0", nil)
	triggerDelete(t, it, r, "filtertrigger")

	t.Log("create a trigger, describe and update it")
	triggerCreate(t, it, r, "updtrigger", "sinksvc0", []string{"a=b"})
	verifyTriggerDescribe(t, it, r, "updtrigger", "default", "sinksvc0", []string{"a", "b"})
	triggerUpdate(t, it, r, "updtrigger", "type=knative.dev.bar", "sinksvc1")
	verifyTriggerDescribe(t, it, r, "updtrigger", "default", "sinksvc1", []string{"a", "b", "type", "knative.dev.bar"})
	triggerDelete(t, it, r, "updtrigger")

	t.Log("create trigger with error return")
	triggerCreateMissingSink(t, it, r, "errtrigger", "notfound")
}

// Private functions

func unlableNamespaceForDefaultBroker(t *testing.T, it *integration.KnTest) {
	_, err := integration.Kubectl{}.Run("label", "namespace", it.Kn().Namespace(), "knative-eventing-injection-")
	if err != nil {
		t.Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection-'. Error: %s", it.Kn().Namespace(), err.Error())
	}
}

func lableNamespaceForDefaultBroker(t *testing.T, it *integration.KnTest) error {
	_, err := integration.Kubectl{}.Run("label", "namespace", it.Kn().Namespace(), "knative-eventing-injection=enabled")

	if err != nil {
		t.Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection=enabled'. Error: %s", it.Kn().Namespace(), err.Error())
	}

	return wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
		out, err := integration.NewKubectl(it.Kn().Namespace()).Run("get", "broker", "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Ready\")].status}'")
		if err != nil {
			return false, nil
		} else {
			return strings.Contains(out, "True"), nil
		}
	})
}

func triggerCreate(t *testing.T, it *integration.KnTest, r *integration.KnRunResultCollector, name string, sinksvc string, filters []string) {
	args := []string{"trigger", "create", name, "--broker", "default", "--sink", "svc:" + sinksvc}
	if len(filters) > 0 {
		for _, v := range filters {
			args = append(args, "--filter", v)
		}
	}
	out := it.Kn().Run(args...)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "created", "namespace", it.Kn().Namespace()))
}

func triggerCreateMissingSink(t *testing.T, it *integration.KnTest, r *integration.KnRunResultCollector, name string, sinksvc string) {
	out := it.Kn().Run("trigger", "create", name, "--broker", "default", "--sink", "svc:"+sinksvc)
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
}

func triggerDelete(t *testing.T, it *integration.KnTest, r *integration.KnRunResultCollector, name string) {
	out := it.Kn().Run("trigger", "delete", name)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "deleted", "namespace", it.Kn().Namespace()))
}

func triggerUpdate(t *testing.T, it *integration.KnTest, r *integration.KnRunResultCollector, name string, filter string, sinksvc string) {
	out := it.Kn().Run("trigger", "update", name, "--filter", filter, "--sink", "svc:"+sinksvc)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Trigger", name, "updated", "namespace", it.Kn().Namespace()))
}

func verifyTriggerList(t *testing.T, it *integration.KnTest, r *integration.KnRunResultCollector, triggers ...string) {
	out := it.Kn().Run("trigger", "list")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, triggers...))
}

func verifyTriggerDescribe(t *testing.T, it *integration.KnTest, r *integration.KnRunResultCollector, name string, broker string, sink string, filters []string) {
	out := it.Kn().Run("trigger", "describe", name)
	r.AssertNoError(out)
	if len(filters) > 0 {
		assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, filters...))
	} else {
		assert.Check(t, util.ContainsNone(out.Stdout, "Filter"))
	}
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, name, broker, sink))
}

func verifyTriggerNotfound(t *testing.T, it *integration.KnTest, r *integration.KnRunResultCollector, name string) {
	out := it.Kn().Run("trigger", "describe", name)
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, name, "not found"))
}
