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
	"encoding/json"
	"strings"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	pkgtest "knative.dev/pkg/test"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/util"
)

// ServiceCreate verifies given service creation in sync mode and also verifies output
func ServiceCreate(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "create", serviceName, "--image", pkgtest.ImagePath("helloworld"))
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "service", serviceName, "creating", "namespace", r.KnTest().Kn().Namespace(), "ready"))
}

// ServiceListEmpty verifies that there are no services present
func ServiceListEmpty(r *KnRunResultCollector) {
	out := r.KnTest().Kn().Run("service", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "No services found."))
}

// ServiceList verifies if given service exists
func ServiceList(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName))
}

// ServiceDescribe describes given service and verifies the keys in the output
func ServiceDescribe(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName)
	r.AssertNoError(out)
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, serviceName, r.KnTest().Kn().Namespace(), pkgtest.ImagePath("helloworld")))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Conditions", "ConfigurationsReady", "Ready", "RoutesReady"))
	assert.Assert(r.T(), util.ContainsAll(out.Stdout, "Name", "Namespace", "URL", "Age", "Revisions"))
}

// ServiceListOutput verifies listing given service using '--output name' flag
func ServiceListOutput(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "list", serviceName, "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, serviceName, "service.serving.knative.dev"))
}

// ServiceUpdate verifies service update operation with given arguments in sync mode
func ServiceUpdate(r *KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := r.KnTest().Kn().Run(fullArgs...)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "updating", "service", serviceName, "ready"))
}

// ServiceUpdateWithError verifies service update operation with given arguments in sync mode
// when expecting an error
func ServiceUpdateWithError(r *KnRunResultCollector, serviceName string, args ...string) {
	fullArgs := append([]string{}, "service", "update", serviceName)
	fullArgs = append(fullArgs, args...)
	out := r.KnTest().Kn().Run(fullArgs...)
	r.AssertError(out)
}

// ServiceDelete verifies service deletion in sync mode
func ServiceDelete(r *KnRunResultCollector, serviceName string) {
	out := r.KnTest().Kn().Run("service", "delete", "--wait", serviceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "Service", serviceName, "successfully deleted in namespace", r.KnTest().Kn().Namespace()))
}

// ServiceDescribeWithJSONPath returns output of given JSON path by describing the service
func ServiceDescribeWithJSONPath(r *KnRunResultCollector, serviceName, jsonpath string) string {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o", jsonpath)
	r.AssertNoError(out)
	return out.Stdout
}

func ValidateServiceResources(r *KnRunResultCollector, serviceName string, requestsMemory, requestsCPU, limitsMemory, limitsCPU string) {
	var err error
	rlist := corev1.ResourceList{}
	rlist[corev1.ResourceCPU], err = resource.ParseQuantity(requestsCPU)
	assert.NilError(r.T(), err)
	rlist[corev1.ResourceMemory], err = resource.ParseQuantity(requestsMemory)
	assert.NilError(r.T(), err)

	llist := corev1.ResourceList{}
	llist[corev1.ResourceCPU], err = resource.ParseQuantity(limitsCPU)
	assert.NilError(r.T(), err)
	llist[corev1.ResourceMemory], err = resource.ParseQuantity(limitsMemory)
	assert.NilError(r.T(), err)

	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-ojson")
	data := json.NewDecoder(strings.NewReader(out.Stdout))
	var service servingv1.Service
	err = data.Decode(&service)
	assert.NilError(r.T(), err)

	serviceRequestResourceList := service.Spec.Template.Spec.Containers[0].Resources.Requests
	assert.DeepEqual(r.T(), serviceRequestResourceList, rlist)

	serviceLimitsResourceList := service.Spec.Template.Spec.Containers[0].Resources.Limits
	assert.DeepEqual(r.T(), serviceLimitsResourceList, llist)
}
