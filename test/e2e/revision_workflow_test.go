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
	"strings"
	"testing"

	"github.com/knative/client/pkg/util"
	"gotest.tools/assert"
)

func TestRevisionWorkflow(t *testing.T) {
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("create hello service and returns no error", func(t *testing.T) {
		test.serviceCreate(t, "hello")
	})

	t.Run("delete latest revision from hello service and returns no error", func(t *testing.T) {
		test.deleteRevision(t, "hello")
	})

	t.Run("delete hello service and returns no error", func(t *testing.T) {
		test.serviceDelete(t, "hello")
	})
}

// Private

func (test *e2eTest) deleteRevision(t *testing.T, serviceName string) {
	revName, err := test.kn.RunWithOpts([]string{"revision", "list", "-o=jsonpath={.items[0].metadata.name}"}, runOpts{})
	assert.NilError(t, err)
	if strings.Contains(revName, "No resources found.") {
		t.Errorf("Could not find revision name.")
	}

	out, err := test.kn.RunWithOpts([]string{"revision", "delete", revName}, runOpts{})
	assert.NilError(t, err)

	assert.Check(t, util.ContainsAll(out, "Revision", revName, "deleted", "namespace", test.kn.namespace))
}
