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

package apiserver

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
)

// NewAPIServerDeleteCommand for deleting source
func NewAPIServerDeleteCommand(p *commands.KnParams) *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete an ApiServer source.",
		Example: `
  # Delete an ApiServerSource 'k8sevents' in default namespace
  kn source apiserver delete k8sevents`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires the name of the source as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			apiSourceClient, err := newAPIServerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			err = apiSourceClient.DeleteAPIServerSource(name)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ApiServer source '%s' deleted in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(deleteCommand.Flags(), false)
	return deleteCommand
}
