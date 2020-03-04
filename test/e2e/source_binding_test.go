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
	"knative.dev/client/pkg/util"
)

func TestSourceBinding(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	test.serviceCreate(t, r, "testsvc0")

	t.Log("create source binding")
	test.sourceBindingCreate(t, r, "my-binding0", "Deployment:apps/v1:myapp", "svc:testsvc0")

	t.Log("delete source binding")
	test.sourceBindingDelete(t, r, "my-binding0")

	t.Log("update source binding")
	test.sourceBindingCreate(t, r, "my-binding1", "Deployment:apps/v1:myapp", "svc:testsvc0")
	test.serviceCreate(t, r, "testsvc1")
	test.sourceBindingUpdate(t, r, "my-binding1", "Deployment:apps/v1:myapp", "svc:testsvc1")
	jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
	out, err := test.getResourceFieldsWithJSONPath("sinkbindings.sources.knative.dev", "my-binding1", jpSinkRefNameInSpec)
	assert.NilError(t, err)
	assert.Equal(t, out, "testsvc1")
}

func (test *e2eTest) sourceBindingCreate(t *testing.T, r *KnRunResultCollector, bindingName string, subject string, sink string) {
	out := test.kn.Run("source", "binding", "create", bindingName, "--subject", subject, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Sink", "binding", bindingName, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) sourceBindingDelete(t *testing.T, r *KnRunResultCollector, bindingName string) {
	out := test.kn.Run("source", "binding", "delete", bindingName)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAllIgnoreCase(out.Stdout, "Sink", "binding", bindingName, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) sourceBindingUpdate(t *testing.T, r *KnRunResultCollector, bindingName string, subject string, sink string) {
	out := test.kn.Run("source", "binding", "update", bindingName, "--subject", subject, "--sink", sink)
	r.AssertNoError(out)
	assert.Check(t, util.ContainsAll(out.Stdout, bindingName, "updated", "namespace", test.kn.namespace))
}
