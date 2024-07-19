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

//go:build e2e && !eventing
// +build e2e,!eventing

package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/test"
	pkgtest "knative.dev/pkg/test"
	"knative.dev/serving/pkg/apis/autoscaling"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TestServiceOptions(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	kubectl := test.NewKubectl(it.Namespace())
	r := test.NewKnRunResultCollector(t, it)

	defer r.DumpIfFailed()

	t.Log("create and validate service with concurrency options")
	serviceCreateWithOptions(r, "svc1", "--concurrency-limit", "250", "--scale-target", "300", "--scale-utilization", "50")
	validateServiceConcurrencyTarget(r, "svc1", "300")
	validateServiceConcurrencyLimit(r, "svc1", "250")
	validateServiceConcurrencyUtilization(r, "svc1", "50")

	t.Log("update and validate service with concurrency limit")
	test.ServiceUpdate(r, "svc1", "--concurrency-limit", "300")
	validateServiceConcurrencyLimit(r, "svc1", "300")

	t.Log("update concurrency options with invalid values for service")
	out := r.KnTest().Kn().Run("service", "update", "svc1", "--concurrency-limit", "-1", "--scale-target", "0")
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "should be at least 0.01"))

	t.Log("returns steady concurrency options for service")
	validateServiceConcurrencyLimit(r, "svc1", "300")
	validateServiceConcurrencyTarget(r, "svc1", "300")
	validateServiceConcurrencyUtilization(r, "svc1", "50")

	t.Log("delete service")
	test.ServiceDelete(r, "svc1")

	t.Log("create and validate service with min/max scale options")
	serviceCreateWithOptions(r, "svc2", "--scale-min", "1", "--scale-max", "3")
	validateServiceMinScale(r, "svc2", "1")
	validateServiceMaxScale(r, "svc2", "3")

	t.Log("update and validate service with max scale option")
	test.ServiceUpdate(r, "svc2", "--scale-max", "2")
	validateServiceMaxScale(r, "svc2", "2")

	t.Log("create and validate service with scale options")
	serviceCreateWithOptions(r, "svc2a", "--scale", "5")
	validateServiceMinScale(r, "svc2a", "5")
	validateServiceMaxScale(r, "svc2a", "5")

	t.Log("update and validate service with scale option")
	test.ServiceUpdate(r, "svc2a", "--scale", "2")
	validateServiceMaxScale(r, "svc2a", "2")
	validateServiceMinScale(r, "svc2a", "2")

	t.Log("delete service")
	test.ServiceDelete(r, "svc2")

	t.Log("create, update and validate service with annotations")
	serviceCreateWithOptions(r, "svc3", "--annotation", "alpha=wolf", "--annotation", "brave=horse")
	validateServiceAnnotations(r, "svc3", map[string]string{"alpha": "wolf", "brave": "horse"})
	test.ServiceUpdate(r, "svc3", "--annotation", "alpha=direwolf", "--annotation", "brave-")
	validateServiceAnnotations(r, "svc3", map[string]string{"alpha": "direwolf", "brave": ""})
	test.ServiceDelete(r, "svc3")

	t.Log("create, update and validate service with annotations but -a")
	serviceCreateWithOptions(r, "svc3a", "-a", "alpha=wolf", "-a", "brave=horse")
	validateServiceAnnotations(r, "svc3a", map[string]string{"alpha": "wolf", "brave": "horse"})
	test.ServiceUpdate(r, "svc3a", "-a", "alpha=direwolf", "-a", "brave-")
	validateServiceAnnotations(r, "svc3a", map[string]string{"alpha": "direwolf", "brave": ""})
	test.ServiceDelete(r, "svc3a")

	t.Log("create, update and validate service with scale window option")
	for _, option := range []string{"--scale-window", "--autoscale-window"} {
		serviceCreateWithOptions(r, "svc4", option, "1m")
		validateScaleWindow(r, "svc4", "1m")
		test.ServiceUpdate(r, "svc4", option, "15s")
		validateScaleWindow(r, "svc4", "15s")
		test.ServiceDelete(r, "svc4")
	}

	t.Log("create, update and validate service with cmd and arg options")
	serviceCreateWithOptions(r, "svc5", "--cmd", "/ko-app/helloworld")
	validateContainerField(r, "svc5", "command", "[\"/ko-app/helloworld\"]")
	test.ServiceUpdate(r, "svc5", "--arg", "myArg1", "--arg", "--myArg2")
	validateContainerField(r, "svc5", "args", "[\"myArg1\",\"--myArg2\"]")
	test.ServiceUpdate(r, "svc5", "--arg", "myArg1")
	validateContainerField(r, "svc5", "args", "[\"myArg1\"]")
	test.ServiceDelete(r, "svc5")

	t.Log("create, update and validate service with user defined")
	var uid int64 = 1000
	if uids, ok := os.LookupEnv("TEST_RUN_AS_UID"); ok {
		uid, err = strconv.ParseInt(uids, 10, 64)
		assert.NilError(t, err)
	}

	serviceCreateWithOptions(r, "svc6", "--user", strconv.FormatInt(uid, 10))
	validateUserID(r, "svc6", uid)
	test.ServiceUpdate(r, "svc6", "--user", strconv.FormatInt(uid+1, 10))
	validateUserID(r, "svc6", uid+1)
	test.ServiceDelete(r, "svc6")

	t.Log("create and validate service and revision labels")
	serviceCreateWithOptions(r, "svc7", "--label-service", "svc=helloworld-svc", "--label-revision", "rev=helloworld-rev")
	validateLabels(r, "svc7", map[string]string{"svc": "helloworld-svc"}, map[string]string{"rev": "helloworld-rev"})
	test.ServiceDelete(r, "svc7")

	t.Log("create and validate service resource options")
	serviceCreateWithOptions(r, "svc8", "--limit", "memory=500Mi,cpu=1000m", "--request", "memory=250Mi,cpu=200m")
	test.ValidateServiceResources(r, "svc8", "250Mi", "200m", "500Mi", "1000m")
	test.ServiceDelete(r, "svc8")

	t.Log("create a grpc service and validate port name")
	serviceCreateWithOptions(r, "svc9", "--image", pkgtest.ImagePath("grpc-ping"), "--port", "h2c:8080")
	validatePort(r, "svc9", 8080, "h2c")
	test.ServiceDelete(r, "svc9")

	t.Log("create and validate service with scale init option")
	serviceCreateWithOptions(r, "svc10", "--scale-init", "1")
	validateServiceInitScale(r, "svc10", "1")
	test.ServiceUpdate(r, "svc10", "--scale-init", "2")
	validateServiceInitScale(r, "svc10", "2")
	t.Log("delete service")
	test.ServiceDelete(r, "svc10")

	t.Log("create and validate service with scale init option via --annotation flag")
	serviceCreateWithOptions(r, "svc11", "--annotation", autoscaling.InitialScaleAnnotationKey+"=2")
	validateServiceInitScale(r, "svc11", "2")
	t.Log("delete service")
	test.ServiceDelete(r, "svc11")

	t.Log("create and validate service and revision annotations")
	serviceCreateWithOptions(r, "svc12", "--annotation-service", "svc=helloworld-svc", "--annotation-revision", "rev=helloworld-rev")
	validateServiceAndRevisionAnnotations(r, "svc12", map[string]string{"svc": "helloworld-svc"}, map[string]string{"rev": "helloworld-rev"})
	test.ServiceDelete(r, "svc12")

	t.Log("create and validate service annotations")
	serviceCreateWithOptions(r, "svc13", "--annotation-service", "svc=helloworld-svc")
	validateServiceAndRevisionAnnotations(r, "svc13", map[string]string{"svc": "helloworld-svc"}, nil)
	test.ServiceDelete(r, "svc13")

	t.Log("create and validate revision annotations")
	serviceCreateWithOptions(r, "svc14", "--annotation-revision", "rev=helloworld-rev")
	validateServiceAndRevisionAnnotations(r, "svc14", nil, map[string]string{"rev": "helloworld-rev"})
	test.ServiceDelete(r, "svc14")

	t.Log("create and validate service env vars")
	env := []corev1.EnvVar{
		{Name: "EXAMPLE", Value: "foo"},
		{Name: "EXAMPLE2", Value: "bar"},
	}
	serviceCreateWithOptions(r, "svc15", "--env", "EXAMPLE=foo", "--env", "EXAMPLE2=bar")
	validateServiceEnvVariables(r, "svc15", env)
	test.ServiceDelete(r, "svc15")

	t.Log("create and validate service env-value-from vars")
	_, err = kubectl.Run("create", "-n", it.Namespace(), "configmap", "test-cm", "--from-literal=key=value")
	assert.NilError(t, err)
	env2 := []corev1.EnvVar{
		{Name: "EXAMPLE", ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test-cm",
				},
				Key: "key",
			},
		}},
	}
	serviceCreateWithOptions(r, "svc16", "--env-value-from", "EXAMPLE=cm:test-cm:key")
	validateServiceEnvVariables(r, "svc16", env2)
	test.ServiceDelete(r, "svc16")
	_, err = kubectl.Run("delete", "-n", it.Namespace(), "configmap", "test-cm")
	assert.NilError(t, err)

	t.Log("create and validate service env vars and env-value-from vars")
	_, err = kubectl.Run("create", "-n", it.Namespace(), "configmap", "test-cm2", "--from-literal=key=value")
	assert.NilError(t, err)
	env3 := []corev1.EnvVar{
		{Name: "EXAMPLE", Value: "foo"},
		{Name: "EXAMPLE2", ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "test-cm2",
				},
				Key: "key",
			},
		}},
		{Name: "EXAMPLE3", Value: "bar"},
	}
	serviceCreateWithOptions(r, "svc17", "--env", "EXAMPLE=foo", "--env-value-from", "EXAMPLE2=config-map:test-cm2:key", "--env", "EXAMPLE3=bar")
	validateServiceEnvVariables(r, "svc17", env3)
	test.ServiceDelete(r, "svc17")
	_, err = kubectl.Run("delete", "-n", it.Namespace(), "configmap", "test-cm2")
	assert.NilError(t, err)

	t.Log("create and validate a service with scale activation")
	serviceCreateWithOptions(r, "svc18", "--scale-activation", "2")
	validateServiceActivation(r, "svc18", 2)
}

