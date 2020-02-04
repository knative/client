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
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	test.serviceCreate(t, "testsvc0")

	t.Run("create source binding", func(t *testing.T) {
		test.sourceBindingCreate(t, "my-binding0", "Deployment:apps/v1:myapp", "svc:testsvc0")
	})

	t.Run("delete source binding", func(t *testing.T) {
		test.sourceBindingDelete(t, "my-binding0")
	})

	t.Run("update source binding", func(t *testing.T) {
		test.sourceBindingCreate(t, "my-binding1", "Deployment:apps/v1:myapp", "svc:testsvc0")
		test.serviceCreate(t, "testsvc1")
		test.sourceBindingUpdate(t, "my-binding1", "Deployment:apps/v1:myapp", "svc:testsvc1")
		jpSinkRefNameInSpec := "jsonpath={.spec.sink.ref.name}"
		out, err := test.getResourceFieldsWithJSONPath(t, "sinkbindings.sources.knative.dev", "my-binding1", jpSinkRefNameInSpec)
		assert.NilError(t, err)
		assert.Equal(t, out, "testsvc1")
	})

}

func (test *e2eTest) sourceBindingCreate(t *testing.T, bindingName string, subject string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"source", "binding", "create", bindingName,
		"--subject", subject, "--sink", sink}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "Sink", "binding", bindingName, "created", "namespace", test.kn.namespace))
}

func (test *e2eTest) sourceBindingDelete(t *testing.T, bindingName string) {
	out, err := test.kn.RunWithOpts([]string{"source", "binding", "delete", bindingName}, runOpts{NoNamespace: false})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAllIgnoreCase(out, "Sink", "binding", bindingName, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) sourceBindingUpdate(t *testing.T, bindingName string, subject string, sink string) {
	out, err := test.kn.RunWithOpts([]string{"source", "binding", "update", bindingName,
		"--subject", subject, "--sink", sink}, runOpts{})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(out, bindingName, "updated", "namespace", test.kn.namespace))
}
