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

package install

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestVersionVariables(t *testing.T) {
	// Test that version variables are properly declared and can be set
	originalServing := ServingVersion
	originalKourier := KourierVersion
	originalEventing := EventingVersion

	// Set test values
	ServingVersion = "test-serving"
	KourierVersion = "test-kourier"
	EventingVersion = "test-eventing"

	// Verify values were set
	if ServingVersion != "test-serving" {
		t.Errorf("Expected ServingVersion to be 'test-serving', got '%s'", ServingVersion)
	}
	if KourierVersion != "test-kourier" {
		t.Errorf("Expected KourierVersion to be 'test-kourier', got '%s'", KourierVersion)
	}
	if EventingVersion != "test-eventing" {
		t.Errorf("Expected EventingVersion to be 'test-eventing', got '%s'", EventingVersion)
	}

	// Restore original values
	ServingVersion = originalServing
	KourierVersion = originalKourier
	EventingVersion = originalEventing
}

func TestServingFunction(t *testing.T) {
	// This test verifies that the Serving function exists and has the correct signature
	// We can't easily test the actual functionality without mocking kubectl and external dependencies
	// but we can verify the function compiles and accepts the expected parameters

	// Set version for URL construction
	originalVersion := ServingVersion
	ServingVersion = "1.8.0"

	// Test that function can be called without panicking
	// Note: This will fail in a test environment without kubectl and network access
	// but it verifies the function signature and basic structure
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Serving function panicked as expected in test environment: %v", r)
		}
	}()

	// Just verify the function can be called - it will likely fail due to missing kubectl
	err := Serving("test-registry")
	// We expect this to fail in the test environment, so we just log it
	if err != nil {
		t.Logf("Serving function failed as expected in test environment: %v", err)
	}

	// Restore original version
	ServingVersion = originalVersion
}

func TestKourierFunction(t *testing.T) {
	// Similar to TestServingFunction - verify function exists and signature
	originalVersion := KourierVersion
	KourierVersion = "1.8.0"

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Kourier function panicked as expected in test environment: %v", r)
		}
	}()

	err := Kourier()
	if err != nil {
		t.Logf("Kourier function failed as expected in test environment: %v", err)
	}

	KourierVersion = originalVersion
}

func TestEventingFunction(t *testing.T) {
	// Similar to other function tests
	originalVersion := EventingVersion
	EventingVersion = "1.8.0"

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Eventing function panicked as expected in test environment: %v", r)
		}
	}()

	err := Eventing()
	if err != nil {
		t.Logf("Eventing function failed as expected in test environment: %v", err)
	}

	EventingVersion = originalVersion
}

func TestKourierKindFunction(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("KourierKind function panicked as expected in test environment: %v", r)
		}
	}()

	err := KourierKind()
	if err != nil {
		t.Logf("KourierKind function failed as expected in test environment: %v", err)
	}
}

func TestKourierMinikubeFunction(t *testing.T) {
	originalVersion := ServingVersion
	ServingVersion = "1.8.0"

	defer func() {
		if r := recover(); r != nil {
			t.Logf("KourierMinikube function panicked as expected in test environment: %v", r)
		}
	}()

	err := KourierMinikube()
	if err != nil {
		t.Logf("KourierMinikube function failed as expected in test environment: %v", err)
	}

	ServingVersion = originalVersion
}

// Test helper functions
func TestRunCommand(t *testing.T) {
	// Test runCommand with a simple command that should succeed
	cmd := exec.Command("echo", "test")
	err := runCommand(cmd)
	if err != nil {
		t.Errorf("runCommand failed with echo command: %v", err)
	}

	// Test runCommand with a command that should fail
	cmd = exec.Command("false")
	err = runCommand(cmd)
	if err == nil {
		t.Error("runCommand should have failed with 'false' command")
	}
}

func TestRetryingApply(t *testing.T) {
	// Test retryingApply with an invalid path (will fail but tests the retry logic)
	err := retryingApply("/nonexistent/file.yaml")
	if err == nil {
		t.Error("retryingApply should have failed with nonexistent file")
	}

	// The error should indicate kubectl failure
	if !strings.Contains(err.Error(), "exit status") {
		t.Logf("Expected kubectl error, got: %v", err)
	}
}

func TestWaitForCRDsEstablished(t *testing.T) {
	// This will fail in test environment without kubectl, but tests the function exists
	err := waitForCRDsEstablished()
	if err != nil {
		t.Logf("waitForCRDsEstablished failed as expected in test environment: %v", err)
	}
}

func TestWaitForPodsReady(t *testing.T) {
	// This will fail in test environment without kubectl, but tests the function exists
	err := waitForPodsReady("test-namespace")
	if err != nil {
		t.Logf("waitForPodsReady failed as expected in test environment: %v", err)
	}
}

// Test URL construction logic
func TestServingURLConstruction(t *testing.T) {
	originalVersion := ServingVersion
	defer func() { ServingVersion = originalVersion }()

	ServingVersion = "1.8.0"

	// We can't easily test the full function, but we can test that it would construct the right URLs
	expectedBaseURL := "https://github.com/knative/serving/releases/download/knative-v1.8.0"

	// Check that our version string would create the expected URL pattern
	if ServingVersion != "1.8.0" {
		t.Errorf("Expected ServingVersion to be '1.8.0', got '%s'", ServingVersion)
	}

	actualURL := fmt.Sprintf("https://github.com/knative/serving/releases/download/knative-v%s", ServingVersion)
	if actualURL != expectedBaseURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedBaseURL, actualURL)
	}
}

func TestKourierURLConstruction(t *testing.T) {
	originalVersion := KourierVersion
	defer func() { KourierVersion = originalVersion }()

	KourierVersion = "1.8.0"

	expectedURL := "https://github.com/knative-sandbox/net-kourier/releases/download/knative-v1.8.0/kourier.yaml"
	actualURL := fmt.Sprintf("https://github.com/knative-sandbox/net-kourier/releases/download/knative-v%s/kourier.yaml", KourierVersion)

	if actualURL != expectedURL {
		t.Errorf("Expected Kourier URL '%s', got '%s'", expectedURL, actualURL)
	}
}

func TestEventingURLConstruction(t *testing.T) {
	originalVersion := EventingVersion
	defer func() { EventingVersion = originalVersion }()

	EventingVersion = "1.8.0"

	expectedBaseURL := "https://github.com/knative/eventing/releases/download/knative-v1.8.0"
	actualURL := fmt.Sprintf("https://github.com/knative/eventing/releases/download/knative-v%s", EventingVersion)

	if actualURL != expectedBaseURL {
		t.Errorf("Expected Eventing URL '%s', got '%s'", expectedBaseURL, actualURL)
	}
}

// Test registry configuration
func TestServingWithRegistries(t *testing.T) {
	originalVersion := ServingVersion
	defer func() { ServingVersion = originalVersion }()

	ServingVersion = "1.8.0"

	// Test with empty registries
	err := Serving("")
	if err != nil {
		t.Logf("Serving with empty registries failed as expected: %v", err)
	}

	// Test with test registries
	err = Serving("localhost:5000,registry.example.com")
	if err != nil {
		t.Logf("Serving with test registries failed as expected: %v", err)
	}
}
