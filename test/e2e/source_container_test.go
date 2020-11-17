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
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
	pkgtest "knative.dev/pkg/test"
)

func TestSourceContainer(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		err := it.Teardown()
		assert.NilError(t, err)
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	test.ServiceCreate(r, "testsvc0")

	t.Log("create container source with a sink to a service")
	containerSourceCreate(r, "testsource0", "ksvc:testsvc0")
	containerSourceListOutputName(r, "testsource0")

	t.Log("list container sources")
	containerSourceList(r, "testsource0")

	t.Log("delete container sources")
	containerSourceDelete(r, "testsource0")

	t.Log("create container source with a missing sink service")
	containerSourceCreateMissingSink(r, "testsource2", "ksvc:unknown")

	t.Log("update container source sink service")
	containerSourceCreate(r, "testsource3", "ksvc:testsvc0")
	test.ServiceCreate(r, "testsvc1")
	containerSourceUpdateSink(r, "testsource3", "ksvc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.GetResourceFieldsWithJSONPath(t, it, "containersource.sources.knative.dev", "testsource3", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")
}

func containerSourceCreate(r *test.KnRunResultCollector, sourceName string, sink string) {
	out := r.KnTest().Kn().Run("source", "container", "create", sourceName, "--image", pkgtest.ImagePath("grpc-ping"), "--port", "h2c:8080", "--sink", sink)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "container", "source", sourceName, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func containerSourceListOutputName(r *test.KnRunResultCollector, containerSources ...string) {
	out := r.KnTest().Kn().Run("source", "container", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, containerSources...))
}

func containerSourceList(r *test.KnRunResultCollector, containerSources ...string) {
	out := r.KnTest().Kn().Run("source", "container", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "NAME", "IMAGE", "SINK", "READY"))
	assert.Check(r.T(), util.ContainsAll(out.Stdout, containerSources...))
	assert.Check(r.T(), util.ContainsAll(out.Stdout, "grpc-ping", "ksvc:testsvc0"))
}

func containerSourceCreateMissingSink(r *test.KnRunResultCollector, sourceName string, sink string) {
	out := r.KnTest().Kn().Run("source", "container", "create", sourceName, "--image", pkgtest.ImagePath("grpc-ping"), "--port", "h2c:8080", "--sink", sink)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, "services.serving.knative.dev", "not found"))
}

func containerSourceDelete(r *test.KnRunResultCollector, sourceName string) {
	out := r.KnTest().Kn().Run("source", "container", "delete", sourceName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "container", "source", sourceName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

func containerSourceUpdateSink(r *test.KnRunResultCollector, sourceName string, sink string) {
	out := r.KnTest().Kn().Run("source", "container", "update", sourceName, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, sourceName, "updated", "namespace", r.KnTest().Kn().Namespace()))
}
