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

func TestNewMinikubeCommand(t *testing.T) {
	p := &commands.KnParams{}
	cmd := NewMinikubeCommand(p)

	if cmd == nil {
		t.Fatal("NewMinikubeCommand() returned nil")
	}

	if cmd.Use != "minikube" {
		t.Errorf("Expected Use to be 'minikube', got '%s'", cmd.Use)
	}

	if cmd.Short != "Quickstart with Minikube" {
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
	}

	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be present", flagName)
		}
	}

	// Check that Kind-specific flags are NOT present
	kindSpecificFlags := []string{
		"registry",
		"extraMountHostPath",
		"extraMountContainerPath",
	}

	for _, flagName := range kindSpecificFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag != nil {
			t.Errorf("Expected Kind-specific flag '%s' to NOT be present in minikube command", flagName)
		}
	}

	// Check specific flag defaults for minikube command
	nameFlag := cmd.Flags().Lookup("name")
	if nameFlag != nil && nameFlag.DefValue != "knative" {
		t.Errorf("Expected name flag default to be 'knative', got '%s'", nameFlag.DefValue)
	}

	kubernetesVersionFlag := cmd.Flags().Lookup("kubernetes-version")
	if kubernetesVersionFlag != nil && kubernetesVersionFlag.Usage != "kubernetes version to use (1.x.y)" {
		t.Errorf("Expected kubernetes-version flag usage text, got '%s'", kubernetesVersionFlag.Usage)
	}
}
