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

package plugin_test

import (
	"fmt"

	. "github.com/knative/client/pkg/kn/commands"
	. "github.com/knative/client/pkg/kn/commands/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("CommandOverrideVerifier", func() {
	var (
		pluginPath string
		rootCmd    *cobra.Command
		verifier   *CommandOverrideVerifier
	)

	BeforeEach(func() {
		knParams := &KnParams{}
		rootCmd, _, _ = CreateTestKnCommand(NewPluginCommand(knParams), knParams)
		verifier = &CommandOverrideVerifier{
			Root:        rootCmd,
			SeenPlugins: make(map[string]string),
		}
	})

	AfterEach(func() {
		if pluginPath != "" {
			DeleteTestPlugin(pluginPath)
		}
	})

	Describe("#Verify", func() {
		Context("when root command nil", func() {
			BeforeEach(func() {
				verifier.Root = nil
			})

			It("fails if root command is nil", func() {
				errs := verifier.Verify(pluginPath)
				Expect(len(errs)).To(Equal(1))
				Expect(errs[0]).NotTo(BeNil())
				Expect(errs[0].Error()).To(ContainSubstring("unable to verify path with nil root"))
			})
		})

		Context("when root command not nil", func() {
			Context("when kn plugin in path is not executable", func() {
				BeforeEach(func() {
					pluginPath = CreateTestPlugin(KnTestPluginName, KnTestPluginScript, FileModeReadable)
				})

				It("fails with not executable error", func() {
					errs := verifier.Verify(pluginPath)
					Expect(len(errs)).To(Equal(1))
					Expect(errs[0]).NotTo(BeNil())
					errorMsg := fmt.Sprintf("warning: %s identified as a kn plugin, but it is not executable", pluginPath)
					Expect(errs[0].Error()).To(ContainSubstring(errorMsg))
				})
			})

			Context("when kn plugin in path is executable", func() {
				BeforeEach(func() {
					pluginPath = CreateTestPlugin(KnTestPluginName, KnTestPluginScript, FileModeExecutable)
				})

				Context("when kn plugin in path shadows a similarly named plugin", func() {
					var (
						shadowPluginPath string
					)

					BeforeEach(func() {
						shadowPluginPath = CreateTestPlugin(KnTestPluginName, KnTestPluginScript, FileModeExecutable)
						verifier.SeenPlugins[KnTestPluginName] = pluginPath
					})

					AfterEach(func() {
						DeleteTestPlugin(shadowPluginPath)
					})

					It("fails with overshadowed error", func() {
						errs := verifier.Verify(shadowPluginPath)
						Expect(len(errs)).To(Equal(1))
						Expect(errs[0]).NotTo(BeNil())
						errorMsg := fmt.Sprintf("warning: %s is overshadowed by a similarly named plugin: %s", shadowPluginPath, pluginPath)
						Expect(errs[0].Error()).To(ContainSubstring(errorMsg))
					})
				})

				Context("when kn plugin in path overwrites existing command", func() {
					var (
						overwritingPluginPath string
					)

					BeforeEach(func() {
						overwritingPluginPath = CreateTestPlugin("kn-plugin", KnTestPluginScript, FileModeExecutable)
					})

					AfterEach(func() {
						DeleteTestPlugin(overwritingPluginPath)
					})

					It("fails with overwrites error", func() {
						errs := verifier.Verify(overwritingPluginPath)
						Expect(len(errs)).To(Equal(1))
						Expect(errs[0]).NotTo(BeNil())
						errorMsg := fmt.Sprintf("warning: %s overwrites existing command: %q", "kn-plugin", "kn plugin")
						Expect(errs[0].Error()).To(ContainSubstring(errorMsg))
					})
				})
			})
		})
	})
})
