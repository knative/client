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

	"knative.dev/client/lib/test/integration"
	"knative.dev/client/pkg/util"
)

func TestRoute(t *testing.T) {
	t.Parallel()
	it, err := integration.NewIntegrationTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := integration.NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("create hello service and return no error")
	serviceCreate(t, it, r, "hello")

	t.Log("return a list of routes")
	routeList(t, it, r)

	t.Log("return a list of routes associated with hello service")
	routeListWithArgument(t, it, r, "hello")

	t.Log("return a list of routes associated with hello service with print flags")
	routeListWithPrintFlags(t, it, r, "hello")

	t.Log("describe route from hello service")
	routeDescribe(t, it, r, "hello")

	t.Log("describe route from hello service with print flags")
	routeDescribeWithPrintFlags(t, it, r, "hello")

	t.Log("delete hello service and return no error")
	serviceDelete(t, it, r, "hello")
}

func routeList(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector) {
	out := it.Kn().Run("route", "list")

	expectedHeaders := []string{"NAME", "URL", "READY"}
	assert.Check(t, util.ContainsAll(out.Stdout, expectedHeaders...))
	r.AssertNoError(out)
}

func routeListWithArgument(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, routeName string) {
	out := it.Kn().Run("route", "list", routeName)

	assert.Check(t, util.ContainsAll(out.Stdout, routeName))
	r.AssertNoError(out)
}

func routeDescribe(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, routeName string) {
	out := it.Kn().Run("route", "describe", routeName)

	assert.Check(t, util.ContainsAll(out.Stdout,
		routeName, it.Kn().Namespace(), "URL", "Service", "Traffic", "Targets", "Conditions"))
	r.AssertNoError(out)
}

func routeDescribeWithPrintFlags(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, routeName string) {
	out := it.Kn().Run("route", "describe", routeName, "-o=name")

	expectedName := fmt.Sprintf("route.serving.knative.dev/%s", routeName)
	assert.Equal(t, strings.TrimSpace(out.Stdout), expectedName)
	r.AssertNoError(out)
}

func routeListWithPrintFlags(t *testing.T, it *integration.Test, r *integration.KnRunResultCollector, names ...string) {
	out := it.Kn().Run("route", "list", "-o=jsonpath={.items[*].metadata.name}")
	assert.Check(t, util.ContainsAll(out.Stdout, names...))
	r.AssertNoError(out)
}
