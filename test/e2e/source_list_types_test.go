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
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

func TestSourceListTypes(t *testing.T) {
	t.Parallel()
	test := NewE2eTest(t)
	test.Setup(t)
	defer test.Teardown(t)

	t.Run("List available source types", func(t *testing.T) {
		output := test.sourceListTypes(t)
		assert.Check(t, util.ContainsAll(output, "TYPE", "NAME", "DESCRIPTION", "CronJob", "ApiServer"))
	})

	t.Run("List available source types in YAML format", func(t *testing.T) {
		output := test.sourceListTypes(t, "-oyaml")
		assert.Check(t, util.ContainsAll(output, "apiextensions.k8s.io/v1beta1", "CustomResourceDefinition", "CronJob", "ApiServer"))

	})
}

func (test *e2eTest) sourceListTypes(t *testing.T, args ...string) (out string) {
	cmd := append([]string{"source", "list-types"}, args...)
	out, err := test.kn.RunWithOpts(cmd, runOpts{})
	assert.NilError(t, err)
	return
}
