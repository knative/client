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
// +build !eventing

package e2e

import (
	"testing"

	"gotest.tools/v3/assert"

	"knative.dev/client/lib/test"
)

func TestRevision(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("create hello service and return no error")
	test.ServiceCreate(r, "hello")

	t.Log("describe revision from hello service with print flags")
	revName := test.FindRevision(r, "hello")
	test.RevisionListOutputName(r, revName)
	test.RevisionDescribeWithPrintFlags(r, revName)

	t.Log("update hello service and increase revision count to 2")
	test.ServiceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "8888")

	t.Log("show a list of revisions sorted by the count of configuration generation")
	test.RevisionListWithService(r, "hello")

	t.Log("update hello service and increase revision count to 3")
	test.ServiceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "9000")

	t.Log("delete three revisions with one revision a nonexistent")
	existRevision1 := test.FindRevisionByGeneration(r, "hello", 1)
	existRevision2 := test.FindRevisionByGeneration(r, "hello", 2)
	nonexistRevision := "hello-nonexist"
	test.RevisionMultipleDelete(r, existRevision1, existRevision2, nonexistRevision)

	t.Log("update hello service and increase revision count to 4")
	test.ServiceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "8888")
	t.Log("delete all unreferenced revision from hello service and return no error")
	unRefRevision := test.FindRevisionByGeneration(r, "hello", 3)
	test.RevisionDeleteWithPruneOption(r, "hello", unRefRevision)

	t.Log("update hello service and increase revision count to 5")
	test.ServiceUpdate(r, "hello", "--env", "TARGET=kn", "--port", "9000")
	t.Log("create hello service and return no error")
	test.ServiceCreate(r, "hello1")
	t.Log("update hello1 service and increase revision count to 2")
	test.ServiceUpdate(r, "hello1", "--env", "TARGET=kn", "--port", "8888")
	t.Log("delete all unreferenced revisions return no error")
	unRefRevision1 := test.FindRevisionByGeneration(r, "hello", 4)
	unRefRevision2 := test.FindRevisionByGeneration(r, "hello1", 1)
	test.RevisionDeleteWithPruneAllOption(r, unRefRevision1, unRefRevision2)
	t.Log("delete hello1 service and return no error")
	test.ServiceDelete(r, "hello1")

	t.Log("delete latest revision from hello service and return no error")
	revName = test.FindRevision(r, "hello")
	test.RevisionDelete(r, revName)

	t.Log("delete hello service and return no error")
	test.ServiceDelete(r, "hello")
}
