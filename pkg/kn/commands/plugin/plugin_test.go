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
	. "github.com/knative/client/pkg/kn/commands"
	. "github.com/knative/client/pkg/kn/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

const PluginCommandUsage = `Provides utilities for interacting with kn plugins.

Plugins provide extended functionality that is not part of the major command-line distribution.
Please refer to the documentation and examples for more information about how write your own plugins.

Usage:
  kn plugin [flags]
  kn plugin [command]

Available Commands:
  list        List all visible plugin executables on a user's PATH

Flags:
  -h, --help   help for plugin

Global Flags:
      --config string       config file (default is $HOME/.kn/config.yaml)
      --kubeconfig string   kubectl config file (default is $HOME/.kube/config)
      --plugin-dir string   kn plugin directory (default is value in kn config or $PATH)

Use "kn plugin [command] --help" for more information about a command.`

var _ = Describe("kn plugin", func() {
	var (
		rootCmd, pluginCmd *cobra.Command
	)

	BeforeEach(func() {
		rootCmd = NewKnCommand(KnParams{})
		Expect(rootCmd).NotTo(BeNil())

		pluginCmd = FindSubCommand(rootCmd, "plugin")
		Expect(pluginCmd).NotTo(BeNil())
	})

	Describe("#NewPluginCommand", func() {
		It("creates a new cobra.Command", func() {
			Expect(pluginCmd).NotTo(BeNil())
			Expect(pluginCmd.Use).To(Equal("plugin"))
			Expect(pluginCmd.Short).To(Equal("Plugin command group"))
			Expect(pluginCmd.Long).To(ContainSubstring("Provides utilities for interacting with kn plugins."))
			Expect(pluginCmd.Args).To(BeNil())
			Expect(pluginCmd.RunE).NotTo(BeNil())
		})

		Context("when called with known subcommand", func() {
			var (
				fakeExecuted bool
			)

			BeforeEach(func() {
				pluginCmd.AddCommand(&cobra.Command{
					Use:   "fake",
					Short: "fake subcommand",
					RunE: func(cmd *cobra.Command, args []string) error {
						fakeExecuted = true
						return nil
					},
				})
			})

			It("executes the subcommand RunE func", func() {
				rootCmd.SetArgs([]string{"plugin", "fake"})
				err := rootCmd.Execute()
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeExecuted).To(BeTrue())
			})

			It("reads flag --plugin-dir", func() {
				rootCmd.SetArgs([]string{"plugin", "fake", "--plugin-dir", "$PATH"})
				err := rootCmd.Execute()
				Expect(err).NotTo(HaveOccurred())

				pluginDir, err := pluginCmd.Flags().GetString("plugin-dir")
				Expect(err).NotTo(HaveOccurred())
				Expect(pluginDir).To(Equal("$PATH"))
			})
		})

		Context("when called with unknown subcommand", func() {
			It("fails with 'unknown command' message", func() {
				rootCmd.SetArgs([]string{"plugin", "unknown"})
				err := rootCmd.Execute()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unknown command \"unknown\" for"))
				Expect(ReadStdout()).To(ContainSubstring(PluginCommandUsage))
			})
		})

		Context("when called with empty subcommand", func() {
			It("fails with 'please provide a valid sub-command' message", func() {
				rootCmd.SetArgs([]string{"plugin"})
				err := rootCmd.Execute()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("please provide a valid sub-command for \"kn plugin\""))
				Expect(ReadStdout()).To(ContainSubstring(PluginCommandUsage))
			})
		})
	})
})
