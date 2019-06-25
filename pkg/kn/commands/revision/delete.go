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
	"errors"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRevisionDeleteCommand represent 'revision delete' command
func NewRevisionDeleteCommand(p *commands.KnParams) *cobra.Command {
	RevisionDeleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a revision.",
		Example: `
  # Delete a revision 'svc1-abcde' in default namespace
  kn revision delete svc1-abcde`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires the revision name.")
			}
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			err = client.Revisions(namespace).Delete(
				args[0],
				&v1.DeleteOptions{},
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Revision '%s' successfully deleted in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(RevisionDeleteCommand.Flags(), false)
	return RevisionDeleteCommand
}
