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

package container

import (
	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

// NewContainerCommand to manage containers
func NewContainerCommand(p *commands.KnParams) *cobra.Command {
	containerCmd := &cobra.Command{
		Use:     "container COMMAND",
		Short:   "Manage service's containers (experimental)",
		Aliases: []string{"containers"},
	}
	containerCmd.AddCommand(NewContainerAddCommand(p))
	return containerCmd
}
