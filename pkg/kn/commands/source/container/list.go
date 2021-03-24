/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

import (
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewContainerListCommand is for listing Container sources
func NewContainerListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(ContainerSourceListHandlers)

	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List container sources",
		Example: `
  # List all Container sources
  kn source container list

  # List all Container sources in YAML format
  kn source apiserver list -o yaml`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			containerClient, err := newContainerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			sourceList, err := containerClient.ListContainerSources(cmd.Context())
			if err != nil {
				return err
			}

			if len(sourceList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No Container source found.\n")
				return nil
			}

			if containerClient.Namespace(cmd.Context()) == "" {
				listFlags.EnsureWithNamespace()
			}

			return listFlags.Print(sourceList, cmd.OutOrStdout())
		},
	}
	commands.AddNamespaceFlags(listCommand.Flags(), true)
	listFlags.AddFlags(listCommand)
	return listCommand
}