func serviceCreateWithOptions(r *test.KnRunResultCollector, serviceName string, options ...string) {
	command := []string{"service", "create", serviceName}
	command = append(command, getOptionsWithImage(options...)...)
	out := r.KnTest().Kn().Run(command...)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "service", serviceName, "Creating", "namespace", r.KnTest().Kn().Namespace(), "Ready"))
	r.AssertNoError(out)
}

func getOptionsWithImage(options ...string) []string {
	for _, o := range options {
		if strings.HasPrefix(o, "--image") {
			//options have an image return options
			return options
		}
	}
	//options dont have a image , add HW image
	options = append(options, "--image="+pkgtest.ImagePath("helloworld"))
	return options
}

func validateServiceConcurrencyLimit(r *test.KnRunResultCollector, serviceName, concurrencyLimit string) {
	jsonpath := "jsonpath={.items[0].spec.template.spec.containerConcurrency}"
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, concurrencyLimit)
	r.AssertNoError(out)
}

func validateServiceConcurrencyTarget(r *test.KnRunResultCollector, serviceName, concurrencyTarget string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/target}"
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, concurrencyTarget)
	r.AssertNoError(out)
}

func validateServiceConcurrencyUtilization(r *test.KnRunResultCollector, serviceName, concurrencyUtilization string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/target-utilization-percentage}"
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, concurrencyUtilization)
	r.AssertNoError(out)
}

