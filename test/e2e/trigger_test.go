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
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	err := test.lableNamespaceForDefaultBroker(t)
	assert.NilError(t, err)
	test.serviceCreate(t, "sinksvc0")
	test.serviceCreate(t, "sinksvc1")

	t.Run("create triggers and list them", func(t *testing.T) {
		test.triggerCreate(t, "trigger1", []string{"a=b"}, "sinksvc0")
		test.triggerCreate(t, "trigger2", []string{"type=knative.dev.bar", "source=cronjob"}, "sinksvc1")
		test.verifyTriggerList(t, []string{"trigger1", "trigger2"})
	})

	t.Run("create a trigger and delete it", func(t *testing.T) {
		test.triggerCreate(t, "deltrigger", []string{"a=b"}, "sinksvc0")
		test.triggerDelete(t, "deltrigger")
		test.verifyTriggerNotfound(t, "deltrigger")
	})

	t.Run("create a trigger, describe and update it", func(t *testing.T) {
		test.triggerCreate(t, "updtrigger", []string{"a=b"}, "sinksvc0")
		test.verifyTriggerDescribe(t, "updtrigger", []string{"a", "b"}, "default", "sinksvc0")
		test.triggerUpdate(t, "updtrigger", "type=knative.dev.bar", "sinksvc1")
		test.verifyTriggerDescribe(t, "updtrigger", []string{"a", "b", "type", "knative.dev.bar"}, "default", "sinksvc1")
	})

	t.Run("create trigger with error return", func(t *testing.T) {
		test.triggerCreateMissingSink(t, "errtrigger", "notfound")
	})
}

func (test *e2eTest) lableNamespaceForDefaultBroker(t *testing.T) error {
	kubectl := kubectl{t, Logger{}}

	_, err := kubectl.RunWithOpts([]string{"label", "namespace", test.kn.namespace, "knative-eventing-injection=enabled"}, runOpts{})
	if err != nil {
		t.Fatalf("Error executing 'kubectl label namespace %s knative-eventing-injection=enabled'. Error: %s", test.kn.namespace, err.Error())
	}

	return wait.PollImmediate(10*time.Second, 5*time.Minute, func() (bool, error) {
		out, err := kubectl.RunWithOpts([]string{"get", "broker", "-n", test.kn.namespace, "-o=jsonpath='{.items[0].status.conditions[?(@.type==\"Ready\")].status}'"}, runOpts{AllowError: true})
		if err != nil {
			return false, nil
		} else {
			return strings.Contains(out, "True"), nil
		}
	})
}

func (test *e2eTest) triggerCreate(t *testing.T, name string, filters []string, sinksvc string) {
	args := []string{"trigger", "create", name, "--broker", "default", "--sink", "svc:" + sinksvc}
	for _, v := range filters {
		args = append(args, "--filter", v)
	}
	out, err := test.kn.RunWithOpts(args, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "Trigger", name, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) triggerCreateMissingSink(t *testing.T, name string, sinksvc string) {
	_, err := test.kn.RunWithOpts([]string{"trigger", "create", name, "--broker", "default", "--sink", "svc:" + sinksvc}, runOpts{NoNamespace: false, AllowError: true})
	assert.ErrorContains(t, err, "services.serving.knative.dev", "not found")
}

func (test *e2eTest) triggerDelete(t *testing.T, name string) {
	out, err := test.kn.RunWithOpts([]string{"trigger", "delete", name}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "Trigger", name, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) triggerUpdate(t *testing.T, name string, filter string, sinksvc string) {
	out, err := test.kn.RunWithOpts([]string{"trigger", "update", name, "--filter", filter, "--sink", "svc:" + sinksvc}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "Trigger", name, "updated", "namespace", test.kn.namespace))
}

func (test *e2eTest) verifyTriggerList(t *testing.T, triggers []string) {
	out, err := test.kn.RunWithOpts([]string{"trigger", "list"}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, triggers...))
}

func (test *e2eTest) verifyTriggerDescribe(t *testing.T, name string, filters []string, broker string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"trigger", "describe", name}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, filters...))
	assert.Check(t, util.ContainsAllIgnoreCase(out, name, broker, sink))
}

func (test *e2eTest) verifyTriggerNotfound(t *testing.T, name string) {
	_, err := test.kn.RunWithOpts([]string{"trigger", "describe", name}, runOpts{NoNamespace: false, AllowError: true})
	assert.ErrorContains(t, err, name, "not found")
}
