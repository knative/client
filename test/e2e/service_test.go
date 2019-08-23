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

	"gotest.tools/assert"
)

func TestService(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("create hello service duplicate and get service already exists error", func(t *testing.T) {
		test.serviceCreate(t, "hello")
		test.serviceCreateDuplicate(t, "hello")
	})

	t.Run("return valid info about hello service with print flags", func(t *testing.T) {
		test.serviceDescribeWithPrintFlags(t, "hello")
	})

	t.Run("delete hello service repeatedly and get an error", func(t *testing.T) {
		test.serviceDelete(t, "hello")
		test.serviceDeleteNonexistent(t, "hello")
	})
}

func (test *e2eTest) serviceCreateDuplicate(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, strings.Contains(out, serviceName), "The service does not exist yet")

	_, err = test.kn.RunWithOpts([]string{"service", "create", serviceName,
		"--image", KnDefaultTestImage}, runOpts{NoNamespace: false, AllowError: true})

	assert.ErrorContains(t, err, "the service already exists")
}

func (test *e2eTest) serviceDescribeWithPrintFlags(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName, "-o=name"}, runOpts{})
	assert.NilError(t, err)

	expectedName := fmt.Sprintf("service.serving.knative.dev/%s", serviceName)
	assert.Equal(t, strings.TrimSpace(out), expectedName)
}

func (test *e2eTest) serviceDeleteNonexistent(t *testing.T, serviceName string) {
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, !strings.Contains(out, serviceName), "The service exists")

	_, err = test.kn.RunWithOpts([]string{"service", "delete", serviceName}, runOpts{NoNamespace: false, AllowError: true})

	expectedErr := fmt.Sprintf(`services.serving.knative.dev "%s" not found`, serviceName)
	assert.ErrorContains(t, err, expectedErr)
}
