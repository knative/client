// Copyright 2019 The knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"testing"
)

type kn struct {
	t         *testing.T
	namespace string
	l         Logger
}

// Run the 'kn' CLI with args
func (k kn) Run(args []string) string {
	out, _ := k.RunWithOpts(args, runOpts{})
	return out
}

// Run the 'kn' CLI with args and opts
func (k kn) RunWithOpts(args []string, opts runOpts) (string, error) {
	if !opts.NoNamespace {
		args = append(args, []string{"-n", k.namespace}...)
	}

	return runCLIWithOpts("kn", args, opts, k.l)
}
