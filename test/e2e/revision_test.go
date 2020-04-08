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

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestRevision(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("create hello service and return no error")
	serviceCreate(r, "hello")

	t.Log("describe revision from hello service with print flags")
	revName := findRevision(r, "hello")
	revisionDescribeWithPrintFlags(r, revName)

	t.Log("update hello service and increase revision count to 2")
	serviceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("show a list of revisions sorted by the count of configuration generation")
	revisionListWithService(r, "hello")

	t.Log("update hello service and increase revision count to 3")
	serviceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("delete three revisions with one revision a nonexistent")
	existRevision1 := findRevisionByGeneration(r, "hello", 1)
	existRevision2 := findRevisionByGeneration(r, "hello", 2)
	nonexistRevision := "hello-nonexist"
	revisionMultipleDelete(r, existRevision1, existRevision2, nonexistRevision)

	t.Log("delete latest revision from hello service and return no error")
	revName = findRevision(r, "hello")
	revisionDelete(r, revName)

	t.Log("delete hello service and return no error")
	serviceDelete(r, "hello")
}

func revisionListWithService(r *test.KnRunResultCollector, serviceNames ...string) {
	for _, svcName := range serviceNames {
		confGen := findConfigurationGeneration(r, svcName)
		out := r.KnTest().Kn().Run("revision", "list", "-s", svcName)
		r.AssertNoError(out)

		outputLines := strings.Split(out.Stdout, "\n")
		// Ignore the last line because it is an empty string caused by splitting a line break
		// at the end of the output string
		for _, line := range outputLines[1 : len(outputLines)-1] {
			revName := findRevisionByGeneration(r, svcName, confGen)
			assert.Check(r.T(), util.ContainsAll(line, revName, svcName, strconv.Itoa(confGen)))
			confGen--
		}
		if r.T().Failed() {
			r.AddDump("service", svcName, r.KnTest().Kn().Namespace())
		}
	}
}

func revisionDelete(r *test.KnRunResultCollector, revName string) {
	out := r.KnTest().Kn().Run("revision", "delete", revName)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Revision", revName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}

func revisionMultipleDelete(r *test.KnRunResultCollector, existRevision1, existRevision2, nonexistRevision string) {
	out := r.KnTest().Kn().Run("revision", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, existRevision1), "Required revision1 does not exist")
	assert.Check(r.T(), strings.Contains(out.Stdout, existRevision2), "Required revision2 does not exist")

	out = r.KnTest().Kn().Run("revision", "delete", existRevision1, existRevision2, nonexistRevision)
	r.AssertNoError(out)

	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Revision", existRevision1, "deleted", "namespace", r.KnTest().Kn().Namespace()), "Failed to get 'deleted' first revision message")
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Revision", existRevision2, "deleted", "namespace", r.KnTest().Kn().Namespace()), "Failed to get 'deleted' second revision message")
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "revisions.serving.knative.dev", nonexistRevision, "not found"), "Failed to get 'not found' error")
}

func revisionDescribeWithPrintFlags(r *test.KnRunResultCollector, revName string) {
	out := r.KnTest().Kn().Run("revision", "describe", revName, "-o=name")
	r.AssertNoError(out)
	expectedName := fmt.Sprintf("revision.serving.knative.dev/%s", revName)
	assert.Equal(r.T(), strings.TrimSpace(out.Stdout), expectedName)
}

func findRevision(r *test.KnRunResultCollector, serviceName string) string {
	out := r.KnTest().Kn().Run("revision", "list", "-s", serviceName, "-o=jsonpath={.items[0].metadata.name}")
	r.AssertNoError(out)
	if strings.Contains(out.Stdout, "No resources") {
		r.T().Errorf("Could not find revision name.")
	}
	return out.Stdout
}

func findRevisionByGeneration(r *test.KnRunResultCollector, serviceName string, generation int) string {
	maxGen := findConfigurationGeneration(r, serviceName)
	out := r.KnTest().Kn().Run("revision", "list", "-s", serviceName,
		fmt.Sprintf("-o=jsonpath={.items[%d].metadata.name}", maxGen-generation))
	r.AssertNoError(out)
	if strings.Contains(out.Stdout, "No resources found.") {
		r.T().Errorf("Could not find revision name.")
	}
	return out.Stdout
}

func findConfigurationGeneration(r *test.KnRunResultCollector, serviceName string) int {
	out := r.KnTest().Kn().Run("revision", "list", "-s", serviceName, "-o=jsonpath={.items[0].metadata.labels.serving\\.knative\\.dev/configurationGeneration}")
	r.AssertNoError(out)
	if out.Stdout == "" {
		r.T().Errorf("Could not find configuration generation.")
	}
	confGen, err := strconv.Atoi(out.Stdout)
	if err != nil {
		r.T().Errorf("Invalid type of configuration generation: %s", err)
	}

	return confGen
}
