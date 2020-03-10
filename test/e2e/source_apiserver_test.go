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
	"strings"
	"testing"

	"github.com/pkg/errors"
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
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		err1 := test.tearDownForSourceApiServer()
		err2 := test.Teardown()
		assert.NilError(t, err1)
		assert.NilError(t, err2)
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	test.setupForSourceApiServer(t)
	test.serviceCreate(t, r, "testsvc0")

	t.Log("create apiserver sources with a sink to a service")
	test.apiServerSourceCreate(t, r, "testapisource0", "Event:v1:true", "testsa", "svc:testsvc0")
	test.apiServerSourceCreate(t, r, "testapisource1", "Event:v1", "testsa", "svc:testsvc0")

	t.Log("list sources")
	output := test.sourceList(t, r)
	assert.Check(t, util.ContainsAll(output, "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output, "testapisource0", "ApiServerSource", "apiserversources.sources.knative.dev", "svc:testsvc0"))
	assert.Check(t, util.ContainsAll(output, "testapisource1", "ApiServerSource", "apiserversources.sources.knative.dev", "svc:testsvc0"))

	t.Log("list sources in YAML format")
	output = test.sourceList(t, r, "-oyaml")
	assert.Check(t, util.ContainsAll(output, "testapisource1", "ApiServerSource", "Service", "testsvc0"))

	t.Log("delete apiserver sources")
	test.apiServerSourceDelete(t, r, "testapisource0")
	test.apiServerSourceDelete(t, r, "testapisource1")

	t.Log("create apiserver source with a missing sink service")
	test.apiServerSourceCreateMissingSink(t, r, "testapisource2", "Event:v1:true", "testsa", "svc:unknown")

	t.Log("update apiserver source sink service")
	test.apiServerSourceCreate(t, r, "testapisource3", "Event:v1:true", "testsa", "svc:testsvc0")
	test.serviceCreate(t, r, "testsvc1")
	test.apiServerSourceUpdateSink(t, r, "testapisource3", "svc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.getResourceFieldsWithJSONPath("apiserversource.sources.knative.dev", "testapisource3", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")
	// TODO(navidshaikh): Verify the source's status with synchronous create/update
}

func (test *e2eTest) apiServerSourceCreate(t *testing.T, r *KnRunResultCollector, sourceName string, resources string, sa string, sink string) {
	out := test.kn.Run("source", "apiserver", "create", sourceName, "--resource", resources, "--service-account", sa, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "apiserver", "source", sourceName, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) apiServerSourceCreateMissingSink(t *testing.T, r *KnRunResultCollector, sourceName string, resources string, sa string, sink string) {
	out := test.kn.Run("source", "apiserver", "create", sourceName, "--resource", resources, "--service-account", sa, "--sink", sink)
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
}

func (test *e2eTest) apiServerSourceDelete(t *testing.T, r *KnRunResultCollector, sourceName string) {
	out := test.kn.Run("source", "apiserver", "delete", sourceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "apiserver", "source", sourceName, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) setupForSourceApiServer(t *testing.T) {
	_, err := kubectl{test.kn.namespace}.Run("create", "serviceaccount", testServiceAccount)
	assert.NilError(t, err)

	_, err = kubectl{}.Run("create", "clusterrole", clusterRolePrefix+test.kn.namespace, "--verb=get,list,watch", "--resource=events,namespaces")
	assert.NilError(t, err)

	_, err = kubectl{}.Run(
		"create",
		"clusterrolebinding",
		clusterRoleBindingPrefix+test.kn.namespace,
		"--clusterrole="+clusterRolePrefix+test.kn.namespace,
		"--serviceaccount="+test.kn.namespace+":"+testServiceAccount)
	assert.NilError(t, err)
}

func (test *e2eTest) tearDownForSourceApiServer() error {

	saCmd := []string{"delete", "serviceaccount", testServiceAccount}
	_, err := kubectl{test.kn.namespace}.Run(saCmd...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error executing '%s'", strings.Join(saCmd, " ")))
	}

	crCmd := []string{"delete", "clusterrole", clusterRolePrefix + test.kn.namespace}
	_, err = kubectl{}.Run(crCmd...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error executing '%s'", strings.Join(saCmd, " ")))
	}

	crbCmd := []string{"delete", "clusterrolebinding", clusterRoleBindingPrefix + test.kn.namespace}
	_, err = kubectl{}.Run(crbCmd...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error executing '%s'", strings.Join(saCmd, " ")))
	}
	return nil
}

func (test *e2eTest) apiServerSourceUpdateSink(t *testing.T, r *KnRunResultCollector, sourceName string, sink string) {
	out := test.kn.Run("source", "apiserver", "update", sourceName, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", test.kn.namespace))
}

func (test *e2eTest) getResourceFieldsWithJSONPath(resource, name, jsonpath string) (string, error) {
	out, err := kubectl{test.kn.namespace}.Run("get", resource, name, "-o", jsonpath, "-n", test.kn.namespace)
	if err != nil {
		return "", err
	}

	return out, nil
}
