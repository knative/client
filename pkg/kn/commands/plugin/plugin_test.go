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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
	"knative.dev/client/pkg/kn/commands"
)

const PluginCommandUsage = `Provides utilities for interacting and managing with kn plugins.

Plugins provide extended functionality that is not part of the core kn command-line distribution.
Please refer to the documentation and examples for more information about how write your own plugins.

Usage:
  kn plugin [flags]
  kn plugin [command]

Available Commands:
  list        List all visible plugin executables

Flags:
  -h, --help                     help for plugin
      --lookup-plugins           look for kn plugins in $PATH
      --plugins-dir string       kn plugins directory (default "~/.config/kn/plugins")

Global Flags:
      --config string       kn config file (default is $HOME/.config/kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)

Use "kn plugin [command] --help" for more information about a command.`

func TestNewPluginCommand(t *testing.T) {
	var (
		rootCmd, pluginCmd *cobra.Command
	)

	setup := func(t *testing.T) {
		knParams := &commands.KnParams{}
		pluginCmd = NewPluginCommand(knParams)
		assert.Assert(t, pluginCmd != nil)

		rootCmd, _, _ = commands.CreateTestKnCommand(pluginCmd, knParams)
		assert.Assert(t, rootCmd != nil)
	}

	t.Run("creates a new cobra.Command", func(t *testing.T) {
		setup(t)

		assert.Assert(t, pluginCmd != nil)
		assert.Assert(t, pluginCmd.Use == "plugin")
		assert.Assert(t, pluginCmd.Short == "Plugin command group")
		assert.Assert(t, strings.Contains(pluginCmd.Long, "Provides utilities for interacting and managing with kn plugins."))
		assert.Assert(t, pluginCmd.Flags().Lookup("plugins-dir") != nil)
		assert.Assert(t, pluginCmd.Flags().Lookup("lookup-plugins") != nil)
		assert.Assert(t, pluginCmd.Args == nil)
	})
}
