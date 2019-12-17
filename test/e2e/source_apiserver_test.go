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
// +build !serving

package e2e

import (
	"fmt"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

func TestSourceApiServer(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	test.setupServiceAccountForApiserver(t, "testsa")
	test.serviceCreate(t, "testsvc0")

	t.Run("create apiserver sources with a sink to a service", func(t *testing.T) {
		test.apiServerSourceCreate(t, "testapisource0", "Event:v1:true", "testsa", "svc:testsvc0")
		test.apiServerSourceCreate(t, "testapisource1", "Event:v1", "testsa", "svc:testsvc0")
	})

	t.Run("delete apiserver sources", func(t *testing.T) {
		test.apiServerSourceDelete(t, "testapisource0")
		test.apiServerSourceDelete(t, "testapisource1")
	})

	t.Run("create apiserver source with a missing sink service", func(t *testing.T) {
		test.apiServerSourceCreateMissingSink(t, "testapisource2", "Event:v1:true", "testsa", "svc:unknown")
	})

	t.Run("update apiserver source sink service", func(t *testing.T) {
		test.apiServerSourceCreate(t, "testapisource3", "Event:v1:true", "testsa", "svc:testsvc0")
		test.serviceCreate(t, "testsvc1")
		test.apiServerSourceUpdateSink(t, "testapisource3", "svc:testsvc1")
		jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
		out, err := test.getResourceFieldsWithJSONPath(t, "apiserversource", "testapisource3", jpSinkRefNameInSpec)
		assert.NilError(t, err)
		assert.Equal(t, out, "testsvc1")
		// TODO(navidshaikh): Verify the source's status with synchronous create/update
	})
}

func (test *e2eTest) apiServerSourceCreate(t *testing.T, sourceName string, resources string, sa string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"source", "apiserver", "create", sourceName,
		"--resource", resources, "--service-account", sa, "--sink", sink}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "apiserver", "source", sourceName, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) apiServerSourceCreateMissingSink(t *testing.T, sourceName string, resources string, sa string, sink string) {
	_, err := test.kn.RunWithOpts([]string{"source", "apiserver", "create", sourceName,
		"--resource", resources, "--service-account", sa, "--sink", sink}, runOpts{NoNamespace: false, AllowError: true})
	assert.ErrorContains(t, err, "services.serving.knative.dev", "not found")
}

func (test *e2eTest) apiServerSourceDelete(t *testing.T, sourceName string) {
	out, err := test.kn.RunWithOpts([]string{"source", "apiserver", "delete", sourceName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "apiserver", "source", sourceName, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) setupServiceAccountForApiserver(t *testing.T, name string) {
	kubectl := kubectl{t, Logger{}}

	_, err := kubectl.RunWithOpts([]string{"create", "serviceaccount", name, "--namespace",test.kn.namespace}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl create serviceaccount test-sa'. Error: %s", err.Error()))
	}
	_, err = kubectl.RunWithOpts([]string{"create", "clusterrole", "testsa-role", "--verb=get,list,watch", "--resource=events,namespaces"}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl clusterrole testsa-role'. Error: %s", err.Error()))
	}
	_, err = kubectl.RunWithOpts([]string{"create", "clusterrolebinding", "testsa-binding", "--clusterrole=testsa-role", "--serviceaccount=" + test.kn.namespace + ":" + name}, runOpts{})
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error executing 'kubectl clusterrolebinding testsa-binding'. Error: %s", err.Error()))
	}
}

func (test *e2eTest) apiServerSourceUpdateSink(t *testing.T, sourceName string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"source", "apiserver", "update", sourceName, "--sink", sink}, runOpts{})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, sourceName, "updated", "namespace", test.kn.namespace))
}

func (test *e2eTest) getResourceFieldsWithJSONPath(t *testing.T, resource, name, jsonpath string) (string, error) {
	kubectl := kubectl{t, Logger{}}
	out, err := kubectl.RunWithOpts([]string{"get", resource, name, "-o", jsonpath, "-n", test.kn.namespace}, runOpts{})
	if err != nil {
		return "", err
	}

	return out, nil
}
