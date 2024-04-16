// Copyright © 2022 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
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
	"knative.dev/client-pkg/pkg/util"
	"knative.dev/client-pkg/pkg/util/test"
)

const (
	testName      = "test-eventtype"
	testName1     = "test-eventtype-1"
	testName2     = "test-eventtype-2"
	testName3     = "test-eventtype-3"
	testType      = "test.type"
	testBroker    = "test-broker"
	testSource    = "test.source.com"
	testSourceBad = "test.source.com\b"
)

func TestEventtype(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("create eventtype, list, describe, and delete it")
	test.EventtypeCreate(r, testName, testType)
	test.EventtypeList(r, testName)
	test.EventtypeDescribe(r, testName)
	test.EventtypeDelete(r, testName)
	verifyEventtypeNotfound(r, testName)

	t.Log("create eventtype with broker and source")
	test.EventtypeCreateWithBrokerSource(r, testName, testType, testBroker, testSource)
	test.EventtypeList(r, testName)

	t.Log("create multiple eventtypes and list them")
	test.EventtypeCreate(r, testName1, testType)
	test.EventtypeCreate(r, testName2, testType)
	test.EventtypeCreate(r, testName3, testType)
	test.EventtypeList(r, testName1, testName2, testName3)

	t.Log("create eventtype with invalid source")
	test.EventtypeCreateWithSourceError(r, testName, testType, testSourceBad)
}

func verifyEventtypeNotfound(r *test.KnRunResultCollector, name string) {
	out := r.KnTest().Kn().Run("eventtype", "describe", name)
	r.AssertError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stderr, name, "not found"))
}
