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
	"testing"

	"github.com/knative/client/pkg/util"
	"gotest.tools/assert"
)

var (
	e env
	k kn
)

const (
	KnDefaultTestImage string = "gcr.io/knative-samples/helloworld-go"
)

func Setup(t *testing.T) func(t *testing.T) {
	e = buildEnv(t)
	k = kn{t, e.Namespace, Logger{}}
	CreateTestNamespace(t, e.Namespace)
	return Teardown
}

func Teardown(t *testing.T) {
	DeleteTestNamespace(t, e.Namespace)
}

func TestBasicWorkflow(t *testing.T) {
	teardown := Setup(t)
	defer teardown(t)

	testServiceListEmpty(t, k)
	testServiceCreate(t, k, "hello")
	testServiceList(t, k, "hello")
	testServiceDescribe(t, k, "hello")
	testServiceUpdate(t, k, "hello", []string{"--env", "TARGET=kn", "--port", "8888"})
	testServiceCreate(t, k, "svc2")
	testRevisionListForService(t, k, "hello")
	testRevisionListForService(t, k, "svc2")
	testRouteList(t, k)
	testRouteListWithArgument(t, k, "hello")
	testServiceDelete(t, k, "hello")
	testServiceDelete(t, k, "svc2")
	testServiceListEmpty(t, k)
}

// Private test functions

func testServiceListEmpty(t *testing.T, k kn) {
	out, err := k.RunWithOpts([]string{"service", "list"}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "No resources found."))
}

func testServiceCreate(t *testing.T, k kn, serviceName string) {
	out, err := k.RunWithOpts([]string{"service", "create",
		fmt.Sprintf("%s", serviceName),
		"--image", KnDefaultTestImage}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully created in namespace", k.namespace, "OK"))
}

func testServiceList(t *testing.T, k kn, serviceName string) {
	out, err := k.RunWithOpts([]string{"service", "list", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("%s", serviceName)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}

func testRevisionListForService(t *testing.T, k kn, serviceName string) {
	out, err := k.RunWithOpts([]string{"revision", "list", "-s", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, serviceName, "True"))
}

func testServiceDescribe(t *testing.T, k kn, serviceName string) {
	out, err := k.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutputHeader := `apiVersion: serving.knative.dev/v1alpha1
kind: Service
metadata:`
	expectedOutput := `generation: 1
  name: %s
  namespace: %s`
	expectedOutput = fmt.Sprintf(expectedOutput, serviceName, k.namespace)
	assert.Check(t, util.ContainsAll(out, expectedOutputHeader, expectedOutput))
}

func testServiceUpdate(t *testing.T, k kn, serviceName string, args []string) {
	out, err := k.RunWithOpts(append([]string{"service", "update", serviceName}, args...), runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("Service '%s' updated", serviceName)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}

func testRouteList(t *testing.T, k kn) {
	out, err := k.RunWithOpts([]string{"route", "list"}, runOpts{})
	assert.NilError(t, err)

	expectedHeaders := []string{"NAME", "URL", "AGE", "CONDITIONS", "TRAFFIC"}
	assert.Check(t, util.ContainsAll(out, expectedHeaders...))
}

func testRouteListWithArgument(t *testing.T, k kn, routeName string) {
	out, err := k.RunWithOpts([]string{"route", "list", routeName}, runOpts{})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("100%% -> %s", routeName)
	assert.Check(t, util.ContainsAll(out, routeName, expectedOutput))
}

func testServiceDelete(t *testing.T, k kn, serviceName string) {
	out, err := k.RunWithOpts([]string{"service", "delete", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully deleted in namespace", k.namespace))
}
