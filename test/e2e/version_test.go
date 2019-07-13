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
	"testing"

	"github.com/knative/client/pkg/util"
	"gotest.tools/assert"
)

func TestVersion(t *testing.T) {
	env := buildEnv(t)
	kn := kn{t, env.Namespace, Logger{}}

	out, _ := kn.RunWithOpts([]string{"version"}, runOpts{NoNamespace: true})

	assert.Check(t, util.ContainsAll(out, "Version"))
}
