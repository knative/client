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
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("create hello service and return no error", func(t *testing.T) {
		test.serviceCreate(t, "hello")
	})

	t.Run("return a list of routes", func(t *testing.T) {
		test.routeList(t)
	})

	t.Run("return a list of routes associated with hello service", func(t *testing.T) {
		test.routeListWithArgument(t, "hello")
	})

	t.Run("return a list of routes associated with hello service with print flags", func(t *testing.T) {
		test.routeListWithPrintFlags(t, "hello")
	})

	t.Run("describe route from hello service", func(t *testing.T) {
		test.routeDescribe(t, "hello")
	})

	t.Run("describe route from hello service with print flags", func(t *testing.T) {
		test.routeDescribeWithPrintFlags(t, "hello")
	})

	t.Run("delete hello service and return no error", func(t *testing.T) {
		test.serviceDelete(t, "hello")
	})
}

func (test *e2eTest) routeList(t *testing.T) {
	out, err := test.kn.RunWithOpts([]string{"route", "list"}, runOpts{})
	assert.NilError(t, err)

	expectedHeaders := []string{"NAME", "URL", "READY"}
	assert.Check(t, util.ContainsAll(out, expectedHeaders...))
}

func (test *e2eTest) routeListWithArgument(t *testing.T, routeName string) {
	out, err := test.kn.RunWithOpts([]string{"route", "list", routeName}, runOpts{})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, routeName))
}

func (test *e2eTest) routeDescribe(t *testing.T, routeName string) {
	out, err := test.kn.RunWithOpts([]string{"route", "describe", routeName}, runOpts{})
	assert.NilError(t, err)

	expectedGVK := `apiVersion: serving.knative.dev/v1alpha1
kind: Route`
	expectedNamespace := fmt.Sprintf("namespace: %s", test.kn.namespace)
	expectedServiceLabel := fmt.Sprintf("serving.knative.dev/service: %s", routeName)
	assert.Check(t, util.ContainsAll(out, expectedGVK, expectedNamespace, expectedServiceLabel))
}

func (test *e2eTest) routeDescribeWithPrintFlags(t *testing.T, routeName string) {
	out, err := test.kn.RunWithOpts([]string{"route", "describe", routeName, "-o=name"}, runOpts{})
	assert.NilError(t, err)

	expectedName := fmt.Sprintf("route.serving.knative.dev/%s", routeName)
	assert.Equal(t, strings.TrimSpace(out), expectedName)
}

func (test *e2eTest) routeListWithPrintFlags(t *testing.T, names ...string) {
	out, err := test.kn.RunWithOpts([]string{"route", "list", "-o=jsonpath={.items[*].metadata.name}"}, runOpts{})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, names...))
}
