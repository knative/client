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

//go:build e2e && !serving && !project_admin
// +build e2e,!serving,!project_admin

package e2e

import (
	"testing"

	"gotest.tools/v3/assert"

	"knative.dev/client-pkg/pkg/util"
	"knative.dev/client-pkg/pkg/util/test"
)

// This test requires cluster admin permissions due to working with CustomResourceDefinitions.
// It can be excluded from test execution by using the project_admin build tag.
// See https://github.com/knative/client/issues/1385
func TestSourceListTypesCRD(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	t.Log("List available source types in YAML format")

	output := sourceListTypes(r, "-oyaml")
	assert.Check(t, util.ContainsAll(output, "apiextensions.k8s.io/v1", "CustomResourceDefinition", "Ping", "ApiServer"))
}
