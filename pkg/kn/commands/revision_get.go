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

package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRevisionGetCommand represents 'kn revision get' command
func NewRevisionGetCommand(p *KnParams) *cobra.Command {
	revisionGetFlags := NewRevisionGetFlags()

	revisionGetCommand := &cobra.Command{
		Use:   "get",
		Short: "Get available revisions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := GetNamespace(cmd)
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

			printer, err := revisionGetFlags.ToPrinter()
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
	AddNamespaceFlags(revisionGetCommand.Flags(), true)
	revisionGetFlags.AddFlags(revisionGetCommand)
	return revisionGetCommand
}