func validateScaleWindow(r *test.KnRunResultCollector, serviceName, window string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/window}"
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, window)
	r.AssertNoError(out)
}

func validateServiceMinScale(r *test.KnRunResultCollector, serviceName, minScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/min-scale}"
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, minScale)
	r.AssertNoError(out)
}

func validateServiceMaxScale(r *test.KnRunResultCollector, serviceName, maxScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/max-scale}"
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, maxScale)
	r.AssertNoError(out)
}

func validateServiceInitScale(r *test.KnRunResultCollector, serviceName, initScale string) {
	jsonpath := "jsonpath={.items[0].spec.template.metadata.annotations.autoscaling\\.knative\\.dev/initial-scale}"
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, initScale)
	r.AssertNoError(out)
}

func validateServiceAnnotations(r *test.KnRunResultCollector, serviceName string, annotations map[string]string) {
	metadataAnnotationsJsonpathFormat := "jsonpath={.metadata.annotations.%s}"
	templateAnnotationsJsonpathFormat := "jsonpath={.spec.template.metadata.annotations.%s}"

	for k, v := range annotations {
		out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o", fmt.Sprintf(metadataAnnotationsJsonpathFormat, k))
		assert.Equal(r.T(), v, out.Stdout)
		r.AssertNoError(out)

		out = r.KnTest().Kn().Run("service", "describe", serviceName, "-o", fmt.Sprintf(templateAnnotationsJsonpathFormat, k))
		assert.Equal(r.T(), v, out.Stdout)
		r.AssertNoError(out)
	}
}

