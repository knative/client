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
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewRevisionListCommand represents 'kn revision list' command
func NewRevisionListCommand(p *commands.KnParams) *cobra.Command {
	revisionListFlags := NewRevisionListFlags()

	revisionListCommand := &cobra.Command{
		Use:   "list",
		Short: "List available revisions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := commands.GetNamespace(cmd)
			if err != nil {
				return err
			}
			revision, err := client.Revisions(namespace).List(v1.ListOptions{})
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
