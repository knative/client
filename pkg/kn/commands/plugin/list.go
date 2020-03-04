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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
)

// ValidPluginFilenamePrefixes controls the prefix for all kn plugins
var ValidPluginFilenamePrefixes = []string{"kn"}

// NewPluginListCommand creates a new `kn plugin list` command
func NewPluginListCommand(p *commands.KnParams) *cobra.Command {

	plFlags := pluginListFlags{}
	pluginListCommand := &cobra.Command{
		Use:   "list",
		Short: "List plugins",
		Long: `List all installed plugins.

Available plugins are those that are:
- executable
- begin with "kn-"
- Kn's plugin directory ` + commands.Cfg.DefaultPluginDir + `
- Anywhere in the execution $PATH (if lookupInPath config variable is enabled)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPlugins(cmd, plFlags)
		},
	}

	AddPluginFlags(pluginListCommand)
	BindPluginsFlagToViper(pluginListCommand)
	plFlags.AddPluginListFlags(pluginListCommand)

	return pluginListCommand
}

// List plugins by looking up in plugin directory and path
func listPlugins(cmd *cobra.Command, flags pluginListFlags) error {
	pluginPath, err := homedir.Expand(commands.Cfg.PluginsDir)
	if err != nil {
		return err
	}
	if *commands.Cfg.LookupPlugins {
		pluginPath = pluginPath + string(os.PathListSeparator) + os.Getenv("PATH")
	}

	pluginsFound, eaw := lookupPlugins(pluginPath)

	out := cmd.OutOrStdout()

	if flags.verbose {
		fmt.Fprintf(out, "The following plugins are available, using options:\n")
		fmt.Fprintf(out, "  - plugins dir: '%s'%s\n", commands.Cfg.PluginsDir, extraLabelIfPathNotExists(pluginPath))
		fmt.Fprintf(out, "  - lookup plugins in $PATH: '%t'\n", *commands.Cfg.LookupPlugins)
	}

	if len(pluginsFound) == 0 {
		if flags.verbose {
			fmt.Fprintf(out, "No plugins found in path '%s'.\n", pluginPath)
		} else {
			fmt.Fprintln(out, "No plugins found.")
		}
		return nil
	}

	verifier := newPluginVerifier(cmd.Root())
	for _, plugin := range pluginsFound {
		name := plugin
		if flags.nameOnly {
			name = filepath.Base(plugin)
		}
		fmt.Fprintf(out, "%s\n", name)
		eaw = verifier.verify(eaw, plugin)
	}
	eaw.printWarningsAndErrors(out)
	return eaw.combinedError()
}

func lookupPlugins(pluginPath string) ([]string, errorsAndWarnings) {
	pluginsFound := make([]string, 0)
	eaw := errorsAndWarnings{}

	for _, dir := range uniquePathsList(filepath.SplitList(pluginPath)) {

		files, err := ioutil.ReadDir(dir)

		// Ignore non-existing directories
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			eaw.addError("unable to read directory '%s' from your plugin path: %v", dir, err)
			continue
		}

		// Check for plugins within given directory
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !hasValidPrefix(f.Name(), ValidPluginFilenamePrefixes) {
				continue
			}
			pluginsFound = append(pluginsFound, filepath.Join(dir, f.Name()))
		}
	}
	return pluginsFound, eaw
}

func hasValidPrefix(filepath string, validPrefixes []string) bool {
	for _, prefix := range validPrefixes {
		if !strings.HasPrefix(filepath, prefix+"-") {
			continue
		}
		return true
	}
	return false
}

// uniquePathsList deduplicates a given slice of strings without
// sorting or otherwise altering its order in any way.
func uniquePathsList(paths []string) []string {
	seen := map[string]bool{}
	var newPaths []string
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		newPaths = append(newPaths, p)
	}
	return newPaths
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
