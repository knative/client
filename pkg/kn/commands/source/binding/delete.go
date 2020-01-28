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

package binding

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

// NewBindingDeleteCommand is for deleting a sink binding
func NewBindingDeleteCommand(p *commands.KnParams) *cobra.Command {
	BindingDeleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a sink binding.",
		Example: `
  # Delete a sink binding with name 'my-binding'
  kn source binding delete my-binding`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'requires the name of the sink bindinbg to delete as single argument")
			}
			name := args[0]

			bindingClient, err := newSinkBindingClient(p, cmd)
			if err != nil {
				return err
			}

			err = bindingClient.DeleteSinkBinding(name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Sink binding '%s' deleted in namespace '%s'.\n", name, bindingClient.Namespace())
			return nil
		},
	}
	commands.AddNamespaceFlags(BindingDeleteCommand.Flags(), false)
	return BindingDeleteCommand
}
