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

package domain

import (
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewDomainMappingListCommand represents 'kn revision list' command
func NewDomainMappingListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(DomainMappingListHandlers)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List domain mappings",
		Aliases: []string{"ls"},
		Example: `
  # List all domain mappings (Beta)
  kn domain list 

  # List all domain mappings in JSON output format
  kn domain list -o json (Beta)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			if namespace == "" {
				listFlags.EnsureWithNamespace()
			}

			client, err := p.NewServingV1alpha1Client(namespace)
			if err != nil {
				return err
			}

			domainMappingList, err := client.ListDomainMappings(cmd.Context())
			if err != nil {
				return err
			}

			if !listFlags.GenericPrintFlags.OutputFlagSpecified() && len(domainMappingList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No domain mapping found.\n")
				return nil
			}

			return listFlags.Print(domainMappingList, cmd.OutOrStdout())
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), true)
	listFlags.AddFlags(cmd)
	return cmd
}
