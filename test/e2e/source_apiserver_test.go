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
	"strings"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

const (
	testServiceAccount = "apiserver-sa"
	// use following prefix + current namespace to generate ClusterRole and ClusterRoleBinding names
	clusterRolePrefix        = "apiserver-role-"
	clusterRoleBindingPrefix = "apiserver-binding-"
)

func TestSourceApiServer(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer func() {
		test.tearDownForSourceApiServer(t)
		test.Teardown(t)
	}()

	test.setupForSourceApiServer(t)
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

func (test *e2eTest) setupForSourceApiServer(t *testing.T) {
	kubectl := kubectl{t, Logger{}}

	saCmd := []string{"create", "serviceaccount", testServiceAccount, "--namespace", test.kn.namespace}
	_, err := kubectl.RunWithOpts(saCmd, runOpts{})
	if err != nil {
		t.Fatalf("Error executing '%s'. Error: %s", strings.Join(saCmd, " "), err.Error())
	}

	crCmd := []string{"create", "clusterrole", clusterRolePrefix + test.kn.namespace, "--verb=get,list,watch", "--resource=events,namespaces"}
	_, err = kubectl.RunWithOpts(crCmd, runOpts{})
	if err != nil {
		t.Fatalf("Error executing '%s'. Error: %s", strings.Join(crCmd, " "), err.Error())
	}

	crbCmd := []string{"create", "clusterrolebinding", clusterRoleBindingPrefix + test.kn.namespace, "--clusterrole=" + clusterRolePrefix + test.kn.namespace, "--serviceaccount=" + test.kn.namespace + ":" + testServiceAccount}
	_, err = kubectl.RunWithOpts(crbCmd, runOpts{})
	if err != nil {
		t.Fatalf("Error executing '%s'. Error: %s", strings.Join(crbCmd, " "), err.Error())
	}
}

func (test *e2eTest) tearDownForSourceApiServer(t *testing.T) {
	kubectl := kubectl{t, Logger{}}

	saCmd := []string{"delete", "serviceaccount", testServiceAccount, "--namespace", test.kn.namespace}
	_, err := kubectl.RunWithOpts(saCmd, runOpts{})
	if err != nil {
		t.Fatalf("Error executing '%s'. Error: %s", strings.Join(saCmd, " "), err.Error())
	}

	crCmd := []string{"delete", "clusterrole", clusterRolePrefix + test.kn.namespace}
	_, err = kubectl.RunWithOpts(crCmd, runOpts{})
	if err != nil {
		t.Fatalf("Error executing '%s'. Error: %s", strings.Join(crCmd, " "), err.Error())
	}

	crbCmd := []string{"delete", "clusterrolebinding", clusterRoleBindingPrefix + test.kn.namespace}
	_, err = kubectl.RunWithOpts(crbCmd, runOpts{})
	if err != nil {
		t.Fatalf("Error executing '%s'. Error: %s", strings.Join(crbCmd, " "), err.Error())
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
