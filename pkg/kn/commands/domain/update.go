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

	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
)

// NewDomainMappingUpdateCommand to create event channels
func NewDomainMappingUpdateCommand(p *commands.KnParams) *cobra.Command {
	var refFlags RefFlags
	cmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a domain mapping",
		Example: `
  # Update a domain mappings 'hello.example.com' for Knative service 'hello'
  kn domain create hello.example.com --refFlags hello`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn domain create' requires the domain name given as single argument")
			}
			name := args[0]
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewServingV1alpha1Client(namespace)
			if err != nil {
				return err
			}

			toUpdate, err := client.GetDomainMapping(cmd.Context(), name)
			if err != nil {
				return err
			}

			if toUpdate.GetDeletionTimestamp() != nil {
				return fmt.Errorf("can't update domain mapping '%s' because it has been marked for deletion", name)
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			reference, err := refFlags.Resolve(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return err
			}
			toUpdate.Spec.Ref = *reference

			err = client.UpdateDomainMapping(cmd.Context(), toUpdate)
			if err != nil {
				return knerrors.GetError(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Domain mapping '%s' updated in namespace '%s'.\n", name, namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	refFlags.Add(cmd)
	cmd.MarkFlagRequired("refFlags")
	return cmd
}
