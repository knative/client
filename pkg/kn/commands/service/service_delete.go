// Copyright © 2018 The Knative Authors
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

package service

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

// NewServiceDeleteCommand represent 'service delete' command
func NewServiceDeleteCommand(p *commands.KnParams) *cobra.Command {
	serviceDeleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a service.",
		Example: `
  # Delete a service 'svc1' in default namespace
  kn service delete svc1

  # Delete a service 'svc2' in 'ns1' namespace
  kn service delete svc2 -n ns1`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires the service name.")
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}

			err = client.DeleteService(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Service '%s' successfully deleted in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(serviceDeleteCommand.Flags(), false)
	return serviceDeleteCommand
}
