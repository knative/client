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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/commands"
	knerrors "knative.dev/client/pkg/errors"
)

// NewDomainMappingDeleteCommand to create event channels
func NewDomainMappingDeleteCommand(p *commands.KnParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a domain mapping",
		Example: `
  # Delete domain mappings 'hello.example.com'
  kn domain delete hello.example.com`,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn domain delete' requires the domain name given as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewServingV1beta1Client(namespace)
			if err != nil {
				return err
			}

			err = client.DeleteDomainMapping(cmd.Context(), name)
			if err != nil {
				return knerrors.GetError(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Domain mapping '%s' deleted in namespace '%s'.\n", name, namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	return cmd
}
