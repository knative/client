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

package plugin

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/config"

	"knative.dev/client/pkg/commands"
	"knative.dev/client/pkg/plugin"
)

// ValidPluginFilenamePrefixes controls the prefix for all kn plugins
var ValidPluginFilenamePrefixes = []string{"kn"}

// pluginListFlags contains all plugin commands flags
type pluginListFlags struct {
	verbose bool
}

// NewPluginListCommand creates a new `kn plugin list` command
func NewPluginListCommand(p *commands.KnParams) *cobra.Command {

	plFlags := pluginListFlags{}
	pluginListCommand := &cobra.Command{
		Use:     "list",
		Short:   "List plugins",
		Aliases: []string{"ls"},
		Long: `List all installed plugins.

Available plugins are those that are:
- executable
- begin with "kn-"
- Kn's plugin directory
- Anywhere in the execution $PATH`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPlugins(cmd, plFlags)
		},
	}

	// Plugin flags
	pluginListCommand.Flags().BoolVar(&plFlags.verbose, "verbose", false, "verbose output")

	return pluginListCommand
}

// List plugins by looking up in plugin directory and path
func listPlugins(cmd *cobra.Command, flags pluginListFlags) error {
	factory := plugin.NewManager(config.GlobalConfig.PluginsDir(), config.GlobalConfig.LookupPluginsInPath())

	pluginsFound, err := factory.ListPlugins()
	if err != nil {
		return fmt.Errorf("cannot list plugins in %s (lookup plugins in $PATH: %t): %w", factory.PluginsDir(), factory.LookupInPath(), err)
	}

	out := cmd.OutOrStdout()
	if flags.verbose {
		fmt.Fprintf(out, "The following plugins are available, using options:\n")
		fmt.Fprintf(out, "  plugins dir: '%s'%s\n", factory.PluginsDir(), extraLabelIfPathNotExists(factory.PluginsDir()))
		fmt.Fprintf(out, "  lookup plugins in $PATH: %t\n\n", factory.LookupInPath())
	}

	if len(pluginsFound) == 0 {
		if flags.verbose {
			fmt.Fprintf(out, "No plugins found in path '%s'.\n", factory.PluginsDir())
		} else {
			fmt.Fprintln(out, "No plugins found.")
		}
		return nil
	}

	eaw := factory.Verify()
	eaw = addErrorIfOverwritingExistingCommand(eaw, cmd.Root(), pluginsFound)

	for _, pl := range pluginsFound {
		desc, _ := pl.Description()
		if desc != "" {
			fmt.Fprintf(out, "- %s : %s", pl.Name(), desc)
		} else {
			fmt.Fprintf(out, "- %s", pl.Name())
		}
		if flags.verbose {
			fmt.Fprintf(out, "  (%s)\n", pl.Path())
		} else {
			fmt.Fprintln(out, "")
		}
	}
	if !eaw.IsEmpty() {
		fmt.Fprintln(out, "")
		eaw.PrintWarningsAndErrors(out)
	}
	if eaw.HasErrors() {
		return fmt.Errorf("plugin validation errors")
	}
	return nil
}

// create an info label which can be appended to an verbose output
func extraLabelIfPathNotExists(path string) string {
	_, err := os.Stat(path)
	if err == nil {
		return ""
	}
	if os.IsNotExist(err) {
		return " (does not exist)"
	}
	return ""
}

func addErrorIfOverwritingExistingCommand(eaw plugin.VerificationErrorsAndWarnings, rootCmd *cobra.Command, plugins []plugin.Plugin) plugin.VerificationErrorsAndWarnings {

	for _, plugin := range plugins {
		cmd, args, err := rootCmd.Find(plugin.CommandParts())
		if err == nil {
			if !cmd.HasSubCommands() || // a leaf command can't be overridden
				cmd.HasSubCommands() && len(args) == 0 { // a group can't be overridden either
				eaw.AddError("%s overwrites existing built-in command: '%s'", plugin.Path(), strings.Join(plugin.CommandParts(), " "))
			}
		}
	}
	return eaw
}
