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
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("returns no service before running tests", func(t *testing.T) {
		test.serviceListEmpty(t)
	})

	t.Run("create hello service and returns no error", func(t *testing.T) {
		test.serviceCreate(t, "hello")
	})

	t.Run("create hello service again and get service already exists error", func(t *testing.T) {
		test.serviceCreateDuplicate(t, "hello")
	})

	t.Run("returns valid info about hello service", func(t *testing.T) {
		test.serviceList(t, "hello")
		test.serviceDescribe(t, "hello")
		test.serviceDescribeWithPrintFlags(t, "hello")
	})

	t.Run("update hello service's configuration and returns no error", func(t *testing.T) {
		test.serviceUpdate(t, "hello", []string{"--env", "TARGET=kn", "--port", "8888"})
	})

	t.Run("create another service and returns no error", func(t *testing.T) {
		test.serviceCreate(t, "svc2")
	})

	t.Run("returns a list of revisions associated with hello and svc2 services", func(t *testing.T) {
		test.revisionListForService(t, "hello")
		test.revisionListForService(t, "svc2")
	})

	t.Run("returns a list of routes associated with hello and svc2 services", func(t *testing.T) {
		test.routeList(t)
		test.routeListWithArgument(t, "hello")
		test.routeListWithPrintFlags(t, "hello", "svc2")
	})

	t.Run("describe route from hello service", func(t *testing.T) {
		test.routeDescribe(t, "hello")
		test.routeDescribeWithPrintFlags(t, "hello")
	})

	t.Run("delete hello and svc2 services and returns no error", func(t *testing.T) {
		test.serviceDelete(t, "hello")
		test.serviceDelete(t, "svc2")
	})

	t.Run("returns no service after completing tests", func(t *testing.T) {
		test.serviceListEmpty(t)
	})
}

// Private test functions

func (test *e2eTest) serviceListEmpty(t *testing.T) {
	out, err := test.kn.RunWithOpts([]string{"service", "list"}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "No resources found."))
}

func (test *e2eTest) serviceCreate(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "create",
		fmt.Sprintf("%s", serviceName),
		"--image", KnDefaultTestImage}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully created in namespace", test.kn.namespace, "OK"))
}

func (test *e2eTest) serviceList(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, serviceName))
}

func (test *e2eTest) serviceCreateDuplicate(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName}, runOpts{NoNamespace: false})
	assert.Check(t, strings.Contains(out, serviceName), "The service does not exist yet")

	_, err = test.kn.RunWithOpts([]string{"service", "create",
		fmt.Sprintf("%s", serviceName),
		"--image", KnDefaultTestImage}, runOpts{NoNamespace: false, AllowError: true})

	assert.ErrorContains(t, err, "the service already exists")
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

func (test *e2eTest) serviceDescribe(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutputHeader := `apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:`
	expectedOutput := `generation: 1
  name: %s
  namespace: %s`
	expectedOutput = fmt.Sprintf(expectedOutput, serviceName, test.kn.namespace)
	assert.Check(t, util.ContainsAll(out, expectedOutputHeader, expectedOutput))
}

func (test *e2eTest) serviceUpdate(t *testing.T, serviceName string, args []string) {
	out, err := test.kn.RunWithOpts(append([]string{"service", "update", serviceName}, args...), runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("Service '%s' updated", serviceName)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}

func (test *e2eTest) serviceDescribeWithPrintFlags(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName, "-o=name"}, runOpts{})
	assert.NilError(t, err)

	expectedName := fmt.Sprintf("service.serving.knative.dev/%s", serviceName)
	assert.Check(t, strings.Contains(out, expectedName))
}

func (test *e2eTest) routeList(t *testing.T) {
	out, err := test.kn.RunWithOpts([]string{"route", "list"}, runOpts{})
	assert.NilError(t, err)

	expectedHeaders := []string{"NAME", "URL", "AGE", "CONDITIONS", "TRAFFIC"}
	assert.Check(t, util.ContainsAll(out, expectedHeaders...))
}

func (test *e2eTest) routeListWithArgument(t *testing.T, routeName string) {
	out, err := test.kn.RunWithOpts([]string{"route", "list", routeName}, runOpts{})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("100%% -> %s", routeName)
	assert.Check(t, util.ContainsAll(out, routeName, expectedOutput))
}

func (test *e2eTest) serviceDelete(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "delete", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully deleted in namespace", test.kn.namespace))
}

func (test *e2eTest) routeDescribe(t *testing.T, routeName string) {
	out, err := test.kn.RunWithOpts([]string{"route", "describe", routeName}, runOpts{})
	assert.NilError(t, err)

	expectedGVK := `apiVersion: serving.knative.dev/v1alpha1
kind: Route`
	expectedNamespace := fmt.Sprintf("namespace: %s", test.kn.namespace)
	expectedServiceLabel := fmt.Sprintf("serving.knative.dev/service: %s", routeName)
	assert.Check(t, util.ContainsAll(out, expectedGVK, expectedNamespace, expectedServiceLabel))
}

func (test *e2eTest) routeDescribeWithPrintFlags(t *testing.T, routeName string) {
	out, err := test.kn.RunWithOpts([]string{"route", "describe", routeName, "-o=name"}, runOpts{})
	assert.NilError(t, err)

	expectedName := fmt.Sprintf("route.serving.knative.dev/%s", routeName)
	assert.Check(t, strings.Contains(out, expectedName))
}

func (test *e2eTest) routeListWithPrintFlags(t *testing.T, names ...string) {
	out, err := test.kn.RunWithOpts([]string{"route", "list", "-o=jsonpath={.items[*].metadata.name}"}, runOpts{})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, names...))
}
