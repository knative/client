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
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
)

func NewPluginCommand(p *commands.KnParams) *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage kn plugins",
		Long: `Manage kn plugins

Plugins provide extended functionality that is not part of the core kn command-line distribution.
Please refer to the documentation and examples for more information about how to write your own plugins.`,
	}

	pluginCmd.AddCommand(NewPluginListCommand(p))

	return pluginCmd
}
