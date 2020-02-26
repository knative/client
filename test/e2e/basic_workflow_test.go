// Copyright 2019 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e
// +build !eventing

package e2e

import (
	"strings"
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/pkg/util"
)

func TestBasicWorkflow(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("returns no service before running tests")
	test.serviceListEmpty(t, r)

	t.Log("create hello service and return no error")
	test.serviceCreate(t, r, "hello")

	t.Log("return valid info about hello service")
	test.serviceList(t, r, "hello")
	test.serviceDescribe(t, r, "hello")

	t.Log("update hello service's configuration and return no error")
	test.serviceUpdate(t, r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("create another service and return no error")
	test.serviceCreate(t, r, "svc2")

	t.Log("return a list of revisions associated with hello and svc2 services")
	test.revisionListForService(t, r, "hello")
	test.revisionListForService(t, r, "svc2")

	t.Log("describe revision from hello service")
	test.revisionDescribe(t, r, "hello")

	t.Log("delete hello and svc2 services and return no error")
	test.serviceDelete(t, r, "hello")
	test.serviceDelete(t, r, "svc2")

	t.Log("return no service after completing tests")
	test.serviceListEmpty(t, r)
}

func TestWrongCommand(t *testing.T) {
	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	out := kn{}.Run("source", "apiserver", "noverb", "--tag=0.13")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown subcommand", "noverb"))
	r.AssertError(out)

	out = kn{}.Run("rev")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown command", "rev"))
	r.AssertError(out)

}

// ==========================================================================

func (test *e2eTest) serviceListEmpty(t *testing.T, r *KnRunResultCollector) {
	out := test.kn.Run("service", "list")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, "No services found."))
}

func (test *e2eTest) serviceCreate(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "create", serviceName, "--image", KnDefaultTestImage)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", test.kn.namespace, "ready"))
}

func (test *e2eTest) serviceList(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, serviceName))
}

func (test *e2eTest) serviceDescribe(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(t, util.ContainsAll(out.Stdout, serviceName, test.kn.namespace, KnDefaultTestImage))
	assert.Assert(t, util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
	assert.Assert(t, util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Age", "Revisions"))
}

func (test *e2eTest) serviceUpdate(t *testing.T, r *KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := test.kn.Run(fullArgs...)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "updating", "service", serviceName, "ready"))
}

func (test *e2eTest) serviceDelete(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "delete", serviceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, "Service", serviceName, "successfully deleted in namespace", test.kn.namespace))
}

func (test *e2eTest) revisionListForService(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("revision", "list", "-s", serviceName)
	r.AssertNoError(out)
	outputLines := strings.Split(out.Stdout, "\n")
	// Ignore the last line because it is an empty string caused by splitting a line break
	// at the end of the output string
	for _, line := range outputLines[1 : len(outputLines)-1] {
		// The last item is the revision status, which should be ready
		assert.Check(t, util.ContainsAll(line, " "+serviceName+" ", "True"))
	}
}

func (test *e2eTest) revisionDescribe(t *testing.T, r *KnRunResultCollector, serviceName string) {
	revName := test.findRevision(t, r, serviceName)

	out := test.kn.Run("revision", "describe", revName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, revName, test.kn.namespace, serviceName, "++ Ready", "TARGET=kn"))
}
