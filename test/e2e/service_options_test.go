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

package e2e

import (
	"fmt"
	"testing"

	"github.com/knative/client/pkg/util"
	"gotest.tools/assert"
)

func TestServiceOptions(t *testing.T) {
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("create hello service with concurrency options and returns no error", func(t *testing.T) {
		test.serviceCreateWithOptions(t, "hello", []string{"--concurrency-limit", "250", "--concurrency-target", "300"})
	})

	t.Run("returns valid concurrency options for hello service", func(t *testing.T) {
		test.serviceDescribeConcurrencyLimit(t, "hello", "250")
		test.serviceDescribeConcurrencyTarget(t, "hello", "300")
	})

	t.Run("update concurrency limit for hello service and returns no error", func(t *testing.T) {
		test.serviceUpdate(t, "hello", []string{"--concurrency-limit", "300"})
	})

	t.Run("returns correct concurrency limit for hello service", func(t *testing.T) {
		test.serviceDescribeConcurrencyLimit(t, "hello", "300")
	})

	t.Run("update concurrency options with invalid value for hello service and returns no change", func(t *testing.T) {
		test.serviceUpdate(t, "hello", []string{"--concurrency-limit", "-1", "--concurrency-target", "0"})
	})

	t.Run("returns steady concurrency options for hello service", func(t *testing.T) {
		test.serviceDescribeConcurrencyLimit(t, "hello", "300")
		test.serviceDescribeConcurrencyTarget(t, "hello", "300")
	})

	t.Run("delete hello service and returns no error", func(t *testing.T) {
		test.serviceDelete(t, "hello")
	})
}

// Private

func (test *e2eTest) serviceCreateWithOptions(t *testing.T, serviceName string, options []string) {
	command := []string{"service", "create", serviceName, "--image", KnDefaultTestImage}
	command = append(command, options...)
	out, err := test.kn.RunWithOpts(command, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully created in namespace", test.kn.namespace, "OK"))
}

func (test *e2eTest) serviceDescribeConcurrencyLimit(t *testing.T, serviceName, concurrencyLimit string) {
	out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("containerConcurrency: %s", concurrencyLimit)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}

func (test *e2eTest) serviceDescribeConcurrencyTarget(t *testing.T, serviceName, concurrencyTarget string) {
	out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("autoscaling.knative.dev/target: \"%s\"", concurrencyTarget)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}
