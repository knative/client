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

	"knative.dev/client/lib/test"
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
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		err1 := tearDownForSourceApiServer(t, it)
		err2 := it.Teardown()
		assert.NilError(t, err1)
		assert.NilError(t, err2)
	}()

	r := test.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	setupForSourceApiServer(t, it)
	serviceCreate(t, it, r, "testsvc0")

	t.Log("create apiserver sources with a sink to a service")
	apiServerSourceCreate(t, it, r, "testapisource0", "Event:v1:true", "testsa", "svc:testsvc0")
	apiServerSourceCreate(t, it, r, "testapisource1", "Event:v1", "testsa", "svc:testsvc0")

	t.Log("list sources")
	output := sourceList(t, it, r)
	assert.Check(t, util.ContainsAll(output, "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output, "testapisource0", "ApiServerSource", "apiserversources.sources.knative.dev", "svc:testsvc0"))
	assert.Check(t, util.ContainsAll(output, "testapisource1", "ApiServerSource", "apiserversources.sources.knative.dev", "svc:testsvc0"))

	t.Log("list sources in YAML format")
	output = sourceList(t, it, r, "-oyaml")
	assert.Check(t, util.ContainsAll(output, "testapisource1", "ApiServerSource", "Service", "testsvc0"))

	t.Log("delete apiserver sources")
	apiServerSourceDelete(t, it, r, "testapisource0")
	apiServerSourceDelete(t, it, r, "testapisource1")

	t.Log("create apiserver source with a missing sink service")
	apiServerSourceCreateMissingSink(t, it, r, "testapisource2", "Event:v1:true", "testsa", "svc:unknown")

	t.Log("update apiserver source sink service")
	apiServerSourceCreate(t, it, r, "testapisource3", "Event:v1:true", "testsa", "svc:testsvc0")
	serviceCreate(t, it, r, "testsvc1")
	apiServerSourceUpdateSink(t, it, r, "testapisource3", "svc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := getResourceFieldsWithJSONPath(t, it, "apiserversource.sources.knative.dev", "testapisource3", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")
	// TODO(navidshaikh): Verify the source's status with synchronous create/update
}

func apiServerSourceCreate(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, resources string, sa string, sink string) {
	out := it.Kn().Run("source", "apiserver", "create", sourceName, "--resource", resources, "--service-account", sa, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "apiserver", "source", sourceName, "created", "namespace", it.Kn().Namespace()))
}

func apiServerSourceCreateMissingSink(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, resources string, sa string, sink string) {
	out := it.Kn().Run("source", "apiserver", "create", sourceName, "--resource", resources, "--service-account", sa, "--sink", sink)
	r.AssertError(out)
	assert.Check(t, util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
}

func apiServerSourceDelete(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string) {
	out := it.Kn().Run("source", "apiserver", "delete", sourceName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "apiserver", "source", sourceName, "deleted", "namespace", it.Kn().Namespace()))
}

func setupForSourceApiServer(t *testing.T, it *test.KnTest) {
	_, err := test.NewKubectl(it.Kn().Namespace()).Run("create", "serviceaccount", testServiceAccount)
	assert.NilError(t, err)

	_, err = test.Kubectl{}.Run("create", "clusterrole", clusterRolePrefix+it.Kn().Namespace(), "--verb=get,list,watch", "--resource=events,namespaces")
	assert.NilError(t, err)

	_, err = test.Kubectl{}.Run(
		"create",
		"clusterrolebinding",
		clusterRoleBindingPrefix+it.Kn().Namespace(),
		"--clusterrole="+clusterRolePrefix+it.Kn().Namespace(),
		"--serviceaccount="+it.Kn().Namespace()+":"+testServiceAccount)
	assert.NilError(t, err)
}

func tearDownForSourceApiServer(t *testing.T, it *test.KnTest) error {
	saCmd := []string{"delete", "serviceaccount", testServiceAccount}
	_, err := test.NewKubectl(it.Kn().Namespace()).Run(saCmd...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error executing '%s'", strings.Join(saCmd, " ")))
	}

	crCmd := []string{"delete", "clusterrole", clusterRolePrefix + it.Kn().Namespace()}
	_, err = test.Kubectl{}.Run(crCmd...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error executing '%s'", strings.Join(saCmd, " ")))
	}

	crbCmd := []string{"delete", "clusterrolebinding", clusterRoleBindingPrefix + it.Kn().Namespace()}
	_, err = test.Kubectl{}.Run(crbCmd...)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error executing '%s'", strings.Join(saCmd, " ")))
	}
	return nil
}

func apiServerSourceUpdateSink(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, sourceName string, sink string) {
	out := it.Kn().Run("source", "apiserver", "update", sourceName, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", it.Kn().Namespace()))
}

func getResourceFieldsWithJSONPath(t *testing.T, it *test.KnTest, resource, name, jsonpath string) (string, error) {
	out, err := test.NewKubectl(it.Kn().Namespace()).Run("get", resource, name, "-o", jsonpath, "-n", it.Kn().Namespace())
	if err != nil {
		return "", err
	}

	return out, nil
}
