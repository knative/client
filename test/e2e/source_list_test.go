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

//go:build e2e && !serving
// +build e2e,!serving

package e2e

import (
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/test"
)

func TestSourceListTypes(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("List available source types")
	output := sourceListTypes(r)
	assert.Check(t, util.ContainsAll(output, "TYPE", "S", "NAME", "DESCRIPTION", "Ping", "ApiServer"))
}

func TestSourceList(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, tearDownForSourceAPIServer(t, it))
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)

	defer r.DumpIfFailed()
	setupForSourceAPIServer(t, it)
	test.ServiceCreate(r, "testsvc0")

	t.Log("List sources empty case")
	output := sourceList(r)
	assert.Check(t, util.ContainsAll(output, "No", "sources", "found."))
	assert.Check(t, util.ContainsNone(output, "NAME", "TYPE", "RESOURCE", "SINK", "READY"))

	t.Log("Create API Server")
	apiServerSourceCreate(r, "testapisource0", "Event:v1:key1=value1", "testsa", "ksvc:testsvc0")
	apiServerSourceListOutputName(r, "testapisource0")

	t.Log("Create source binding")
	sourceBindingCreate(r, "my-binding0", "Deployment:apps/v1:myapp", "ksvc:testsvc0")
	sourceBindingListOutputName(r, "my-binding0")

	t.Log("Create ping source")
	pingSourceCreate(r, "testpingsource0", "* * * * */1", "ping", "ksvc:testsvc0")
	pingSourceListOutputName(r, "testpingsource0")

	t.Log("List sources filter valid case")
	output = sourceList(r, "--type", "PingSource")
	assert.Check(t, util.ContainsAll(output, "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output, "testpingsource0", "PingSource", "pingsources.sources.knative.dev", "ksvc:testsvc0"))

	t.Log("List sources filter invalid case")
	output = sourceList(r, "--type", "testapisource0")
	assert.Check(t, util.ContainsAll(output, "No", "sources", "found."))
	output = sourceList(r, "--type", "TestSource", "-oyaml")
	assert.Check(t, util.ContainsAll(output, "apiVersion", "client.knative.dev/v1alpha1", "items", "[]", "kind", "SourceList"))

	t.Log("List available source in YAML format")
	output = sourceList(r, "--type", "PingSource,ApiServerSource", "-oyaml")
	assert.Check(t, util.ContainsAll(output, "testpingsource0", "PingSource", "Service", "testsvc0"))
	assert.Check(t, util.ContainsAll(output, "testapisource0", "ApiServerSource", "Service", "testsvc0"))

	t.Log("Delete apiserver sources")
	apiServerSourceDelete(r, "testapisource0")
	t.Log("Delete source binding")
	sourceBindingDelete(r, "my-binding0")
	t.Log("Delete Ping sources")
	pingSourceDelete(r, "testpingsource0")
	// non empty list case is tested in test/e2e/source_apiserver_it.go where source setup is present
}

func sourceListTypes(r *test.KnRunResultCollector, args ...string) string {
	cmd := append([]string{"source", "list-types"}, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	return out.Stdout
}

func sourceList(r *test.KnRunResultCollector, args ...string) string {
	cmd := append([]string{"source", "list"}, args...)
	out := r.KnTest().Kn().Run(cmd...)
	r.AssertNoError(out)
	return out.Stdout
}
