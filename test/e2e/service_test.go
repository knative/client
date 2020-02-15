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
	"fmt"
	"strings"
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/pkg/util"
)

func TestService(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("create hello service duplicate and get service already exists error")
	test.serviceCreate(t, r, "hello")
	test.serviceCreateDuplicate(t, r, "hello")

	t.Log("return valid info about hello service with print flags")
	test.serviceDescribeWithPrintFlags(t, r, "hello")

	t.Log("delete hello service repeatedly and get an error")
	test.serviceDelete(t, r, "hello")
	test.serviceDeleteNonexistent(t, r, "hello")

	t.Log("delete two services with a service nonexistent")
	test.serviceCreate(t, r, "hello")

	test.serviceMultipleDelete(t, r, "hello", "bla123")
}

func (test *e2eTest) serviceCreateDuplicate(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "list", serviceName)
	out.ErrorExpected = true
	r.AssertNoError(out)
	assert.Check(t, strings.Contains(out.Stdout, serviceName), "The service does not exist yet")

	out = test.kn.Run("service", "create", serviceName, "--image", KnDefaultTestImage)
	out.ErrorExpected = true
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "the service already exists"))
}

func (test *e2eTest) serviceDescribeWithPrintFlags(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "describe", serviceName, "-o=name")
	r.AssertNoError(out)

	expectedName := fmt.Sprintf("service.serving.knative.dev/%s", serviceName)
	assert.Equal(t, strings.TrimSpace(out.Stdout), expectedName)
}

func (test *e2eTest) serviceDeleteNonexistent(t *testing.T, r *KnRunResultCollector, serviceName string) {
	out := test.kn.Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(t, !strings.Contains(out.Stdout, serviceName), "The service exists")

	out = test.kn.Run("service", "delete", serviceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, "hello", "not found"), "Failed to get 'not found' error")
}

func (test *e2eTest) serviceMultipleDelete(t *testing.T, r *KnRunResultCollector, existService, nonexistService string) {
	out := test.kn.Run("service", "list")
	r.AssertNoError(out)
	assert.Check(t, strings.Contains(out.Stdout, existService), "The service ", existService, " does not exist (but is expected to exist)")
	assert.Check(t, !strings.Contains(out.Stdout, nonexistService), "The service", nonexistService, " exists (but is supposed to be not)")

	out = test.kn.Run("service", "delete", existService, nonexistService)
	r.AssertNoError(out)

	expectedSuccess := fmt.Sprintf(`Service '%s' successfully deleted in namespace '%s'.`, existService, test.kn.namespace)
	expectedErr := fmt.Sprintf(`services.serving.knative.dev "%s" not found`, nonexistService)
	assert.Check(t, strings.Contains(out.Stdout, expectedSuccess), "Failed to get 'successfully deleted' message")
	assert.Check(t, strings.Contains(out.Stdout, expectedErr), "Failed to get 'not found' error")
}
