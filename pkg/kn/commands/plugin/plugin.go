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
	"github.com/spf13/viper"
	"knative.dev/client/pkg/kn/commands"
)

func NewPluginCommand(p *commands.KnParams) *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin command group",
		Long: `Provides utilities for interacting and managing with kn plugins.

Plugins provide extended functionality that is not part of the core kn command-line distribution.
Please refer to the documentation and examples for more information about how write your own plugins.`,
	}

	AddPluginFlags(pluginCmd)
	BindPluginsFlagToViper(pluginCmd)

	pluginCmd.AddCommand(NewPluginListCommand(p))

	return pluginCmd
}

// AddPluginFlags plugins-dir and lookup-plugins to cmd
func AddPluginFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&commands.Cfg.PluginsDir, "plugins-dir", commands.Cfg.DefaultPluginDir, "kn plugins directory")
	cmd.Flags().BoolVar(commands.Cfg.LookupPlugins, "lookup-plugins", false, "look for kn plugins in $PATH")
}

// BindPluginsFlagToViper bind and set default with viper for plugins flags
func BindPluginsFlagToViper(cmd *cobra.Command) {
	viper.BindPFlag("plugins-dir", cmd.Flags().Lookup("plugins-dir"))
	viper.BindPFlag("lookup-plugins", cmd.Flags().Lookup("lookup-plugins"))

	viper.SetDefault("plugins-dir", commands.Cfg.DefaultPluginDir)
	viper.SetDefault("lookup-plugins", false)
}

// AllowedExtensibleCommandGroups the list of command groups that can be
// extended with plugins, e.g., a plugin named `kn-source-kafka` for Kafka
// event sources is allowed. This is defined as a fixed [...]string since
// cannot defined Golang []string constants
var AllowedExtensibleCommandGroups = [...]string{"source"}

// InAllowedExtensibleCommandGroups checks that the name is in the list of allowed
// extensible command groups
func InAllowedExtensibleCommandGroups(name string) bool {
	for _, groupName := range AllowedExtensibleCommandGroups {
		if name == groupName {
			return true
		}
	}
	return false
}
