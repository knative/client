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
// +build !eventing

package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test/integration"
	"knative.dev/client/pkg/util"
)

func TestRevision(t *testing.T) {
	t.Parallel()
	it, err := integration.NewIntegrationTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := integration.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("create hello service and return no error")
	serviceCreate(t, it, r, "hello")

	t.Log("describe revision from hello service with print flags")
	revName := findRevision(t, it, r, "hello")
	revisionDescribeWithPrintFlags(t, it, r, revName)

	t.Log("update hello service and increase revision count to 2")
	serviceUpdate(t, it, r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("show a list of revisions sorted by the count of configuration generation")
	revisionListWithService(t, it, r, "hello")

	t.Log("update hello service and increase revision count to 3")
	serviceUpdate(t, it, r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("delete three revisions with one revision a nonexistent")
	existRevision1 := findRevisionByGeneration(t, it, r, "hello", 1)
	existRevision2 := findRevisionByGeneration(t, it, r, "hello", 2)
	nonexistRevision := "hello-nonexist"
	revisionMultipleDelete(t, it, r, existRevision1, existRevision2, nonexistRevision)

	t.Log("delete latest revision from hello service and return no error")
	revName = findRevision(t, it, r, "hello")
	revisionDelete(t, it, r, revName)

	t.Log("delete hello service and return no error")
	serviceDelete(t, it, r, "hello")
}

func revisionListWithService(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceNames ...string) {
	for _, svcName := range serviceNames {
		confGen := findConfigurationGeneration(t, it, r, svcName)
		out := it.Kn().Run("revision", "list", "-s", svcName)
		r.AssertNoError(out)

		outputLines := strings.Split(out.Stdout, "\n")
		// Ignore the last line because it is an empty string caused by splitting a line break
		// at the end of the output string
		for _, line := range outputLines[1 : len(outputLines)-1] {
			revName := findRevisionByGeneration(t, it, r, svcName, confGen)
			assert.Check(t, util.ContainsAll(line, revName, svcName, strconv.Itoa(confGen)))
			confGen--
		}
		if t.Failed() {
			r.AddDump("service", svcName, it.Kn().Namespace())
		}
	}
}

func revisionDelete(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, revName string) {
	out := it.Kn().Run("revision", "delete", revName)
	assert.Check(t, util.ContainsAll(out.Stdout, "Revision", revName, "deleted", "namespace", it.Kn().Namespace()))
	r.AssertNoError(out)
}

func revisionMultipleDelete(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, existRevision1, existRevision2, nonexistRevision string) {
	out := it.Kn().Run("revision", "list")
	r.AssertNoError(out)
	assert.Check(t, strings.Contains(out.Stdout, existRevision1), "Required revision1 does not exist")
	assert.Check(t, strings.Contains(out.Stdout, existRevision2), "Required revision2 does not exist")

	out = it.Kn().Run("revision", "delete", existRevision1, existRevision2, nonexistRevision)
	r.AssertNoError(out)

	assert.Check(t, util.ContainsAll(out.Stdout, "Revision", existRevision1, "deleted", "namespace", it.Kn().Namespace()), "Failed to get 'deleted' first revision message")
	assert.Check(t, util.ContainsAll(out.Stdout, "Revision", existRevision2, "deleted", "namespace", it.Kn().Namespace()), "Failed to get 'deleted' second revision message")
	assert.Check(t, util.ContainsAll(out.Stdout, "revisions.serving.knative.dev", nonexistRevision, "not found"), "Failed to get 'not found' error")
}

func revisionDescribeWithPrintFlags(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, revName string) {
	out := it.Kn().Run("revision", "describe", revName, "-o=name")
	r.AssertNoError(out)
	expectedName := fmt.Sprintf("revision.serving.knative.dev/%s", revName)
	assert.Equal(t, strings.TrimSpace(out.Stdout), expectedName)
}

func findRevision(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) string {
	out := it.Kn().Run("revision", "list", "-s", serviceName, "-o=jsonpath={.items[0].metadata.name}")
	r.AssertNoError(out)
	if strings.Contains(out.Stdout, "No resources") {
		t.Errorf("Could not find revision name.")
	}
	return out.Stdout
}

func findRevisionByGeneration(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string, generation int) string {
	maxGen := findConfigurationGeneration(t, it, r, serviceName)
	out := it.Kn().Run("revision", "list", "-s", serviceName,
		fmt.Sprintf("-o=jsonpath={.items[%d].metadata.name}", maxGen-generation))
	r.AssertNoError(out)
	if strings.Contains(out.Stdout, "No resources found.") {
		t.Errorf("Could not find revision name.")
	}
	return out.Stdout
}

func findConfigurationGeneration(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) int {
	out := it.Kn().Run("revision", "list", "-s", serviceName, "-o=jsonpath={.items[0].metadata.labels.serving\\.knative\\.dev/configurationGeneration}")
	r.AssertNoError(out)
	if out.Stdout == "" {
		t.Errorf("Could not find configuration generation.")
	}
	confGen, err := strconv.Atoi(out.Stdout)
	if err != nil {
		t.Errorf("Invalid type of configuration generation: %s", err)
	}

	return confGen
}
