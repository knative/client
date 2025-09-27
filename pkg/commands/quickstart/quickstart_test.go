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

func TestNewQuickstartCommand(t *testing.T) {
	p := &commands.KnParams{}
	cmd := NewQuickstartCommand(p)

	if cmd == nil {
		t.Fatal("NewQuickstartCommand() returned nil")
	}

	if cmd.Use != "quickstart" {
		t.Errorf("Expected Use to be 'quickstart', got '%s'", cmd.Use)
	}

	if cmd.Short != "Get started quickly with Knative" {
		t.Errorf("Expected Short description, got '%s'", cmd.Short)
	}

	if cmd.Long != "Get up and running with a local Knative environment" {
		t.Errorf("Expected Long description, got '%s'", cmd.Long)
	}

	// Check that subcommands are added
	expectedSubcommands := []string{"kind", "minikube", "version"}
	actualSubcommands := make([]string, 0)
	for _, subCmd := range cmd.Commands() {
		actualSubcommands = append(actualSubcommands, subCmd.Use)
	}

	if len(actualSubcommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(actualSubcommands))
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, actual := range actualSubcommands {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", expected)
		}
	}
}
