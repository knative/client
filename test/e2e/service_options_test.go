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
	"fmt"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

func TestServiceOptions(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("create and validate service with concurrency options", func(t *testing.T) {
		test.serviceCreateWithOptions(t, "svc1", []string{"--concurrency-limit", "250", "--concurrency-target", "300"})
		test.validateServiceConcurrencyTarget(t, "svc1", "300")
		test.validateServiceConcurrencyLimit(t, "svc1", "250")
	})

	t.Run("update and validate service with concurrency limit", func(t *testing.T) {
		test.serviceUpdate(t, "svc1", []string{"--concurrency-limit", "300"})
		test.validateServiceConcurrencyLimit(t, "svc1", "300")
	})

	t.Run("update concurrency options with invalid values for service", func(t *testing.T) {
		command := []string{"service", "update", "svc1", "--concurrency-limit", "-1", "--concurrency-target", "0"}
		_, err := test.kn.RunWithOpts(command, runOpts{NoNamespace: false, AllowError: true})
		assert.ErrorContains(t, err, "invalid")
	})

	t.Run("returns steady concurrency options for service", func(t *testing.T) {
		test.validateServiceConcurrencyLimit(t, "svc1", "300")
		test.validateServiceConcurrencyTarget(t, "svc1", "300")
	})

	t.Run("delete service", func(t *testing.T) {
		test.serviceDelete(t, "svc1")
	})

	t.Run("create and validate service with min/max scale options ", func(t *testing.T) {
		test.serviceCreateWithOptions(t, "svc2", []string{"--min-scale", "1", "--max-scale", "3"})
		test.validateServiceMinScale(t, "svc2", "1")
		test.validateServiceMaxScale(t, "svc2", "3")
	})

	t.Run("update and validate service with max scale option", func(t *testing.T) {
		test.serviceUpdate(t, "svc2", []string{"--max-scale", "2"})
		test.validateServiceMaxScale(t, "svc2", "2")
	})

	t.Run("delete service", func(t *testing.T) {
		test.serviceDelete(t, "svc2")
	})

	t.Run("create, update and validate service with annotations", func(t *testing.T) {
		test.serviceCreateWithOptions(t, "svc3", []string{"--annotation", "alpha=wolf", "--annotation", "brave=horse"})
		test.validateServiceAnnotations(t, "svc3", map[string]string{"alpha": "wolf", "brave": "horse"})
		test.serviceUpdate(t, "svc3", []string{"--annotation", "alpha=direwolf", "--annotation", "brave-"})
		test.validateServiceAnnotations(t, "svc3", map[string]string{"alpha": "direwolf", "brave": ""})
		test.serviceDelete(t, "svc3")
	})

	t.Run("create, update and validate service with autoscale window option", func(t *testing.T) {
		test.serviceCreateWithOptions(t, "svc4", []string{"--autoscale-window", "1m"})
		test.validateAutoscaleWindow(t, "svc4", "1m")
		test.serviceUpdate(t, "svc4", []string{"--autoscale-window", "15s"})
		test.validateAutoscaleWindow(t, "svc4", "15s")
		test.serviceDelete(t, "svc4")
	})

	t.Run("create, update and validate service with cmd and arg options", func(t *testing.T) {
		test.serviceCreateWithOptions(t, "svc5", []string{"--cmd", "/go/bin/helloworld"})
		test.validateContainerField(t, "svc5", "command", "[/go/bin/helloworld]")
		test.serviceUpdate(t, "svc5", []string{"--arg", "myArg1", "--arg", "--myArg2"})
		test.validateContainerField(t, "svc5", "args", "[myArg1 --myArg2]")
		test.serviceUpdate(t, "svc5", []string{"--arg", "myArg1"})
		test.validateContainerField(t, "svc5", "args", "[myArg1]")
	})
}

func (test *e2eTest) serviceCreateWithOptions(t *testing.T, serviceName string, options []string) {
	command := []string{"service", "create", serviceName, "--image", KnDefaultTestImage}
	command = append(command, options...)
	out, err := test.kn.RunWithOpts(command, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, "service", serviceName, "Creating", "namespace", test.kn.namespace, "Ready"))
}

