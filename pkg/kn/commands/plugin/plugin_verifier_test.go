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
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/util"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestPluginVerifier(t *testing.T) {
	var (
		pluginPath string
		rootCmd    *cobra.Command
		verifier   *pluginVerifier
	)

	setup := func(t *testing.T) {
		knParams := &commands.KnParams{}
		rootCmd, _, _ = commands.CreateTestKnCommand(NewPluginCommand(knParams), knParams)
		verifier = newPluginVerifier(rootCmd)
	}

	cleanup := func(t *testing.T) {
		if pluginPath != "" {
			DeleteTestPlugin(t, pluginPath)
		}
	}

	t.Run("with nil root command", func(t *testing.T) {
		t.Run("returns error verifying path", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			verifier.root = nil
			eaw := errorsAndWarnings{}
			eaw = verifier.verify(eaw, pluginPath)
			assert.Assert(t, len(eaw.errors) == 1)
			assert.Assert(t, len(eaw.warnings) == 0)
			assert.Assert(t, util.ContainsAll(eaw.errors[0], "nil root"))
		})
	})

	t.Run("with root command", func(t *testing.T) {
		t.Run("when plugin in path not executable", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			pluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeReadable)

			t.Run("fails with not executable error", func(t *testing.T) {
				eaw := errorsAndWarnings{}
				eaw = verifier.verify(eaw, pluginPath)
				assert.Assert(t, len(eaw.warnings) == 1)
				assert.Assert(t, len(eaw.errors) == 0)
				assert.Assert(t, util.ContainsAll(eaw.warnings[0], pluginPath, "not executable"))
			})
		})

		t.Run("when kn plugin in path is executable", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			pluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable)

			t.Run("when kn plugin in path shadows another", func(t *testing.T) {
				var shadowPluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable)
				verifier.seenPlugins[KnTestPluginName] = pluginPath
				defer DeleteTestPlugin(t, shadowPluginPath)

				t.Run("fails with overshadowed error", func(t *testing.T) {
					eaw := errorsAndWarnings{}
					eaw = verifier.verify(eaw, shadowPluginPath)
					assert.Assert(t, len(eaw.errors) == 0)
					assert.Assert(t, len(eaw.warnings) == 1)
					assert.Assert(t, util.ContainsAll(eaw.warnings[0], "shadowed", "ignored"))
				})
			})
		})

		t.Run("when kn plugin in path overwrites existing command", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			var overwritingPluginPath = CreateTestPlugin(t, "kn-plugin", KnTestPluginScript, FileModeExecutable)
			defer DeleteTestPlugin(t, overwritingPluginPath)

			t.Run("fails with overwrites error", func(t *testing.T) {
				eaw := errorsAndWarnings{}
				eaw = verifier.verify(eaw, overwritingPluginPath)
				assert.Assert(t, len(eaw.errors) == 1)
				assert.Assert(t, len(eaw.warnings) == 0)
				assert.Assert(t, util.ContainsAll(eaw.errors[0], "overwrite", "kn-plugin"))
			})
		})
	})
}
