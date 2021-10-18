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

//go:build e2e && !serving
// +build e2e,!serving

package e2e

import (
	"testing"

	"gotest.tools/v3/assert"

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
	test.BrokerCreate(r, "foo1")
	verifyBrokerList(r, "foo1")
	verifyBrokerListOutputName(r, "foo1")
	verifyBrokerDescribe(r, "foo1")
	test.BrokerDelete(r, "foo1", false)

	t.Log("create broker and delete it")
	test.BrokerCreate(r, "foo2")
	verifyBrokerList(r, "foo2")
	test.BrokerDelete(r, "foo2", true)
	verifyBrokerNotfound(r, "foo2")

	t.Log("create multiple brokers and list them")
	test.BrokerCreate(r, "foo3")
	test.BrokerCreate(r, "foo4")
	verifyBrokerList(r, "foo3", "foo4")
	verifyBrokerListOutputName(r, "foo3", "foo4")
	test.BrokerDelete(r, "foo3", true)
	test.BrokerDelete(r, "foo4", true)
	verifyBrokerNotfound(r, "foo3")
	verifyBrokerNotfound(r, "foo4")

	t.Log("create broker with class")
	test.BrokerCreateWithClass(r, "foo5", "foo-class")
	verifyBrokerList(r, "foo5")
	verifyBrokerListOutputName(r, "foo5")
	verifyBrokerDescribeContains(r, "foo5", "foo-class")
}

// Private functions

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

func verifyBrokerDescribeContains(r *test.KnRunResultCollector, name, str string) {
	out := r.KnTest().Kn().Run("broker", "describe", name)
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAllIgnoreCase(out.Stdout, name, str))
}
