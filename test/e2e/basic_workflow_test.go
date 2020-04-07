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

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestBasicWorkflow(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("returns no service before running tests")
	serviceListEmpty(r)

	t.Log("create hello service and return no error")
	serviceCreate(r, "hello")

	t.Log("return valid info about hello service")
	serviceList(r, "hello")
	serviceDescribe(r, "hello")

	t.Log("return list --output name about hello service")
	serviceListOutput(r, "hello")

	t.Log("update hello service's configuration and return no error")
	serviceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("create another service and return no error")
	serviceCreate(r, "svc2")

	t.Log("return a list of revisions associated with hello and svc2 services")
	revisionListForService(r, "hello")
	revisionListForService(r, "svc2")

	t.Log("describe revision from hello service")
	revisionDescribe(r, "hello")

	t.Log("delete hello and svc2 services and return no error")
	serviceDelete(r, "hello")
	serviceDelete(r, "svc2")

	t.Log("return no service after completing tests")
	serviceListEmpty(r)
}

func TestWrongCommand(t *testing.T) {
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	out := test.Kn{}.Run("source", "apiserver", "noverb", "--tag=0.13")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown subcommand", "noverb"))
	r.AssertError(out)

	out = test.Kn{}.Run("rev")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown command", "rev"))
	r.AssertError(out)

}

// ==========================================================================

func serviceListEmpty(r *test.KnRunResultCollector) {
	out := r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "No services found."))
}

func serviceCreate(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName, "--image", test.KnDefaultTestImage)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))
}

func serviceList(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName))
}

func serviceDescribe(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, serviceName, r.KnTest().Kn().Namespace(), test.KnDefaultTestImage))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Age", "Revisions"))
}

func serviceListOutput(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName, "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName, "service.serving.knative.dev"))
}

func serviceUpdate(r *test.KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := r.KnTest().Kn().Run(fullArgs...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "updating", "service", serviceName, "ready"))
}

func serviceDelete(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "delete", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Service", serviceName, "successfully deleted in namespace", r.KnTest().Kn().Namespace()))
}

func revisionListForService(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("revision", "list", "-s", serviceName)
	r.AssertNoError(out)
	outputLines := strings.Split(out.Stdout, "\n")
	// Ignore the last line because it is an empty string caused by splitting a line break
	// at the end of the output string
	for _, line := range outputLines[1 : len(outputLines)-1] {
		// The last item is the revision status, which should be ready
		assert.Check(r.T(), util.ContainsAll(line, " "+serviceName+" ", "True"))
	}
}

func revisionDescribe(r *test.KnRunResultCollector, serviceName string) {
	revName := findRevision(r, serviceName)

	out := r.KnTest().Kn().Run("revision", "describe", revName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, revName, r.KnTest().Kn().Namespace(), serviceName, "++ Ready", "TARGET=kn"))
}
