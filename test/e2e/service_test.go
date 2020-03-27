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

	"knative.dev/client/lib/test/integration"
	"knative.dev/client/pkg/util"
	"knative.dev/serving/pkg/apis/serving"
)

func TestService(t *testing.T) {
	t.Parallel()
	it, err := integration.NewIntegrationTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := integration.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("create hello service, delete, and try to create duplicate and get service already exists error")
	serviceCreate(t, it, r, "hello")
	serviceCreatePrivate(t, it, r, "hello-private")
	serviceCreateDuplicate(t, it, r, "hello-private")

	t.Log("return valid info about hello service with print flags")
	serviceDescribeWithPrintFlags(t, it, r, "hello")

	t.Log("delete hello service repeatedly and get an error")
	serviceDelete(t, it, r, "hello")
	serviceDeleteNonexistent(t, it, r, "hello")

	t.Log("delete two services with a service nonexistent")
	serviceCreate(t, it, r, "hello")
	serviceMultipleDelete(t, it, r, "hello", "bla123")

	t.Log("create service private and make public")
	serviceCreatePrivateUpdatePublic(t, it, r, "hello-private-public")
}

func serviceCreatePrivate(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "create", serviceName,
		"--image", integration.KnDefaultTestImage, "--cluster-local")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", it.Kn().Namespace(), "ready"))

	out = it.Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, serving.VisibilityLabelKey, serving.VisibilityClusterLocal))
}

func serviceCreatePrivateUpdatePublic(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "create", serviceName,
		"--image", integration.KnDefaultTestImage, "--cluster-local")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", it.Kn().Namespace(), "ready"))

	out = it.Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, serving.VisibilityLabelKey, serving.VisibilityClusterLocal))

	out = it.Kn().Run("service", "update", serviceName,
		"--image", integration.KnDefaultTestImage, "--no-cluster-local")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "updated", "namespace", it.Kn().Namespace(), "ready"))

	out = it.Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsNone(out.Stdout, serving.VisibilityLabelKey, serving.VisibilityClusterLocal))
}

func serviceCreateDuplicate(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(t, strings.Contains(out.Stdout, serviceName), "The service does not exist yet")

	out = it.Kn().Run("service", "create", serviceName, "--image", integration.KnDefaultTestImage)
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "the service already exists"))
}

func serviceDescribeWithPrintFlags(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "describe", serviceName, "-o=name")
	r.AssertNoError(out)

	expectedName := fmt.Sprintf("service.serving.knative.dev/%s", serviceName)
	assert.Equal(t, strings.TrimSpace(out.Stdout), expectedName)
}

func serviceDeleteNonexistent(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, serviceName string) {
	out := it.Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(t, !strings.Contains(out.Stdout, serviceName), "The service exists")

	out = it.Kn().Run("service", "delete", serviceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, "hello", "not found"), "Failed to get 'not found' error")
}

func serviceMultipleDelete(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, existService, nonexistService string) {
	out := it.Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(t, strings.Contains(out.Stdout, existService), "The service ", existService, " does not exist (but is expected to exist)")
	assert.Check(t, !strings.Contains(out.Stdout, nonexistService), "The service", nonexistService, " exists (but is supposed to be not)")

	out = it.Kn().Run("service", "delete", existService, nonexistService)
	r.AssertNoError(out)

	expectedSuccess := fmt.Sprintf(`Service '%s' successfully deleted in namespace '%s'.`, existService, it.Kn().Namespace())
	expectedErr := fmt.Sprintf(`services.serving.knative.dev "%s" not found`, nonexistService)
	assert.Check(t, strings.Contains(out.Stdout, expectedSuccess), "Failed to get 'successfully deleted' message")
	assert.Check(t, strings.Contains(out.Stdout, expectedErr), "Failed to get 'not found' error")
}
