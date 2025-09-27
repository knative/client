// Copyright © 2021 The Knative Authors
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
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/commands"
	"knative.dev/client/pkg/quickstart/kind"
)

// NewKindCommand implements 'kn quickstart kind' command
func NewKindCommand(p *commands.KnParams) *cobra.Command {
	var kindCmd = &cobra.Command{
		Use:   "kind",
		Short: "Quickstart with Kind",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running Knative Quickstart using Kind")
			return kind.SetUp(name, kubernetesVersion, installServing, installEventing, installKindRegistry, installKindExtraMountHostPath, installKindExtraMountContainerPath)
		},
	}
	// Set kindCmd options
	clusterNameOption(kindCmd, "knative")
	kubernetesVersionOption(kindCmd, "", "kubernetes version to use (1.x.y) or (kindest/node:v1.x.y)")
	installServingOption(kindCmd)
	installEventingOption(kindCmd)
	installKindRegistryOption(kindCmd)
	installKindExtraMountHostPathOption(kindCmd)
	installKindExtraMountContainerPathOption(kindCmd)

	return kindCmd
}
