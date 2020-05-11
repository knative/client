// Copyright 2019 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build e2e
// +build !eventing

package e2e

import (
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestBasicWorkflow(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("returns no service before running tests")
	test.ServiceListEmpty(r)

	t.Log("create hello service and return no error")
	test.ServiceCreate(r, "hello")

	t.Log("return valid info about hello service")
	test.ServiceList(r, "hello")
	test.ServiceDescribe(r, "hello")

	t.Log("return list --output name about hello service")
	test.ServiceListOutput(r, "hello")

	t.Log("update hello service's configuration and return no error")
	test.ServiceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("create another service and return no error")
	test.ServiceCreate(r, "svc2")

	t.Log("return a list of revisions associated with hello and svc2 services")
	test.RevisionListForService(r, "hello")
	test.RevisionListForService(r, "svc2")

	t.Log("describe revision from hello service")
	test.RevisionDescribe(r, "hello")

	t.Log("delete hello and svc2 services and return no error")
	test.ServiceDelete(r, "hello")
	test.ServiceDelete(r, "svc2")

	t.Log("return no service after completing tests")
	test.ServiceListEmpty(r)
}

func TestWrongCommand(t *testing.T) {
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	out := test.Kn{}.Run("source", "apiserver", "noverb", "--tag=0.13")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown subcommand", "noverb"))
	r.AssertError(out)

	out = test.Kn{}.Run("rev")
	assert.Check(t, util.ContainsAll(out.Stderr, "Error", "unknown command", "rev"))
	r.AssertError(out)

}
