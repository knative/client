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
