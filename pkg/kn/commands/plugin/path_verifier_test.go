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
	"fmt"
	"strings"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

var (
	pluginPath string
	rootCmd    *cobra.Command
	verifier   *CommandOverrideVerifier
)

var setup = func() {
	knParams := &commands.KnParams{}
	rootCmd, _, _ = commands.CreateTestKnCommand(NewPluginCommand(knParams), knParams)
	verifier = &CommandOverrideVerifier{
		Root:        rootCmd,
		SeenPlugins: make(map[string]string),
	}
}

var cleanup = func(t *testing.T) {
	if pluginPath != "" {
		DeleteTestPlugin(t, pluginPath)
	}
}

func TestCommandOverrideVerifier_Verify_with_nil_root_command(t *testing.T) {
	t.Run("returns error verifying path", func(t *testing.T) {
		setup()
		defer cleanup(t)
		verifier.Root = nil

		errs := verifier.Verify(pluginPath)
		assert.Assert(t, len(errs) == 1)
		assert.Assert(t, errs[0] != nil)
		assert.Assert(t, strings.Contains(errs[0].Error(), "unable to verify path with nil root"))
	})
}

func TestCommandOverrideVerifier_Verify_with_root_command(t *testing.T) {
	t.Run("when plugin in path not executable", func(t *testing.T) {
		setup()
		defer cleanup(t)
		pluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeReadable)

		t.Run("fails with not executable error", func(t *testing.T) {
			errs := verifier.Verify(pluginPath)
			assert.Assert(t, len(errs) == 1)
			assert.Assert(t, errs[0] != nil)
			errorMsg := fmt.Sprintf("warning: %s identified as a kn plugin, but it is not executable", pluginPath)
			assert.Assert(t, strings.Contains(errs[0].Error(), errorMsg))
		})
	})

	t.Run("when kn plugin in path is executable", func(t *testing.T) {
		setup()
		defer cleanup(t)
		pluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable)

		t.Run("when kn plugin in path shadows another", func(t *testing.T) {
			var shadowPluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable)
			verifier.SeenPlugins[KnTestPluginName] = pluginPath
			defer DeleteTestPlugin(t, shadowPluginPath)

			t.Run("fails with overshadowed error", func(t *testing.T) {
				errs := verifier.Verify(shadowPluginPath)
				assert.Assert(t, len(errs) == 1)
				assert.Assert(t, errs[0] != nil)
				errorMsg := fmt.Sprintf("warning: %s is overshadowed by a similarly named plugin: %s", shadowPluginPath, pluginPath)
				assert.Assert(t, strings.Contains(errs[0].Error(), errorMsg))
			})
		})
	})

	t.Run("when kn plugin in path overwrites existing command", func(t *testing.T) {
		setup()
		defer cleanup(t)
		var overwritingPluginPath = CreateTestPlugin(t, "kn-plugin", KnTestPluginScript, FileModeExecutable)
		defer DeleteTestPlugin(t, overwritingPluginPath)

		t.Run("fails with overwrites error", func(t *testing.T) {
			errs := verifier.Verify(overwritingPluginPath)
			assert.Assert(t, len(errs) == 1)
			assert.Assert(t, errs[0] != nil)
			errorMsg := fmt.Sprintf("warning: %s overwrites existing command: %q", "kn-plugin", "kn plugin")
			assert.Assert(t, strings.Contains(errs[0].Error(), errorMsg))
		})
	})
}
