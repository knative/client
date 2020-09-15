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

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
	pkgtest "knative.dev/pkg/test"
	"knative.dev/serving/pkg/apis/serving"
)

func TestService(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("create hello service, delete, and try to create duplicate and get service already exists error")
	test.ServiceCreate(r, "hello")
	serviceCreatePrivate(r, "hello-private")
	serviceCreateDuplicate(r, "hello-private")

	t.Log("return valid info about hello service with print flags")
	serviceDescribeWithPrintFlags(r, "hello")

	t.Log("delete hello service repeatedly and get an error")
	test.ServiceDelete(r, "hello")
	serviceDeleteNonexistent(r, "hello")

	t.Log("delete two services with a service nonexistent")
	test.ServiceCreate(r, "hello")
	serviceMultipleDelete(r, "hello", "bla123")

	t.Log("create service private and make public")
	serviceCreatePrivateUpdatePublic(r, "hello-private-public")

	t.Log("error message from --untag with tag that doesn't exist")
	test.ServiceCreate(r, "untag")
	serviceUntagTagThatDoesNotExist(r, "untag")

	t.Log("delete all services in a namespace")
	test.ServiceCreate(r, "svc1")
	test.ServiceCreate(r, "service2")
	test.ServiceCreate(r, "ksvc3")
	serviceDeleteAll(r)
}

func serviceCreatePrivate(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName,
		"--image", pkgtest.ImagePath("helloworld"), "--cluster-local")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))

	out = r.KnTest().Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, serving.VisibilityLabelKeyObsolete, serving.VisibilityClusterLocal))
}

func serviceCreatePrivateUpdatePublic(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName,
		"--image", pkgtest.ImagePath("helloworld"), "--cluster-local")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))

	out = r.KnTest().Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, serving.VisibilityLabelKeyObsolete, serving.VisibilityClusterLocal))

	out = r.KnTest().Kn().Run("service", "update", serviceName,
		"--image", pkgtest.ImagePath("helloworld"), "--no-cluster-local")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "updated", "namespace", r.KnTest().Kn().Namespace(), "ready"))

	out = r.KnTest().Kn().Run("service", "describe", serviceName, "--verbose")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsNone(out.Stdout, serving.VisibilityLabelKeyObsolete, serving.VisibilityClusterLocal))
}

func serviceCreateDuplicate(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, serviceName), "The service does not exist yet")

	out = r.KnTest().Kn().Run("service", "create", serviceName, "--image", pkgtest.ImagePath("helloworld"))
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "the service already exists"))
}

func serviceDescribeWithPrintFlags(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o=name")
	r.AssertNoError(out)

	expectedName := fmt.Sprintf("service.serving.knative.dev/%s", serviceName)
	assert.Equal(r.T(), strings.TrimSpace(out.Stdout), expectedName)
}

func serviceDeleteNonexistent(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), !strings.Contains(out.Stdout, serviceName), "The service exists")

	out = r.KnTest().Kn().Run("service", "delete", serviceName)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "hello", "not found"), "Failed to get 'not found' error")
}

func serviceMultipleDelete(r *test.KnRunResultCollector, existService, nonexistService string) {
	out := r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, existService), "The service ", existService, " does not exist (but is expected to exist)")
	assert.Check(r.T(), !strings.Contains(out.Stdout, nonexistService), "The service", nonexistService, " exists (but is supposed to be not)")

	out = r.KnTest().Kn().Run("service", "delete", existService, nonexistService)
	r.AssertError(out)

	expectedSuccess := fmt.Sprintf(`Service '%s' successfully deleted in namespace '%s'.`, existService, r.KnTest().Kn().Namespace())
	expectedErr := fmt.Sprintf(`services.serving.knative.dev "%s" not found`, nonexistService)
	assert.Check(r.T(), strings.Contains(out.Stdout, expectedSuccess), "Failed to get 'successfully deleted' message")
	assert.Check(r.T(), strings.Contains(out.Stderr, expectedErr), "Failed to get 'not found' error")
}

func serviceUntagTagThatDoesNotExist(r *test.KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), strings.Contains(out.Stdout, serviceName), "Service "+serviceName+" does not exist for test (but should exist)")

	out = r.KnTest().Kn().Run("service", "update", serviceName, "--untag", "foo", "--no-wait")
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "tag(s)", "foo", "not present", "service", "untag"), "Expected error message for using --untag with nonexistent tag")
}

func serviceDeleteAll(r *test.KnRunResultCollector) {
	out := r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	// Check if services created successfully/available for test.
	assert.Check(r.T(), !strings.Contains(out.Stdout, "No services found."), "No services created for kn service delete --all e2e (but should exist)")

	out = r.KnTest().Kn().Run("service", "delete", "--all")
	r.AssertNoError(out)
	// Check if output contains successfully deleted to verify deletion took place.
	assert.Check(r.T(), strings.Contains(out.Stdout, "successfully deleted"), "Failed to get 'successfully deleted' message")

	out = r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	// Check if no services present after kn service delete --all.
	assert.Check(r.T(), strings.Contains(out.Stdout, "No services found."), "Failed to show 'No services found' after kn service delete --all")
}