func validateServiceAndRevisionAnnotations(r *test.KnRunResultCollector, serviceName string, expectedServiceAnnotations, expectedRevisionAnnotations map[string]string) {

	metadataAnnotationsJsonpathFormat := "jsonpath={.metadata.annotations.%s}"
	templateAnnotationsJsonpathFormat := "jsonpath={.spec.template.metadata.annotations.%s}"

	for k, v := range expectedServiceAnnotations {

		out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o", fmt.Sprintf(metadataAnnotationsJsonpathFormat, k))
		assert.Equal(r.T(), v, out.Stdout)
		r.AssertNoError(out)
	}

	for k, v := range expectedRevisionAnnotations {

		out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o", fmt.Sprintf(templateAnnotationsJsonpathFormat, k))
		assert.Equal(r.T(), v, out.Stdout)
		r.AssertNoError(out)
	}
}

func validateLabels(r *test.KnRunResultCollector, serviceName string, expectedServiceLabels, expectedRevisionLabels map[string]string) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-ojson")
	data := json.NewDecoder(strings.NewReader(out.Stdout))
	var service servingv1.Service
	err := data.Decode(&service)
	assert.NilError(r.T(), err)

	gotRevisionLabels := service.Spec.Template.ObjectMeta.GetLabels()
	assert.DeepEqual(r.T(), gotRevisionLabels, expectedRevisionLabels)

	gotServiceLabels := service.ObjectMeta.GetLabels()
	assert.DeepEqual(r.T(), gotServiceLabels, expectedServiceLabels)
}

func validateContainerField(r *test.KnRunResultCollector, serviceName, field, expected string) {
	jsonpath := fmt.Sprintf("jsonpath={.items[0].spec.template.spec.containers[0].%s}", field)
	out := r.KnTest().Kn().Run("service", "list", serviceName, "-o", jsonpath)
	assert.Equal(r.T(), out.Stdout, expected)
	r.AssertNoError(out)
}

func validateUserID(r *test.KnRunResultCollector, serviceName string, uid int64) {
	svc := test.GetServiceFromKNServiceDescribe(r, serviceName)
	assert.Equal(r.T(), *svc.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser, uid)
}

func validatePort(r *test.KnRunResultCollector, serviceName string, portNumber int32, portName string) {
	svc := test.GetServiceFromKNServiceDescribe(r, serviceName)
	assert.Equal(r.T(), svc.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort, portNumber)
	assert.Equal(r.T(), svc.Spec.Template.Spec.Containers[0].Ports[0].Name, portName)
}

func validateServiceEnvVariables(r *test.KnRunResultCollector, serviceName string, envVar []corev1.EnvVar) {
	svc := test.GetServiceFromKNServiceDescribe(r, serviceName)
	for i, env := range svc.Spec.Template.Spec.Containers[0].Env {
		assert.Equal(r.T(), env.Name, envVar[i].Name)
		if envVar[i].ValueFrom != nil {
			if envVar[i].ValueFrom.ConfigMapKeyRef != nil {
				assert.Equal(r.T(), env.ValueFrom.ConfigMapKeyRef.Key, envVar[i].ValueFrom.ConfigMapKeyRef.Key)
			} else if envVar[i].ValueFrom.SecretKeyRef != nil {
				assert.Equal(r.T(), env.ValueFrom.SecretKeyRef.Key, envVar[i].ValueFrom.SecretKeyRef.Key)
			}
		} else {
			assert.Equal(r.T(), env.Value, envVar[i].Value)
		}
	}
}

func validateServiceActivation(r *test.KnRunResultCollector, serviceName string, activation int) {
	out := r.KnTest().Kn().Run("service", "describe", serviceName, "-o", "json")
	r.AssertNoError(out)
	var svc servingv1.Service
	json.Unmarshal([]byte(out.Stdout), &svc)
	activationStr := strconv.Itoa(activation)
	assert.Check(r.T(), svc.Spec.Template.Annotations[autoscaling.ActivationScaleKey] == activationStr)
}
