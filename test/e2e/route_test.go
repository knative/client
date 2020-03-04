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
	"fmt"
	"strings"
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

func TestRoute(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("create hello service and return no error")
	test.serviceCreate(t, r, "hello")

	t.Log("return a list of routes")
	test.routeList(t, r)

	t.Log("return a list of routes associated with hello service")
	test.routeListWithArgument(t, r, "hello")

	t.Log("return a list of routes associated with hello service with print flags")
	test.routeListWithPrintFlags(t, r, "hello")

	t.Log("describe route from hello service")
	test.routeDescribe(t, r, "hello")

	t.Log("describe route from hello service with print flags")
	test.routeDescribeWithPrintFlags(t, r, "hello")

	t.Log("delete hello service and return no error")
	test.serviceDelete(t, r, "hello")
}

func (test *e2eTest) routeList(t *testing.T, r *KnRunResultCollector) {
	out := test.kn.Run("route", "list")

	expectedHeaders := []string{"NAME", "URL", "READY"}
	assert.Check(t, util.ContainsAll(out.Stdout, expectedHeaders...))
	r.AssertNoError(out)
}

func (test *e2eTest) routeListWithArgument(t *testing.T, r *KnRunResultCollector, routeName string) {
	out := test.kn.Run("route", "list", routeName)

	assert.Check(t, util.ContainsAll(out.Stdout, routeName))
	r.AssertNoError(out)
}

func (test *e2eTest) routeDescribe(t *testing.T, r *KnRunResultCollector, routeName string) {
	out := test.kn.Run("route", "describe", routeName)

	assert.Check(t, util.ContainsAll(out.Stdout,
		routeName, test.kn.namespace, "URL", "Service", "Traffic", "Targets", "Conditions"))
	r.AssertNoError(out)
}

func (test *e2eTest) routeDescribeWithPrintFlags(t *testing.T, r *KnRunResultCollector, routeName string) {
	out := test.kn.Run("route", "describe", routeName, "-o=name")

	expectedName := fmt.Sprintf("route.serving.knative.dev/%s", routeName)
	assert.Equal(t, strings.TrimSpace(out.Stdout), expectedName)
	r.AssertNoError(out)
}

func (test *e2eTest) routeListWithPrintFlags(t *testing.T, r *KnRunResultCollector, names ...string) {
	out := test.kn.Run("route", "list", "-o=jsonpath={.items[*].metadata.name}")
	assert.Check(t, util.ContainsAll(out.Stdout, names...))
	r.AssertNoError(out)
}
