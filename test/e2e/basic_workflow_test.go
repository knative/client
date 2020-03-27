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

	"knative.dev/client/lib/test/integration"
	"knative.dev/client/pkg/util"
)

func TestBasicWorkflow(t *testing.T) {
	t.Parallel()
	it, err := integration.NewIntegrationTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := integration.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("returns no service before running tests")
	serviceListEmpty(t, it, r)

	t.Log("create hello service and return no error")
	serviceCreate(t, it, r, "hello")

	t.Log("return valid info about hello service")
	serviceList(t, it, r, "hello")
	serviceDescribe(t, it, r, "hello")

	t.Log("update hello service's configuration and return no error")
	serviceUpdate(t, it, r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("create another service and return no error")
	serviceCreate(t, it, r, "svc2")

	t.Log("return a list of revisions associated with hello and svc2 services")
	revisionListForService(t, it, r, "hello")
	revisionListForService(t, it, r, "svc2")

	t.Log("describe revision from hello service")
	revisionDescribe(t, it, r, "hello")

	t.Log("delete hello and svc2 services and return no error")
	serviceDelete(t, it, r, "hello")
	serviceDelete(t, it, r, "svc2")

	t.Log("return no service after completing tests")
	serviceListEmpty(t, it, r)
}

func TestWrongCommand(t *testing.T) {
	r := integration.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	out := integration.Kn{}.Run("source", "apiserver", "noverb", "--tag=0.13")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown subcommand", "noverb"))
	r.AssertError(out)

	out = integration.Kn{}.Run("rev")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown command", "rev"))
	r.AssertError(out)

}

// ==========================================================================

func serviceListEmpty(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector) {
	out := it.Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, "No services found."))
}

func serviceCreate(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "create", serviceName, "--image", integration.KnDefaultTestImage)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", it.Kn().Namespace(), "ready"))
}

func serviceList(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, serviceName))
}

func serviceDescribe(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(t, util.ContainsAll(out.Stdout, serviceName, it.Kn().Namespace(), integration.KnDefaultTestImage))
	assert.Assert(t, util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
	assert.Assert(t, util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Age", "Revisions"))
}

func serviceUpdate(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := it.Kn().Run(fullArgs...)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "updating", "service", serviceName, "ready"))
}

func serviceDelete(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "delete", serviceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, "Service", serviceName, "successfully deleted in namespace", it.Kn().Namespace()))
}

func revisionListForService(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("revision", "list", "-s", serviceName)
	r.AssertNoError(out)
	outputLines := strings.Split(out.Stdout, "\n")
	// Ignore the last line because it is an empty string caused by splitting a line break
	// at the end of the output string
	for _, line := range outputLines[1 : len(outputLines)-1] {
		// The last item is the revision status, which should be ready
		assert.Check(t, util.ContainsAll(line, " "+serviceName+" ", "True"))
	}
}

func revisionDescribe(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	revName := findRevision(t, it, r, serviceName)

	out := it.Kn().Run("revision", "describe", revName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, revName, it.Kn().Namespace(), serviceName, "++ Ready", "TARGET=kn"))
}