func (test *e2eTest) validateServiceConcurrencyLimit(t *testing.T, serviceName, concurrencyLimit string) {
	jsonpath := "jsonpath={.items[0].spec.template.spec.containerConcurrency}"
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
	assert.NilError(t, err)
	if out != "" {
		assert.Equal(t, out, concurrencyLimit)
	} else {
		// case where server returns fields like spec.runLatest.configuration.revisionTemplate.spec.containerConcurrency
		// TODO: Remove this case when `runLatest` field is deprecated altogether / v1beta1
		jsonpath = "jsonpath={.items[0].spec.runLatest.configuration.revisionTemplate.spec.containerConcurrency}"
		out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
		assert.NilError(t, err)
		assert.Equal(t, out, concurrencyLimit)
	}
}

func (test *e2eTest) validateServiceConcurrencyTarget(t *testing.T, serviceName, concurrencyTarget string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/target}"
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
	assert.NilError(t, err)
	if out != "" {
		assert.Equal(t, out, concurrencyTarget)
	} else {
		// case where server returns fields like  spec.runLatest.configuration.revisionTemplate.spec.containerConcurrency
		// TODO: Remove this case when `runLatest` field is deprecated altogether / v1beta1
		jsonpath = "jsonpath={.items[0].spec.runLatest.configuration.revisionTemplate.metadata.annotations.autoscaling\\.knative\\.dev/target}"
		out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
		assert.NilError(t, err)
		assert.Equal(t, out, concurrencyTarget)
	}
}

func (test *e2eTest) validateAutoscaleWindow(t *testing.T, serviceName, window string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/window}"
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
	assert.NilError(t, err)
	assert.Equal(t, out, window)
}

func (test *e2eTest) validateServiceMinScale(t *testing.T, serviceName, minScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/minScale}"
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
	assert.NilError(t, err)
	if out != "" {
		assert.Equal(t, minScale, out)
	} else {
		// case where server could return either old or new fields
		// #TODO: remove this when old fields are deprecated, v1beta1
		jsonpath = "jsonpath={.items[0].spec.runLatest.configuration.revisionTemplate.metadata.annotations.autoscaling\\.knative\\.dev/minScale}"
		out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
		assert.NilError(t, err)
		assert.Equal(t, minScale, out)
	}
}

func (test *e2eTest) validateServiceMaxScale(t *testing.T, serviceName, maxScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/maxScale}"
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
	assert.NilError(t, err)
	if out != "" {
		assert.Equal(t, maxScale, out)
	} else {
		// case where server could return either old or new fields
		// #TODO: remove this when old fields are deprecated, v1beta1
		jsonpath = "jsonpath={.items[0].spec.runLatest.configuration.revisionTemplate.metadata.annotations.autoscaling\\.knative\\.dev/maxScale}"
		out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
		assert.NilError(t, err)
		assert.Equal(t, maxScale, out)
	}
}

func (test *e2eTest) validateServiceAnnotations(t *testing.T, serviceName string, annotations map[string]string) {
	metadataAnnotationsJsonpathFormat := "jsonpath={.metadata.annotations.%s}"
	templateAnnotationsJsonpathFormat := "jsonpath={.spec.template.metadata.annotations.%s}"
	oldTemplateAnnotationsJsonpathFormat := "jsonpath={.spec.runLatest.configuration.revisionTemplate.metadata.annotations.%s}"

	for k, v := range annotations {
		out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName, "-o", fmt.Sprintf(metadataAnnotationsJsonpathFormat, k)}, runOpts{})
		assert.NilError(t, err)
		assert.Equal(t, v, out)

		out, err = test.kn.RunWithOpts([]string{"service", "describe", serviceName, "-o", fmt.Sprintf(templateAnnotationsJsonpathFormat, k)}, runOpts{})
		assert.NilError(t, err)
		if out != "" || v == "" {
			assert.Equal(t, v, out)
		} else {
			// case where server returns fields like  spec.runLatest.configuration.revisionTemplate.metadata.annotations
			// TODO: Remove this case when `runLatest` field is deprecated altogether / v1beta1
			out, err := test.kn.RunWithOpts([]string{"service", "describe", serviceName, "-o", fmt.Sprintf(oldTemplateAnnotationsJsonpathFormat, k)}, runOpts{})
			assert.NilError(t, err)
			assert.Equal(t, v, out)
		}
	}
}

func (test *e2eTest) validateContainerField(t *testing.T, serviceName, field, expected string) {
	jsonpath := fmt.Sprintf("jsonpath={.items[0].spec.template.spec.containers[0].%s}", field)
	out, err := test.kn.RunWithOpts([]string{"service", "list", serviceName, "-o", jsonpath}, runOpts{})
	assert.NilError(t, err)
	assert.Equal(t, out, expected)
}
