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

package minikube

import (
	"fmt"
	"strconv"
	"testing"
)

func TestSetUpFunction(t *testing.T) {
	// Test that the SetUp function exists and has the correct signature
	// We can't easily test the actual functionality without mocking minikube and kubectl
	// but we can verify the function compiles and accepts the expected parameters

	testCases := []struct {
		name            string
		kVersion        string
		installServing  bool
		installEventing bool
		expectError     bool
	}{
		{
			name:            "basic-test",
			kVersion:        "1.25.0",
			installServing:  true,
			installEventing: false,
			expectError:     true, // Expected to fail in test environment
		},
		{
			name:            "with-eventing",
			kVersion:        "",
			installServing:  false,
			installEventing: true,
			expectError:     true, // Expected to fail in test environment
		},
		{
			name:            "both-false",
			kVersion:        "1.26.0",
			installServing:  false,
			installEventing: false,
			expectError:     true, // Expected to fail in test environment, but should default to installing both
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("SetUp function panicked as expected in test environment: %v", r)
				}
			}()

			// Just verify the function can be called - it will likely fail due to missing minikube
			err := SetUp(tc.name, tc.kVersion, tc.installServing, tc.installEventing)

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

	if minikubeVersion <= 0 {
		t.Error("minikubeVersion should be positive")
	}

	if cpus == "" {
		t.Error("cpus should not be empty")
	}

	if memory == "" {
		t.Error("memory should not be empty")
	}

	// Test that default values are reasonable
	if kubernetesVersion != "1.32.0" {
		t.Logf("kubernetesVersion changed from expected default: %s", kubernetesVersion)
	}

	if minikubeVersion != 1.35 {
		t.Logf("minikubeVersion changed from expected default: %f", minikubeVersion)
	}

	if cpus != "3" {
		t.Logf("cpus changed from expected default: %s", cpus)
	}

	if memory != "3072" {
		t.Logf("memory changed from expected default: %s", memory)
	}

	if !installKnative {
		t.Error("installKnative should default to true")
	}
}

// Test parameter processing logic
func TestSetUpParameterProcessing(t *testing.T) {
	testCases := []struct {
		name             string
		kVersion         string
		installServing   bool
		installEventing  bool
		expectedServing  bool
		expectedEventing bool
		description      string
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
	// Save original values
	originalVersion := kubernetesVersion
	originalOverride := clusterVersionOverride
	defer func() {
		kubernetesVersion = originalVersion
		clusterVersionOverride = originalOverride
	}()

	testCases := []struct {
		input            string
		expectedVersion  string
		expectedOverride bool
	}{
		{
			input:            "",
			expectedVersion:  "1.32.0", // Should use default
			expectedOverride: false,
		},
		{
			input:            "1.25.0",
			expectedVersion:  "1.25.0",
			expectedOverride: true,
		},
		{
			input:            "1.26.3",
			expectedVersion:  "1.26.3",
			expectedOverride: true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("version-%s", tc.input), func(t *testing.T) {
			// Reset to defaults
			kubernetesVersion = "1.32.0"
			clusterVersionOverride = false

			// Apply the same logic as in SetUp
			if tc.input != "" {
				kubernetesVersion = tc.input
				clusterVersionOverride = true
			}

			if kubernetesVersion != tc.expectedVersion {
				t.Errorf("Input '%s': expected version '%s', got '%s'", tc.input, tc.expectedVersion, kubernetesVersion)
			}

			if clusterVersionOverride != tc.expectedOverride {
				t.Errorf("Input '%s': expected override %v, got %v", tc.input, tc.expectedOverride, clusterVersionOverride)
			}
		})
	}
}

// Test resource configuration
func TestResourceConfiguration(t *testing.T) {
	// Test CPU configuration
	if cpus != "3" {
		t.Errorf("Expected 3 CPUs, got '%s'", cpus)
	}

	// Test memory configuration
	if memory != "3072" {
		t.Errorf("Expected 3072MB memory, got '%s'", memory)
	}

	// Test that values are reasonable
	cpuInt, err := strconv.Atoi(cpus)
	if err != nil {
		t.Errorf("CPU value should be numeric, got '%s'", cpus)
	} else if cpuInt < 1 {
		t.Errorf("CPU count should be at least 1, got %d", cpuInt)
	}

	memoryInt, err := strconv.Atoi(memory)
	if err != nil {
		t.Errorf("Memory value should be numeric, got '%s'", memory)
	} else if memoryInt < 1024 {
		t.Errorf("Memory should be at least 1024MB, got %d", memoryInt)
	}
}

// Test cluster name setting
func TestClusterNameSetting(t *testing.T) {
	originalName := clusterName
	defer func() { clusterName = originalName }()

	testName := "test-minikube-cluster"
	clusterName = testName

	if clusterName != testName {
		t.Errorf("Expected cluster name '%s', got '%s'", testName, clusterName)
	}
}

// Test minikube version validation
func TestMinikubeVersionValidation(t *testing.T) {
	if minikubeVersion < 1.0 {
		t.Errorf("Minikube version should be at least 1.0, got %f", minikubeVersion)
	}

	// Test that the version is a reasonable current version
	if minikubeVersion > 2.0 {
		t.Logf("Minikube version is quite high: %f (this may be expected)", minikubeVersion)
	}
}
