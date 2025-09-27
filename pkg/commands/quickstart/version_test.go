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
	"bytes"
	"strings"
	"testing"

	"knative.dev/client/pkg/commands"
)

func TestNewVersionCommand(t *testing.T) {
	p := &commands.KnParams{}
	cmd := NewVersionCommand(p)

	if cmd == nil {
		t.Fatal("NewVersionCommand() returned nil")
	}

	if cmd.Use != "version" {
		t.Errorf("Expected Use to be 'version', got '%s'", cmd.Use)
	}

	if cmd.Short != "Prints the quickstart version" {
		t.Errorf("Expected Short description, got '%s'", cmd.Short)
	}

	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}
}

func TestVersionCommandOutput(t *testing.T) {
	// Set test values
	Version = "test-version"
	BuildDate = "test-date"
	GitRevision = "test-revision"

	p := &commands.KnParams{}
	cmd := NewVersionCommand(p)

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Execute command
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	output := buf.String()

	// Check that all version information is included in output
	expectedStrings := []string{
		"Version:      test-version",
		"Build Date:   test-date",
		"Git Revision: test-revision",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
		}
	}
}
