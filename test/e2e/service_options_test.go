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
	"knative.dev/client/pkg/util"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestServiceOptions(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)

	t.Log("create and validate service with concurrency options")
	defer r.DumpIfFailed()

	test.serviceCreateWithOptions(t, r, "svc1", "--concurrency-limit", "250", "--concurrency-target", "300")
	test.validateServiceConcurrencyTarget(t, r, "svc1", "300")
	test.validateServiceConcurrencyLimit(t, r, "svc1", "250")

	t.Log("update and validate service with concurrency limit")
	test.serviceUpdate(t, r, "svc1", "--concurrency-limit", "300")
	test.validateServiceConcurrencyLimit(t, r, "svc1", "300")

	t.Log("update concurrency options with invalid values for service")
	out := test.kn.Run("service", "update", "svc1", "--concurrency-limit", "-1", "--concurrency-target", "0")
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "invalid"))

	t.Log("returns steady concurrency options for service")
	test.validateServiceConcurrencyLimit(t, r, "svc1", "300")
	test.validateServiceConcurrencyTarget(t, r, "svc1", "300")

	t.Log("delete service")
	test.serviceDelete(t, r, "svc1")

	t.Log("create and validate service with min/max scale options ")
	test.serviceCreateWithOptions(t, r, "svc2", "--min-scale", "1", "--max-scale", "3")
	test.validateServiceMinScale(t, r, "svc2", "1")
	test.validateServiceMaxScale(t, r, "svc2", "3")

	t.Log("update and validate service with max scale option")
	test.serviceUpdate(t, r, "svc2", "--max-scale", "2")
	test.validateServiceMaxScale(t, r, "svc2", "2")

	t.Log("delete service")
	test.serviceDelete(t, r, "svc2")

	t.Log("create, update and validate service with annotations")
	test.serviceCreateWithOptions(t, r, "svc3", "--annotation", "alpha=wolf", "--annotation", "brave=horse")
	test.validateServiceAnnotations(t, r, "svc3", map[string]string{"alpha": "wolf", "brave": "horse"})
	test.serviceUpdate(t, r, "svc3", "--annotation", "alpha=direwolf", "--annotation", "brave-")
	test.validateServiceAnnotations(t, r, "svc3", map[string]string{"alpha": "direwolf", "brave": ""})
	test.serviceDelete(t, r, "svc3")

	t.Log("create, update and validate service with autoscale window option")
	test.serviceCreateWithOptions(t, r, "svc4", "--autoscale-window", "1m")
	test.validateAutoscaleWindow(t, r, "svc4", "1m")
	test.serviceUpdate(t, r, "svc4", "--autoscale-window", "15s")
	test.validateAutoscaleWindow(t, r, "svc4", "15s")
	test.serviceDelete(t, r, "svc4")

	t.Log("create, update and validate service with cmd and arg options")
	test.serviceCreateWithOptions(t, r, "svc5", "--cmd", "/go/bin/helloworld")
	test.validateContainerField(t, r, "svc5", "command", "[/go/bin/helloworld]")
	test.serviceUpdate(t, r, "svc5", "--arg", "myArg1", "--arg", "--myArg2")
	test.validateContainerField(t, r, "svc5", "args", "[myArg1 --myArg2]")
	test.serviceUpdate(t, r, "svc5", "--arg", "myArg1")
	test.validateContainerField(t, r, "svc5", "args", "[myArg1]")

	t.Log("create, update and validate service with user defined")
	var uid int64 = 1000
	if uids, ok := os.LookupEnv("TEST_RUN_AS_UID"); ok {
		uid, err = strconv.ParseInt(uids, 10, 64)
		assert.NilError(t, err)
	}
	test.serviceCreateWithOptions(t, r, "svc6", "--user", strconv.FormatInt(uid, 10))
	test.validateUserId(t, r, "svc6", uid)
	test.serviceUpdate(t, r, "svc6", "--user", strconv.FormatInt(uid+1, 10))
	test.validateUserId(t, r, "svc6", uid+1)
}

func (test *e2eTest) serviceCreateWithOptions(t *testing.T, r *KnRunResultCollector, serviceName string, options ...string) {
	command := []string{"service", "create", serviceName, "--image", KnDefaultTestImage}
	command = append(command, options...)
	out := test.kn.Run(command...)
	assert.Check(t, util.ContainsAll(out.Stdout, "service", serviceName, "Creating", "namespace", test.kn.namespace, "Ready"))
	r.AssertNoError(out)
}

func (test *e2eTest) validateServiceConcurrencyLimit(t *testing.T, r *KnRunResultCollector, serviceName, concurrencyLimit string) {
	jsonpath := "jsonpath={.items[0].spec.template.spec.containerConcurrency}"
	out := test.kn.Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, concurrencyLimit)
	r.AssertNoError(out)
}

func (test *e2eTest) validateServiceConcurrencyTarget(t *testing.T, r *KnRunResultCollector, serviceName, concurrencyTarget string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/target}"
	out := test.kn.Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, concurrencyTarget)
	r.AssertNoError(out)
}

func (test *e2eTest) validateAutoscaleWindow(t *testing.T, r *KnRunResultCollector, serviceName, window string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/window}"
	out := test.kn.Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, window)
	r.AssertNoError(out)
}

func (test *e2eTest) validateServiceMinScale(t *testing.T, r *KnRunResultCollector, serviceName, minScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/minScale}"
	out := test.kn.Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, minScale)
	r.AssertNoError(out)
}

func (test *e2eTest) validateServiceMaxScale(t *testing.T, r *KnRunResultCollector, serviceName, maxScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/maxScale}"
	out := test.kn.Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, maxScale)
	r.AssertNoError(out)
}

func (test *e2eTest) validateServiceAnnotations(t *testing.T, r *KnRunResultCollector, serviceName string, annotations map[string]string) {
	metadataAnnotationsJsonpathFormat := "jsonpath={.metadata.annotations.%s}"
	templateAnnotationsJsonpathFormat := "jsonpath={.spec.template.metadata.annotations.%s}"

	for k, v := range annotations {
		out := test.kn.Run("service", "describe", serviceName, "-o", fmt.Sprintf(metadataAnnotationsJsonpathFormat, k))
		assert.Equal(t, v, out.Stdout)
		r.AssertNoError(out)

		out = test.kn.Run("service", "describe", serviceName, "-o", fmt.Sprintf(templateAnnotationsJsonpathFormat, k))
		assert.Equal(t, v, out.Stdout)
		r.AssertNoError(out)
	}
}

func (test *e2eTest) validateContainerField(t *testing.T, r *KnRunResultCollector, serviceName, field, expected string) {
	jsonpath := fmt.Sprintf("jsonpath={.items[0].spec.template.spec.containers[0].%s}", field)
	out := test.kn.Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(t, out.Stdout, expected)
	r.AssertNoError(out)
}

func (test *e2eTest) validateUserId(t *testing.T, r *KnRunResultCollector, serviceName string, uid int64) {
	out := test.kn.Run("service", "describe", serviceName, "-ojson")
	data := json.NewDecoder(strings.NewReader(out.Stdout))
	data.UseNumber()
	var service servingv1.Service
	err := data.Decode(&service)
	assert.NilError(t, err)
	assert.Equal(t, *service.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser, uid)
}
