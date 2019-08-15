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

package e2e

import (
	"os"
	"strings"
	"testing"
)

type env struct {
	Namespace string
}

const defaultKnE2ENamespace = "kne2etests"

var (
	namespaceCount = 0
	serviceCount   = 0
)

func buildEnv(t *testing.T) env {
	env := env{
		Namespace: os.Getenv("KN_E2E_NAMESPACE"),
	}
	env.validate(t)
	return env
}

func (e *env) validate(t *testing.T) {
	errStrs := []string{}

	if e.Namespace == "" {
		e.Namespace = defaultKnE2ENamespace
	}

	if len(errStrs) > 0 {
		t.Fatalf("%s", strings.Join(errStrs, "\n"))
	}
}
