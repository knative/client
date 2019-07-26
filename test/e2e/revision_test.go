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

package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/knative/client/pkg/util"
	"gotest.tools/assert"
)

func TestRevision(t *testing.T) {
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("create hello service and return no error", func(t *testing.T) {
		test.serviceCreate(t, "hello")
	})

	t.Run("describe revision from hello service with print flags", func(t *testing.T) {
		test.revisionDescribeWithPrintFlags(t, "hello")
	})

	t.Run("delete latest revision from hello service and return no error", func(t *testing.T) {
		test.revisionDelete(t, "hello")
	})

	t.Run("delete hello service and return no error", func(t *testing.T) {
		test.serviceDelete(t, "hello")
	})
}

func (test *e2eTest) revisionDelete(t *testing.T, serviceName string) {
	revName := test.findRevision(t, serviceName)

	out, err := test.kn.RunWithOpts([]string{"revision", "delete", revName}, runOpts{})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Revision", revName, "deleted", "namespace", test.kn.namespace))
}

func (test *e2eTest) revisionDescribeWithPrintFlags(t *testing.T, serviceName string) {
	revName := test.findRevision(t, serviceName)

	out, err := test.kn.RunWithOpts([]string{"revision", "describe", revName, "-o=name"}, runOpts{})
	assert.NilError(t, err)

	expectedName := fmt.Sprintf("revision.serving.knative.dev/%s", revName)
	assert.Equal(t, strings.TrimSpace(out), expectedName)
}

func (test *e2eTest) findRevision(t *testing.T, serviceName string) string {
	revName, err := test.kn.RunWithOpts([]string{"revision", "list", "-o=jsonpath={.items[0].metadata.name}"}, runOpts{})
	assert.NilError(t, err)
	if strings.Contains(revName, "No resources found.") {
		t.Errorf("Could not find revision name.")
	}
	return revName
}
