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

package plugin

import (
	"errors"
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
)

func NewPluginCommand(p *commands.KnParams) *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin command group",
		Long: `Provides utilities for interacting with kn plugins.

Plugins provide extended functionality that is not part of the major command-line distribution.
Please refer to the documentation and examples for more information about how write your own plugins.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Help()

			if len(args) == 0 {
				return errors.New("please provide a valid sub-command for \"kn plugin\"")
			} else {
				return errors.New(fmt.Sprintf("unknown command \"%s\" for \"kn plugin\"", args[0]))
			}
		},
	}
	pluginCmd.AddCommand(NewPluginListCommand(p))
	return pluginCmd
}
