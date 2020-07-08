// Copyright 2020 The Knative Authors

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
)

func TestSourceBinding(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	test.ServiceCreate(r, "testsvc0")

	t.Log("create source binding")
	sourceBindingCreate(r, "my-binding0", "Deployment:apps/v1:myapp", "ksvc:testsvc0")
	sourceBindingListOutputName(r, "my-binding0")

	t.Log("delete source binding")
	sourceBindingDelete(r, "my-binding0")

	t.Log("update source binding")
	sourceBindingCreate(r, "my-binding1", "Deployment:apps/v1:myapp", "ksvc:testsvc0")
	test.ServiceCreate(r, "testsvc1")
	sourceBindingUpdate(r, "my-binding1", "Deployment:apps/v1:myapp", "ksvc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.GetResourceFieldsWithJSONPath(t, it, "sinkbindings.sources.knative.dev", "my-binding1", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")
}

func sourceBindingCreate(r *test.KnRunResultCollector, bindingName string, subject string, sink string) {
	out := r.KnTest().Kn().Run("source", "binding", "create", bindingName, "--subject", subject, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Sink", "binding", bindingName, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func sourceBindingDelete(r *test.KnRunResultCollector, bindingName string) {
	out := r.KnTest().Kn().Run("source", "binding", "delete", bindingName)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Sink", "binding", bindingName, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

func sourceBindingUpdate(r *test.KnRunResultCollector, bindingName string, subject string, sink string) {
	out := r.KnTest().Kn().Run("source", "binding", "update", bindingName, "--subject", subject, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, bindingName, "updated", "namespace", r.KnTest().Kn().Namespace()))
}

func sourceBindingListOutputName(r *test.KnRunResultCollector, bindingName string) {
	out := r.KnTest().Kn().Run("source", "binding", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, bindingName))
}
