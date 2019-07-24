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

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/knative/client/pkg/util"
	"gotest.tools/assert"
)

func TestBasicWorkflow(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("returns no service before running tests", func(t *testing.T) {
		test.serviceListEmpty(t)
	})

	t.Run("create hello service and return no error", func(t *testing.T) {
		test.serviceCreate(t, "hello")
	})

	t.Run("return valid info about hello service", func(t *testing.T) {
		test.serviceList(t, "hello")
		test.serviceDescribe(t, "hello")
	})

	t.Run("update hello service's configuration and return no error", func(t *testing.T) {
		test.serviceUpdate(t, "hello", []string{"--env", "TARGET=kn", "--port", "8888"})
	})

	t.Run("create another service and return no error", func(t *testing.T) {
		test.serviceCreate(t, "svc2")
	})

	t.Run("return a list of revisions associated with hello and svc2 services", func(t *testing.T) {
		test.revisionListForService(t, "hello")
		test.revisionListForService(t, "svc2")
	})

	t.Run("describe revision from hello service", func(t *testing.T) {
		test.revisionDescribe(t, "hello")
	})

	t.Run("delete hello and svc2 services and return no error", func(t *testing.T) {
		test.serviceDelete(t, "hello")
		test.serviceDelete(t, "svc2")
	})

	t.Run("return no service after completing tests", func(t *testing.T) {
		test.serviceListEmpty(t)
	})
}

func (test *e2eTest) serviceListEmpty(t *testing.T) {
	out, err := test.kn.RunWithOpts([]string{"service", "list"}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "No resources found."))
}

func (test *e2eTest) serviceCreate(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "create", serviceName,
		"--image", KnDefaultTestImage}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully created in namespace", test.kn.namespace, "OK"))
}

func (test *e2eTest) serviceList(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, serviceName))
}

func (test *e2eTest) serviceDescribe(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Assert(t, util.ContainsAll(out, serviceName, test.kn.namespace, KnDefaultTestImage))
	assert.Assert(t, util.ContainsAll(out, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
	assert.Assert(t, util.ContainsAll(out, "Name", "Namespace", "URL", "Address", "Annotations", "Age", "Revisions"))
}

func (test *e2eTest) serviceUpdate(t *testing.T, serviceName string, args []string) {
	out, err := test.kn.RunWithOpts(append([]string{"service", "update", serviceName}, args...), runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("Service '%s' updated", serviceName)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}

func (test *e2eTest) serviceDelete(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "delete", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully deleted in namespace", test.kn.namespace))
}

func (test *e2eTest) revisionListForService(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"revision", "list", "-s", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	outputLines := strings.Split(out, "\n")
	// Ignore the last line because it is an empty string caused by splitting a line break
	// at the end of the output string
	for _, line := range outputLines[1 : len(outputLines)-1] {
		// The last item is the revision status, which should be ready
		assert.Check(t, util.ContainsAll(line, " "+serviceName+" ", "True"))
	}
}

func (test *e2eTest) revisionDescribe(t *testing.T, serviceName string) {
	revName := test.findRevision(t, serviceName)

	out, err := test.kn.RunWithOpts([]string{"revision", "describe", revName}, runOpts{})
	assert.NilError(t, err)

	expectedGVK := `apiVersion: serving.knative.dev/v1alpha1
kind: Revision`
	expectedNamespace := fmt.Sprintf("namespace: %s", test.kn.namespace)
	expectedServiceLabel := fmt.Sprintf("serving.knative.dev/service: %s", serviceName)
	assert.Check(t, util.ContainsAll(out, expectedGVK, expectedNamespace, expectedServiceLabel))
}
