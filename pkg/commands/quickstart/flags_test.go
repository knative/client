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

	"github.com/spf13/cobra"
)

func TestClusterNameOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	clusterNameOption(cmd, "default-name")

	flag := cmd.Flags().Lookup("name")
	if flag == nil {
		t.Error("Expected 'name' flag to be added")
	}

	if flag.Shorthand != "n" {
		t.Errorf("Expected shorthand 'n', got '%s'", flag.Shorthand)
	}

	if flag.DefValue != "default-name" {
		t.Errorf("Expected default value 'default-name', got '%s'", flag.DefValue)
	}
}

func TestKubernetesVersionOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	kubernetesVersionOption(cmd, "1.25.0", "test usage")

	flag := cmd.Flags().Lookup("kubernetes-version")
	if flag == nil {
		t.Error("Expected 'kubernetes-version' flag to be added")
	}

	if flag.Shorthand != "k" {
		t.Errorf("Expected shorthand 'k', got '%s'", flag.Shorthand)
	}

	if flag.DefValue != "1.25.0" {
		t.Errorf("Expected default value '1.25.0', got '%s'", flag.DefValue)
	}

	if flag.Usage != "test usage" {
		t.Errorf("Expected usage 'test usage', got '%s'", flag.Usage)
	}
}

func TestInstallServingOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	installServingOption(cmd)

	flag := cmd.Flags().Lookup("install-serving")
	if flag == nil {
		t.Error("Expected 'install-serving' flag to be added")
	}

	if flag.DefValue != "false" {
		t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
	}

	if flag.Usage != "install Serving on quickstart cluster" {
		t.Errorf("Expected usage text, got '%s'", flag.Usage)
	}
}

func TestInstallEventingOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	installEventingOption(cmd)

	flag := cmd.Flags().Lookup("install-eventing")
	if flag == nil {
		t.Error("Expected 'install-eventing' flag to be added")
	}

	if flag.DefValue != "false" {
		t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
	}

	if flag.Usage != "install Eventing on quickstart cluster" {
		t.Errorf("Expected usage text, got '%s'", flag.Usage)
	}
}

func TestInstallKindRegistryOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	installKindRegistryOption(cmd)

	flag := cmd.Flags().Lookup("registry")
	if flag == nil {
		t.Error("Expected 'registry' flag to be added")
	}

	if flag.DefValue != "false" {
		t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
	}

	if flag.Usage != "install registry for Kind quickstart cluster" {
		t.Errorf("Expected usage text, got '%s'", flag.Usage)
	}
}

func TestInstallKindExtraMountHostPathOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	installKindExtraMountHostPathOption(cmd)

	flag := cmd.Flags().Lookup("extraMountHostPath")
	if flag == nil {
		t.Error("Expected 'extraMountHostPath' flag to be added")
	}

	if flag.DefValue != "" {
		t.Errorf("Expected default value '', got '%s'", flag.DefValue)
	}

	if flag.Usage != "set the extraMount hostPath on Kind quickstart cluster" {
		t.Errorf("Expected usage text, got '%s'", flag.Usage)
	}
}

func TestInstallKindExtraMountContainerPathOption(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	installKindExtraMountContainerPathOption(cmd)

	flag := cmd.Flags().Lookup("extraMountContainerPath")
	if flag == nil {
		t.Error("Expected 'extraMountContainerPath' flag to be added")
	}

	if flag.DefValue != "" {
		t.Errorf("Expected default value '', got '%s'", flag.DefValue)
	}

	if flag.Usage != "set the extraMount containerPath on Kind quickstart cluster" {
		t.Errorf("Expected usage text, got '%s'", flag.Usage)
	}
}
