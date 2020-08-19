// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"strconv"
	"strings"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

// RevisionListForService list revisions of given service and verifies if their status is True
func RevisionListForService(r *KnRunResultCollector, serviceName string) {
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

// RevisionDescribe verifies revision describe output for given service's revision
func RevisionDescribe(r *KnRunResultCollector, serviceName string) {
	revName := FindRevision(r, serviceName)

	out := r.KnTest().Kn().Run("revision", "describe", revName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, revName, r.KnTest().Kn().Namespace(), serviceName, "++ Ready", "TARGET=kn"))
}

// RevisionDelete verifies deleting given revision in sync mode
func RevisionDelete(r *KnRunResultCollector, revName string) {
	out := r.KnTest().Kn().Run("revision", "delete", "--wait", revName)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Revision", revName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
	r.AssertNoError(out)
}

// RevisionMultipleDelete verifies deleting multiple revisions
func RevisionMultipleDelete(r *KnRunResultCollector, existRevision1, existRevision2, nonexistRevision string) {
	out := r.KnTest().Kn().Run("revision", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, existRevision1), "Required revision1 does not exist")
	assert.Check(r.T(), strings.Contains(out.Stdout, existRevision2), "Required revision2 does not exist")

	out = r.KnTest().Kn().Run("revision", "delete", existRevision1, existRevision2, nonexistRevision)
	r.AssertError(out)

	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Revision", existRevision1, "deleted", "namespace", r.KnTest().Kn().Namespace()), "Failed to get 'deleted' first revision message")
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Revision", existRevision2, "deleted", "namespace", r.KnTest().Kn().Namespace()), "Failed to get 'deleted' second revision message")
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "revisions.serving.knative.dev", nonexistRevision, "not found"), "Failed to get 'not found' error")
}

// RevisionDescribeWithPrintFlags verifies describing given revision using print flag '--output=name'
func RevisionDescribeWithPrintFlags(r *KnRunResultCollector, revName string) {
	out := r.KnTest().Kn().Run("revision", "describe", revName, "-o=name")
	r.AssertNoError(out)
	expectedName := fmt.Sprintf("revision.serving.knative.dev/%s", revName)
	assert.Equal(r.T(), strings.TrimSpace(out.Stdout), expectedName)
}

// FindRevision returns a revision name (at index 0) for given service
func FindRevision(r *KnRunResultCollector, serviceName string) string {
	out := r.KnTest().Kn().Run("revision", "list", "-s", serviceName, "-o=jsonpath={.items[0].metadata.name}")
	r.AssertNoError(out)
	if strings.Contains(out.Stdout, "No resources") {
		r.T().Errorf("Could not find revision name.")
	}
	return out.Stdout
}

// FindRevisionByGeneration returns a revision name for given revision at given generation number
func FindRevisionByGeneration(r *KnRunResultCollector, serviceName string, generation int) string {
	maxGen := FindConfigurationGeneration(r, serviceName)
	out := r.KnTest().Kn().Run("revision", "list", "-s", serviceName,
		fmt.Sprintf("-o=jsonpath={.items[%d].metadata.name}", maxGen-generation))
	r.AssertNoError(out)
	if strings.Contains(out.Stdout, "No resources found.") {
		r.T().Errorf("Could not find revision name.")
	}
	return out.Stdout
}

// FindConfigurationGeneration returns the configuration generation number of given service
func FindConfigurationGeneration(r *KnRunResultCollector, serviceName string) int {
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

// RevisionListOutputName verifies listing given revision using print flag '--output name'
func RevisionListOutputName(r *KnRunResultCollector, revisionName string) {
	out := r.KnTest().Kn().Run("revision", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, revisionName, "revision.serving.knative.dev"))
}

// RevisionListWithService verifies listing revisions per service from each given service names
func RevisionListWithService(r *KnRunResultCollector, serviceNames ...string) {
	for _, svcName := range serviceNames {
		confGen := FindConfigurationGeneration(r, svcName)
		out := r.KnTest().Kn().Run("revision", "list", "-s", svcName)
		r.AssertNoError(out)

		outputLines := strings.Split(out.Stdout, "\n")
		// Ignore the last line because it is an empty string caused by splitting a line break
		// at the end of the output string
		for _, line := range outputLines[1 : len(outputLines)-1] {
			revName := FindRevisionByGeneration(r, svcName, confGen)
			assert.Check(r.T(), util.ContainsAll(line, revName, svcName, strconv.Itoa(confGen)))
			confGen--
		}
		if r.T().Failed() {
			r.AddDump("service", svcName, r.KnTest().Kn().Namespace())
		}
	}
}
