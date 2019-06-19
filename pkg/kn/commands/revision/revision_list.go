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

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/serving/pkg/apis/serving"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := commands.GetNamespace(cmd)
			if err != nil {
				return err
			}
			listOptions := v1.ListOptions{}
			if cmd.Flags().Changed("service") {
				service := cmd.Flag("service").Value.String()
				// Ensure requested service exist
				_, err := client.Services(namespace).Get(service, v1.GetOptions{})
				if err != nil {
					return err
				}
				listOptions.LabelSelector = labels.Set(
					map[string]string{serving.ConfigurationLabelKey: service}).String()
			}
			revision, err := client.Revisions(namespace).List(listOptions)
			if err != nil {
				return err
			}
			if len(revision.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
				return nil
			}
			revision.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "revision"})
			printer, err := revisionListFlags.ToPrinter()
			if err != nil {
				return err
			}
			err = printer.PrintObj(revision, cmd.OutOrStdout())
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
