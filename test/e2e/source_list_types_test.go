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
// +build !serving

package e2e

import (
	"testing"

	"gotest.tools/assert"
	"knative.dev/client/pkg/util"
)

func TestSourceListTypes(t *testing.T) {
	t.Parallel()
	test, err := NewE2eTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, test.Teardown())
	}()

	r := NewKnRunResultCollector(t)
	defer r.DumpIfFailed()

	t.Log("List available source types")
	output := test.sourceListTypes(t, r)
	assert.Check(t, util.ContainsAll(output, "TYPE", "NAME", "DESCRIPTION", "Ping", "ApiServer"))

	t.Log("List available source types in YAML format")

	output = test.sourceListTypes(t, r, "-oyaml")
	assert.Check(t, util.ContainsAll(output, "apiextensions.k8s.io/v1beta1", "CustomResourceDefinition", "Ping", "ApiServer"))
}

func (test *e2eTest) sourceListTypes(t *testing.T, r *KnRunResultCollector, args ...string) string {
	cmd := append([]string{"source", "list-types"}, args...)
	out := test.kn.Run(cmd...)
	r.AssertNoError(out)
	return out.Stdout
}
