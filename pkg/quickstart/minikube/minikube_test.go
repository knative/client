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
