// Copyright 2020 The Knative Authors

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
// +build !serving

package e2e

import (
	"testing"

	"gotest.tools/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestBroker(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("create broker, list and describe it")
	brokerCreate(r, "foo1")
	verifyBrokerList(r, "foo1")
	verifyBrokerListOutputName(r, "foo1")
	verifyBrokerDescribe(r, "foo1")
	brokerDelete(r, "foo1")

	t.Log("create broker and delete it")
	brokerCreate(r, "foo2")
	verifyBrokerList(r, "foo2")
	brokerDelete(r, "foo2")
	verifyBrokerNotfound(r, "foo2")

	t.Log("create multiple brokers and list them")
	brokerCreate(r, "foo3")
	brokerCreate(r, "foo4")
	verifyBrokerList(r, "foo3", "foo4")
	verifyBrokerListOutputName(r, "foo3", "foo4")
	brokerDelete(r, "foo3")
	verifyBrokerNotfound(r, "foo3")
	brokerDelete(r, "foo4")
	verifyBrokerNotfound(r, "foo4")
}

// Private functions

func brokerCreate(r *test.KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("broker", "create", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Broker", name, "created", "namespace", r.KnTest().Kn().Namespace()))
}

func brokerDelete(r *test.KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("broker", "delete", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, "Broker", name, "deleted", "namespace", r.KnTest().Kn().Namespace()))
}

func verifyBrokerList(r *test.KnRunResultCollector, brokers ...string) {
	out := r.KnTest().Kn().Run("broker", "list")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, brokers...))
}

func verifyBrokerListOutputName(r *test.KnRunResultCollector, broker ...string) {
	out := r.KnTest().Kn().Run("broker", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, broker...))
}

func verifyBrokerDescribe(r *test.KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("broker", "describe", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, name, "Address:", "URL:", "Conditions:"))
}

func verifyBrokerNotfound(r *test.KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("broker", "describe", name)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, name, "not found"))
}
