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
	teardown := Setup(t)
	defer teardown(t)

	t.Run("create hello service with concurrency options and returns no error", func(t *testing.T) {
		testServiceCreateWithOptions(t, k, "hello", []string{"--concurrency-limit", "250", "--concurrency-target", "300"})
	})

	t.Run("returns valid concurrency options for hello service", func(t *testing.T) {
		testServiceDescribeConcurrencyLimit(t, k, "hello", "250")
		testServiceDescribeConcurrencyTarget(t, k, "hello", "300")
	})

	t.Run("update concurrency limit for hello service and returns no error", func(t *testing.T) {
		testServiceUpdate(t, k, "hello", []string{"--concurrency-limit", "300"})
	})

	t.Run("returns correct concurrency limit for hello service", func(t *testing.T) {
		testServiceDescribeConcurrencyLimit(t, k, "hello", "300")
	})

	t.Run("delete hello service and returns no error", func(t *testing.T) {
		testServiceDelete(t, k, "hello")
	})
}

func testServiceCreateWithOptions(t *testing.T, k kn, serviceName string, options []string) {
	command := []string{"service", "create", serviceName, "--image", KnDefaultTestImage}
	command = append(command, options...)
	out, err := k.RunWithOpts(command, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Service", serviceName, "successfully created in namespace", k.namespace, "OK"))
}

func testServiceDescribeConcurrencyLimit(t *testing.T, k kn, serviceName, concurrencyLimit string) {
	out, err := k.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("containerConcurrency: %s", concurrencyLimit)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}

func testServiceDescribeConcurrencyTarget(t *testing.T, k kn, serviceName, concurrencyTarget string) {
	out, err := k.RunWithOpts([]string{"service", "describe", serviceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)

	expectedOutput := fmt.Sprintf("autoscaling.knative.dev/target: \"%s\"", concurrencyTarget)
	assert.Check(t, util.ContainsAll(out, expectedOutput))
}
