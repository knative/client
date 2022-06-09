// Copyright © 2022 The Knative Authors
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

package configCmd

import (
	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
)

var cfgCmdExamples = ``

var ConfigKeyNotFoundErrorMsg = "Could not find %q in the configuration file"

func NewConfigCommand(p *commands.KnParams) *cobra.Command {
	cfgCmd := &cobra.Command{
		Use:     "config COMMAND",
		Short:   "Manage Knative configuration file",
		Aliases: []string{"cfg"},
		Example: cfgCmdExamples,
	}
	cfgCmd.AddCommand(NewConfigGetCommand(p))
	cfgCmd.AddCommand(NewConfigSetCommand(p))
	cfgCmd.AddCommand(NewConfigShowCommand(p))
	cfgCmd.AddCommand(NewConfigSinkMappingCommand(p))
	cfgCmd.AddCommand(NewConfigChannelTypeMappingCommand(p))
	return cfgCmd
}
