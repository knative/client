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
// +build !eventing

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestServiceOptions(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t)

	t.Log("create and validate service with concurrency options")
	defer r.DumpIfFailed()

	serviceCreateWithOptions(t, it, r, "svc1", "--concurrency-limit", "250", "--concurrency-target", "300")
	validateServiceConcurrencyTarget(t, it, r, "svc1", "300")
	validateServiceConcurrencyLimit(t, it, r, "svc1", "250")

	t.Log("update and validate service with concurrency limit")
	serviceUpdate(t, it, r, "svc1", "--concurrency-limit", "300")
	validateServiceConcurrencyLimit(t, it, r, "svc1", "300")

	t.Log("update concurrency options with invalid values for service")
	out := it.Kn().Run("service", "update", "svc1", "--concurrency-limit", "-1", "--concurrency-target", "0")
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "invalid"))

	t.Log("returns steady concurrency options for service")
	validateServiceConcurrencyLimit(t, it, r, "svc1", "300")
	validateServiceConcurrencyTarget(t, it, r, "svc1", "300")

	t.Log("delete service")
	serviceDelete(t, it, r, "svc1")

	t.Log("create and validate service with min/max scale options ")
	serviceCreateWithOptions(t, it, r, "svc2", "--min-scale", "1", "--max-scale", "3")
	validateServiceMinScale(t, it, r, "svc2", "1")
	validateServiceMaxScale(t, it, r, "svc2", "3")

	t.Log("update and validate service with max scale option")
	serviceUpdate(t, it, r, "svc2", "--max-scale", "2")
	validateServiceMaxScale(t, it, r, "svc2", "2")

	t.Log("delete service")
	serviceDelete(t, it, r, "svc2")

	t.Log("create, update and validate service with annotations")
	serviceCreateWithOptions(t, it, r, "svc3", "--annotation", "alpha=wolf", "--annotation", "brave=horse")
	validateServiceAnnotations(t, it, r, "svc3", map[string]string{"alpha": "wolf", "brave": "horse"})
	serviceUpdate(t, it, r, "svc3", "--annotation", "alpha=direwolf", "--annotation", "brave-")
	validateServiceAnnotations(t, it, r, "svc3", map[string]string{"alpha": "direwolf", "brave": ""})
	serviceDelete(t, it, r, "svc3")

	t.Log("create, update and validate service with autoscale window option")
	serviceCreateWithOptions(t, it, r, "svc4", "--autoscale-window", "1m")
	validateAutoscaleWindow(t, it, r, "svc4", "1m")
	serviceUpdate(t, it, r, "svc4", "--autoscale-window", "15s")
	validateAutoscaleWindow(t, it, r, "svc4", "15s")
	serviceDelete(t, it, r, "svc4")

	t.Log("create, update and validate service with cmd and arg options")
	serviceCreateWithOptions(t, it, r, "svc5", "--cmd", "/go/bin/helloworld")
	validateContainerField(t, it, r, "svc5", "command", "[/go/bin/helloworld]")
	serviceUpdate(t, it, r, "svc5", "--arg", "myArg1", "--arg", "--myArg2")
	validateContainerField(t, it, r, "svc5", "args", "[myArg1 --myArg2]")
	serviceUpdate(t, it, r, "svc5", "--arg", "myArg1")
	validateContainerField(t, it, r, "svc5", "args", "[myArg1]")

	t.Log("create, update and validate service with user defined")
	var uid int64 = 1000
	if uids, ok := os.LookupEnv("TEST_RUN_AS_UID"); ok {
		uid, err = strconv.ParseInt(uids, 10, 64)
		assert.NilError(t, err)
	}
	serviceCreateWithOptions(t, it, r, "svc6", "--user", strconv.FormatInt(uid, 10))
	validateUserID(t, it, r, "svc6", uid)
	serviceUpdate(t, it, r, "svc6", "--user", strconv.FormatInt(uid+1, 10))
	validateUserID(t, it, r, "svc6", uid+1)
}

func serviceCreateWithOptions(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName string, options ...string) {
	command := []string{"service", "create", serviceName, "--image", test.KnDefaultTestImage}
	command = append(command, options...)
	out := it.Kn().Run(command...)
	assert.Check(t, util.ContainsAll(out.Stdout, "service", serviceName, "Creating", "namespace", it.Kn().Namespace(), "Ready"))
	r.AssertNoError(out)
}

func validateServiceConcurrencyLimit(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName, concurrencyLimit string) {
	jsonpath := "jsonpath={.items[0].spec.template.spec.containerConcurrency}"
	out := it.Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, concurrencyLimit)
	r.AssertNoError(out)
}

func validateServiceConcurrencyTarget(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName, concurrencyTarget string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/target}"
	out := it.Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, concurrencyTarget)
	r.AssertNoError(out)
}

func validateAutoscaleWindow(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName, window string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/window}"
	out := it.Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, window)
	r.AssertNoError(out)
}

func validateServiceMinScale(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName, minScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/minScale}"
	out := it.Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, minScale)
	r.AssertNoError(out)
}

func validateServiceMaxScale(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName, maxScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/maxScale}"
	out := it.Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, maxScale)
	r.AssertNoError(out)
}

func validateServiceAnnotations(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName string, annotations map[string]string) {
	metadataAnnotationsJsonpathFormat := "jsonpath={.metadata.annotations.%s}"
	templateAnnotationsJsonpathFormat := "jsonpath={.spec.template.metadata.annotations.%s}"

	for k, v := range annotations {
		out := it.Kn().Run("service", "describe", serviceName, "-o", fmt.Sprintf(metadataAnnotationsJsonpathFormat, k))
		assert.Equal(t, v, out.Stdout)
		r.AssertNoError(out)

		out = it.Kn().Run("service", "describe", serviceName, "-o", fmt.Sprintf(templateAnnotationsJsonpathFormat, k))
		assert.Equal(t, v, out.Stdout)
		r.AssertNoError(out)
	}
}

func validateContainerField(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName, field, expected string) {
	jsonpath := fmt.Sprintf("jsonpath={.items[0].spec.template.spec.containers[0].%s}", field)
	out := it.Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, expected)
	r.AssertNoError(out)
}

func validateUserID(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, serviceName string, uid int64) {
	out := it.Kn().Run("service", "describe", serviceName, "-ojson")
	data := json.NewDecoder(strings.NewReader(out.Stdout))
	data.UseNumber()
	var service servingv1.Service
	err := data.Decode(&service)
	assert.NilError(t, err)
	assert.Equal(t, *service.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser, uid)
}
