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

	"gotest.tools/v3/assert"

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
		err1 := tearDownForSourceAPIServer(t, it)
		err2 := it.Teardown()
		assert.NilError(t, err1)
		assert.NilError(t, err2)
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	setupForSourceAPIServer(t, it)
	test.ServiceCreate(r, "testsvc0")

	t.Log("create apiserver sources with a sink to a service")
	apiServerSourceCreate(r, "testapisource0", "Event:v1:key1=value1", "testsa", "ksvc:testsvc0")
	apiServerSourceCreate(r, "testapisource1", "Event:v1", "testsa", "ksvc:testsvc0")
	apiServerSourceListOutputName(r, "testapisource0", "testapisource1")

	t.Log("list sources")
	output := sourceList(r)
	assert.Check(t, util.ContainsAll(output, "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output, "testapisource0", "ApiServerSource", "apiserversources.sources.knative.dev", "ksvc:testsvc0"))
	assert.Check(t, util.ContainsAll(output, "testapisource1", "ApiServerSource", "apiserversources.sources.knative.dev", "ksvc:testsvc0"))

	t.Log("list sources in YAML format")
	output = sourceList(r, "-oyaml")
	assert.Check(t, util.ContainsAll(output, "testapisource1", "ApiServerSource", "Service", "testsvc0"))

	t.Log("delete apiserver sources")
	apiServerSourceDelete(r, "testapisource0")
	apiServerSourceDelete(r, "testapisource1")

	t.Log("create apiserver source with a missing sink service")
	apiServerSourceCreateMissingSink(r, "testapisource2", "Event:v1", "testsa", "ksvc:unknown")

	t.Log("update apiserver source sink service")
	apiServerSourceCreate(r, "testapisource3", "Event:v1", "testsa", "ksvc:testsvc0")
	test.ServiceCreate(r, "testsvc1")
	apiServerSourceUpdateSink(r, "testapisource3", "ksvc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.GetResourceFieldsWithJSONPath(t, it, "apiserversource.sources.knative.dev", "testapisource3", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")
	// TODO(navidshaikh): Verify the source's status with synchronous create/update
}

func apiServerSourceCreate(r *test.KnRunResultCollector, sourceName string, resources string, sa string, sink string) {
	out := r.KnTest().Kn().Run("source", "apiserver", "create", sourceName, "--resource", resources, "--service-account", sa, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "apiserver", "source", sourceName, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func apiServerSourceListOutputName(r *test.KnRunResultCollector, apiserverSources ...string) {
	out := r.KnTest().Kn().Run("source", "apiserver", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, apiserverSources...))
}

func apiServerSourceCreateMissingSink(r *test.KnRunResultCollector, sourceName string, resources string, sa string, sink string) {
	out := r.KnTest().Kn().Run("source", "apiserver", "create", sourceName, "--resource", resources, "--service-account", sa, "--sink", sink)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
}

func apiServerSourceDelete(r *test.KnRunResultCollector, sourceName string) {
	out := r.KnTest().Kn().Run("source", "apiserver", "delete", sourceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "apiserver", "source", sourceName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

func setupForSourceAPIServer(t *testing.T, it *test.KnTest) {
	_, err := test.NewKubectl(it.Kn().Namespace()).Run("create", "serviceaccount", testServiceAccount)
	assert.NilError(t, err)

	_, err = test.Kubectl{}.Run("create", "role", clusterRolePrefix+it.Kn().Namespace(), "--verb=get,list,watch", "--resource=events,namespaces")
	assert.NilError(t, err)

	_, err = test.Kubectl{}.Run(
		"create",
		"rolebinding",
		clusterRoleBindingPrefix+it.Kn().Namespace(),
		"--role="+clusterRolePrefix+it.Kn().Namespace(),
		"--serviceaccount="+it.Kn().Namespace()+":"+testServiceAccount)
	assert.NilError(t, err)
}

func tearDownForSourceAPIServer(t *testing.T, it *test.KnTest) error {
	saCmd := []string{"delete", "serviceaccount", testServiceAccount}
	_, err := test.NewKubectl(it.Kn().Namespace()).Run(saCmd...)
	if err != nil {
		return fmt.Errorf("Error executing %q: %w", strings.Join(saCmd, " "), err)
	}

	crCmd := []string{"delete", "role", clusterRolePrefix + it.Kn().Namespace()}
	_, err = test.Kubectl{}.Run(crCmd...)
	if err != nil {
		return fmt.Errorf("Error executing %q: %w", strings.Join(saCmd, " "), err)
	}

	crbCmd := []string{"delete", "rolebinding", clusterRoleBindingPrefix + it.Kn().Namespace()}
	_, err = test.Kubectl{}.Run(crbCmd...)
	if err != nil {
		return fmt.Errorf("Error executing %q: %w", strings.Join(saCmd, " "), err)
	}
	return nil
}

func apiServerSourceUpdateSink(r *test.KnRunResultCollector, sourceName string, sink string) {
	out := r.KnTest().Kn().Run("source", "apiserver", "update", sourceName, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", r.KnTest().Kn().Namespace()))
}
