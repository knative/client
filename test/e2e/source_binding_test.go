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

	r := test.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	serviceCreate(t, it, r, "testsvc0")

	t.Log("create source binding")
	sourceBindingCreate(t, it, r, "my-binding0", "Deployment:apps/v1:myapp", "svc:testsvc0")
	sourceBindingListOutputName(t, it, r, "my-binding0")

	t.Log("delete source binding")
	sourceBindingDelete(t, it, r, "my-binding0")

	t.Log("update source binding")
	sourceBindingCreate(t, it, r, "my-binding1", "Deployment:apps/v1:myapp", "svc:testsvc0")
	serviceCreate(t, it, r, "testsvc1")
	sourceBindingUpdate(t, it, r, "my-binding1", "Deployment:apps/v1:myapp", "svc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := getResourceFieldsWithJSONPath(t, it, "sinkbindings.sources.knative.dev", "my-binding1", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")
}

func sourceBindingCreate(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, bindingName string, subject string, sink string) {
	out := it.Kn().Run("source", "binding", "create", bindingName, "--subject", subject, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Sink", "binding", bindingName, "created", "namespace", it.Kn().Namespace()))
}

func sourceBindingDelete(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, bindingName string) {
	out := it.Kn().Run("source", "binding", "delete", bindingName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Sink", "binding", bindingName, "deleted", "namespace", it.Kn().Namespace()))
}

func sourceBindingUpdate(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, bindingName string, subject string, sink string) {
	out := it.Kn().Run("source", "binding", "update", bindingName, "--subject", subject, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, bindingName, "updated", "namespace", it.Kn().Namespace()))
}

func sourceBindingListOutputName(t *testing.T, it *test.KnTest, r *test.KnRunResultCollector, bindingName string) {
	out := it.Kn().Run("source", "binding", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, bindingName))
}
