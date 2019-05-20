// Copyright © 2019 The Knative Authors
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

package revision

import (
	"fmt"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"

	"github.com/knative/client/pkg/kn/commands"
)

// NewRevisionListCommand represents 'kn revision list' command
func NewRevisionListCommand(p *commands.KnParams) *cobra.Command {
	revisionListFlags := NewRevisionListFlags()

	revisionListCommand := &cobra.Command{
		Use:   "list",
		Short: "List available revisions.",
		Long:  "List revisions for a given service.",
		Example: `
  # List all revisions
  kn revision list

  # List revisions for a service 'svc1' in namespace 'myapp'
  kn revision list -s svc1 -n myapp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}
			var revisionList *v1alpha1.RevisionList
			if cmd.Flags().Changed("service") {
				serviceName := cmd.Flag("service").Value.String()

				// Verify that service exists
				_, err := client.GetService(serviceName)
				if err != nil {
					return err
				}
				revisionList, err = client.ListRevisionsForService(serviceName)
				if err != nil {
					return err
				}
				if len(revisionList.Items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No revisions found for service '%s'.\n", serviceName)
					return nil
				}
			} else {
				revisionList, err = client.ListRevisions()
				if err != nil {
					return err
				}
				if len(revisionList.Items) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No revisions found.\n")
					return nil
				}
			}

			printer, err := revisionListFlags.ToPrinter()
			if err != nil {
				return err
			}
			err = printer.PrintObj(revisionList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(revisionListCommand.Flags(), true)
	revisionListFlags.AddFlags(revisionListCommand)
	return revisionListCommand
}
