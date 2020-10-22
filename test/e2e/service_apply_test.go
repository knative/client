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
	"testing"

	"gotest.tools/assert"

	pkgtest "knative.dev/pkg/test"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestServiceApply(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("apply hello service (initially)")
	result := serviceApply(r, "hello-apply")
	r.AssertNoError(result)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(result.Stdout, "creating", "service", "hello-apply", "ready", "http"))
	t.Log("apply hello service (unchanged)")
	result = serviceApply(r, "hello-apply")
	r.AssertNoError(result)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(result.Stdout, "no changes", "service", "hello-apply", "http"))

	t.Log("apply hello service (update env)")
	result = serviceApply(r, "hello-apply", "--env", "tik=tok")
	r.AssertNoError(result)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(result.Stdout, "applying", "service", "hello-apply", "ready", "http"))
}

// ServiceApply applies a test service and returns the output
func serviceApply(r *test.KnRunResultCollector, serviceName string, args ...string) test.KnRunResult {
	fullArgs := append([]string{}, "service", "apply", serviceName, "--image", pkgtest.ImagePath("helloworld"))
	fullArgs = append(fullArgs, args...)
	return r.KnTest().Kn().Run(fullArgs...)
}
