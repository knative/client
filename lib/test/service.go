// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

func ServiceCreate(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName, "--image", KnDefaultTestImage)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))
}

func ServiceListEmpty(r *KnRunResultCollector) {
	out := r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "No services found."))
}

func ServiceList(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName))
}

func ServiceDescribe(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, serviceName, r.KnTest().Kn().Namespace(), KnDefaultTestImage))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Age", "Revisions"))
}

func ServiceListOutput(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName, "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName, "service.serving.knative.dev"))
}

func ServiceUpdate(r *KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := r.KnTest().Kn().Run(fullArgs...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "updating", "service", serviceName, "ready"))
}

func ServiceDelete(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "delete", "--wait", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Service", serviceName, "successfully deleted in namespace", r.KnTest().Kn().Namespace()))
}

func ServiceDescribeWithJSONPath(r *KnRunResultCollector, serviceName, jsonpath string) string {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o", jsonpath)
	r.AssertNoError(out)
	return out.Stdout
}

func ServiceUpdateWithOptions(r *KnRunResultCollector, serviceName string, options ...string) {
	command := []string{"service", "update", serviceName}
	command = append(command, options...)
	out := r.KnTest().Kn().Run(command...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Service", serviceName, "updating", "namespace", r.KnTest().Kn().Namespace()))
}
