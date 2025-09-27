// Copyright © 2025 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package quickstart

import (
	"testing"

	"knative.dev/client/pkg/commands"
)

func TestNewKindCommand(t *testing.T) {
	p := &commands.KnParams{}
	cmd := NewKindCommand(p)

	if cmd == nil {
		t.Fatal("NewKindCommand() returned nil")
	}

	if cmd.Use != "kind" {
		t.Errorf("Expected Use to be 'kind', got '%s'", cmd.Use)
	}

	if cmd.Short != "Quickstart with Kind" {
		t.Errorf("Expected Short description, got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Check that expected flags are present
	expectedFlags := []string{
		"name",
		"kubernetes-version",
		"install-serving",
		"install-eventing",
		"registry",
		"extraMountHostPath",
		"extraMountContainerPath",
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be present", flagName)
		}
	}

	// Check specific flag defaults for kind command
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag != nil && nameFlag.DefValue != "knative" {
		t.Errorf("Expected name flag default to be 'knative', got '%s'", nameFlag.DefValue)
	}

	kubernetesVersionFlag := cmd.Flags().Lookup("kubernetes-version")
	if kubernetesVersionFlag != nil && kubernetesVersionFlag.Usage != "kubernetes version to use (1.x.y) or (kindest/node:v1.x.y)" {
		t.Errorf("Expected kubernetes-version flag usage text, got '%s'", kubernetesVersionFlag.Usage)
	}
}