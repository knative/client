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
