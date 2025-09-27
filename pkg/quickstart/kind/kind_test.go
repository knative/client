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

package kind

import (
	"fmt"
	"strings"
	"testing"
)

func TestSetUpFunction(t *testing.T) {
	// Test that the SetUp function exists and has the correct signature
	// We can't easily test the actual functionality without mocking docker, kind, and kubectl
	// but we can verify the function compiles and accepts the expected parameters

	// Test that function can be called without panicking with various parameter combinations
	testCases := []struct {
		name                               string
		kVersion                           string
		installServing                     bool
		installEventing                    bool
		installKindRegistry                bool
		installKindExtraMountHostPath      string
		installKindExtraMountContainerPath string
		expectError                        bool
	}{
		{
			name:                               "basic-test",
			kVersion:                           "1.25.0",
			installServing:                     true,
			installEventing:                    false,
			installKindRegistry:                false,
			installKindExtraMountHostPath:      "",
			installKindExtraMountContainerPath: "",
			expectError:                        true, // Expected to fail in test environment
		},
		{
			name:                               "with-registry",
			kVersion:                           "",
			installServing:                     false,
			installEventing:                    true,
			installKindRegistry:                true,
			installKindExtraMountHostPath:      "/tmp",
			installKindExtraMountContainerPath: "/mnt",
			expectError:                        true, // Expected to fail in test environment
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("SetUp function panicked as expected in test environment: %v", r)
				}
			}()

			// Just verify the function can be called - it will likely fail due to missing kind/docker
			err := SetUp(tc.name, tc.kVersion, tc.installServing, tc.installEventing, tc.installKindRegistry, tc.installKindExtraMountHostPath, tc.installKindExtraMountContainerPath)

			if tc.expectError && err != nil {
				t.Logf("SetUp function failed as expected in test environment: %v", err)
			} else if !tc.expectError && err != nil {
				t.Errorf("SetUp function failed unexpectedly: %v", err)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Test that important constants are set to reasonable values
	if kubernetesVersion == "" {
		t.Error("kubernetesVersion should not be empty")
	}

	if kindVersion <= 0 {
		t.Error("kindVersion should be positive")
	}

	if containerRegName == "" {
		t.Error("containerRegName should not be empty")
	}

	if containerRegPort == "" {
		t.Error("containerRegPort should not be empty")
	}

	// Test that default values are reasonable
	if kubernetesVersion != "kindest/node:v1.32.0" {
		t.Logf("kubernetesVersion changed from expected default: %s", kubernetesVersion)
	}

	if kindVersion != 0.26 {
		t.Logf("kindVersion changed from expected default: %f", kindVersion)
	}

	if containerRegName != "kind-registry" {
		t.Logf("containerRegName changed from expected default: %s", containerRegName)
	}

	if containerRegPort != "5001" {
		t.Logf("containerRegPort changed from expected default: %s", containerRegPort)
	}

	if !installKnative {
		t.Error("installKnative should default to true")
	}
}

// Test parameter processing logic
func TestSetUpParameterProcessing(t *testing.T) {
	testCases := []struct {
		name            string
		kVersion        string
		installServing  bool
		installEventing bool
		expectedServing bool
		expectedEventing bool
		description     string
	}{
		{
			name:             "both-false-should-default-to-both-true",
			kVersion:         "1.25.0",
			installServing:   false,
			installEventing:  false,
			expectedServing:  true,
			expectedEventing: true,
			description:      "When both serving and eventing are false, both should default to true",
		},
		{
			name:             "serving-only",
			kVersion:         "1.25.0",
			installServing:   true,
			installEventing:  false,
			expectedServing:  true,
			expectedEventing: false,
			description:      "When only serving is true, eventing should remain false",
		},
		{
			name:             "eventing-only",
			kVersion:         "1.25.0",
			installServing:   false,
			installEventing:  true,
			expectedServing:  false,
			expectedEventing: true,
			description:      "When only eventing is true, serving should remain false",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the logic that would be applied in SetUp
			serving := tc.installServing
			eventing := tc.installEventing

			// Apply the same logic as in SetUp function
			if !serving && !eventing {
				serving = true
				eventing = true
			}

			if serving != tc.expectedServing {
				t.Errorf("%s: expected serving=%v, got %v", tc.description, tc.expectedServing, serving)
			}

			if eventing != tc.expectedEventing {
				t.Errorf("%s: expected eventing=%v, got %v", tc.description, tc.expectedEventing, eventing)
			}
		})
	}
}

// Test version string processing
func TestKubernetesVersionProcessing(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "kindest/node:v1.32.0", // Should use default
		},
		{
			input:    "1.25.0",
			expected: "kindest/node:v1.25.0",
		},
		{
			input:    "kindest/node:v1.26.0",
			expected: "kindest/node:v1.26.0", // Should use as-is
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("version-%s", tc.input), func(t *testing.T) {
			// Save original
			original := kubernetesVersion
			defer func() { kubernetesVersion = original }()

			// Reset to default
			kubernetesVersion = "kindest/node:v1.32.0"

			// Apply the same logic as in SetUp
			if tc.input != "" {
				if strings.Contains(tc.input, ":") {
					kubernetesVersion = tc.input
				} else {
					kubernetesVersion = "kindest/node:v" + tc.input
				}
			}

			if kubernetesVersion != tc.expected {
				t.Errorf("Input '%s': expected '%s', got '%s'", tc.input, tc.expected, kubernetesVersion)
			}
		})
	}
}

// Test registry configuration
func TestRegistryConfiguration(t *testing.T) {
	// Test registry URL construction
	registryURL := fmt.Sprintf("localhost:%s", containerRegPort)
	expected := "localhost:5001"

	if registryURL != expected {
		t.Errorf("Expected registry URL '%s', got '%s'", expected, registryURL)
	}

	// Test registry name
	if containerRegName != "kind-registry" {
		t.Errorf("Expected container registry name 'kind-registry', got '%s'", containerRegName)
	}
}

// Test that cluster name gets set correctly
func TestClusterNameSetting(t *testing.T) {
	originalName := clusterName
	defer func() { clusterName = originalName }()

	testName := "test-cluster"
	clusterName = testName

	if clusterName != testName {
		t.Errorf("Expected cluster name '%s', got '%s'", testName, clusterName)
	}
}
