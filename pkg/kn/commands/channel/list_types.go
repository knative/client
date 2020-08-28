// Copyright © 2020 The Knative Authors
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

package channel

import (
	"fmt"

	"github.com/spf13/cobra"
	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewListTypesCommand defines and processes `kn channel list-types`
func NewListTypesCommand(p *commands.KnParams) *cobra.Command {
	listTypesFlags := flags.NewListPrintFlags(ListTypesHandlers)
	listTypesCommand := &cobra.Command{
		Use:   "list-types",
		Short: "List channel types",
		Example: `
  # List available channel types
  kn channel list-types

  # List available channel types in YAML format
  kn channel list-types -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			channelListTypes, err := dynamicClient.ListChannelsTypes()
			if err != nil {
				return knerrors.GetError(err)
			}

			if channelListTypes == nil || len(channelListTypes.Items) == 0 {
				return fmt.Errorf("no channel found on the backend, please verify the installation")
			}

			printer, err := listTypesFlags.ToPrinter()
			if err != nil {
				return nil
			}

			err = printer.PrintObj(channelListTypes, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(listTypesCommand.Flags(), false)
	listTypesFlags.AddFlags(listTypesCommand)
	return listTypesCommand
}
