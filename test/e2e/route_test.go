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

	"gotest.tools/v3/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

func TestRoute(t *testing.T) {
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

	t.Log("return a list of routes")
	routeList(r)

	t.Log("return a list of routes associated with hello service")
	routeListWithArgument(r, "hello")

	t.Log("return a list of routes associated with hello service with -oname flag")
	routeListOutputName(r, "hello")

	t.Log("return a list of routes associated with hello service with print flags")
	routeListWithPrintFlags(r, "hello")

	t.Log("describe route from hello service")
	routeDescribe(r, "hello")

	t.Log("describe route from hello service with print flags")
	routeDescribeWithPrintFlags(r, "hello")

	t.Log("delete hello service and return no error")
	test.ServiceDelete(r, "hello")
}

func routeList(r *test.KnRunResultCollector) {
	out := r.KnTest().Kn().Run("route", "list")

	expectedHeaders := []string{"NAME", "URL", "READY"}
	assert.Check(r.T(), util.ContainsAll(out.Stdout, expectedHeaders...))
	r.AssertNoError(out)
}

func routeListOutputName(r *test.KnRunResultCollector, routeName string) {
	out := r.KnTest().Kn().Run("route", "list", "--output", "name")
	r.AssertNoError(out)
	assert.Check(r.T(), util.ContainsAll(out.Stdout, routeName, "route.serving.knative.dev"))
}

func routeListWithArgument(r *test.KnRunResultCollector, routeName string) {
	out := r.KnTest().Kn().Run("route", "list", routeName)

	assert.Check(r.T(), util.ContainsAll(out.Stdout, routeName))
	r.AssertNoError(out)
}

func routeDescribe(r *test.KnRunResultCollector, routeName string) {
	out := r.KnTest().Kn().Run("route", "describe", routeName)

	assert.Check(r.T(), util.ContainsAll(out.Stdout,
		routeName, r.KnTest().Kn().Namespace(), "URL", "Service", "Traffic", "Targets", "Conditions"))
	r.AssertNoError(out)
}

func routeDescribeWithPrintFlags(r *test.KnRunResultCollector, routeName string) {
	out := r.KnTest().Kn().Run("route", "describe", routeName, "-o=name")

	expectedName := fmt.Sprintf("route.serving.knative.dev/%s", routeName)
	assert.Equal(r.T(), strings.TrimSpace(out.Stdout), expectedName)
	r.AssertNoError(out)
}

func routeListWithPrintFlags(r *test.KnRunResultCollector, names ...string) {
	out := r.KnTest().Kn().Run("route", "list", "-o=jsonpath={.items[*].metadata.name}")
	assert.Check(r.T(), util.ContainsAll(out.Stdout, names...))
	r.AssertNoError(out)
}
