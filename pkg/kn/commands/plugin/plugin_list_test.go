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
	"io/ioutil"
	"os"

	. "github.com/knative/client/pkg/kn/commands"
	. "github.com/knative/client/pkg/kn/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

var _ = Describe("kn plugin list", func() {
	var (
		rootCmd, pluginCmd, pluginListCmd *cobra.Command
		tmpPathDir                        string
		err                               error
	)

	BeforeEach(func() {
		rootCmd = NewKnCommand(KnParams{})
		Expect(rootCmd).NotTo(BeNil())

		pluginCmd = FindSubCommand(rootCmd, "plugin")
		Expect(pluginCmd).NotTo(BeNil())

		pluginListCmd = FindSubCommand(pluginCmd, "list")
		Expect(pluginListCmd).NotTo(BeNil())

		tmpPathDir, err = ioutil.TempDir("", "plugin_list")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err = os.RemoveAll(tmpPathDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("#NewPluginListCommand", func() {
		It("creates a new cobra.Command", func() {
			Expect(pluginListCmd).NotTo(BeNil())
			Expect(pluginListCmd.Use).To(Equal("list"))
			Expect(pluginListCmd.Short).To(Equal("List all visible plugin executables on a user's PATH"))
			Expect(pluginListCmd.Long).To(ContainSubstring("List all visible plugin executables on a user's PATH"))
			Expect(pluginListCmd.RunE).NotTo(BeNil())
		})

		Context("when using $PATH as plugin location", func() {
			var pluginPath string

			BeforeEach(func() {
				err = os.Setenv("PATH", tmpPathDir)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("no plugins installed", func() {
				It("warns user that no plugins found", func() {
					rootCmd.SetArgs([]string{"plugin", "list"})
					err = rootCmd.Execute()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("warning: unable to find any kn plugins in your PATH"))
				})
			})

			Context("plugins installed", func() {
				Context("with valid plugin in $PATH", func() {
					BeforeEach(func() {
						pluginPath = CreateTestPluginInPath(KnTestPluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir)
						Expect(pluginPath).NotTo(BeEmpty())
					})

					It("list plugins in $PATH", func() {
						rootCmd.SetArgs([]string{"plugin", "list"})
						err = rootCmd.Execute()
						Expect(err).NotTo(HaveOccurred())

						//TODO: test output to contain the plugin
					})
				})

				Context("with non-executable plugin", func() {
					BeforeEach(func() {
						pluginPath = CreateTestPluginInPath(KnTestPluginName, KnTestPluginScript, FileModeReadable, tmpPathDir)
						Expect(pluginPath).NotTo(BeEmpty())
					})

					It("warns user plugin invalid", func() {
						rootCmd.SetArgs([]string{"plugin", "list"})
						err = rootCmd.Execute()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("error: one plugin warning was found"))
					})
				})

				Context("with plugins with same name", func() {
					var tmpPathDir2 string

					BeforeEach(func() {
						pluginPath = CreateTestPluginInPath(KnTestPluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir)
						Expect(pluginPath).NotTo(BeEmpty())

						tmpPathDir2, err = ioutil.TempDir("", "plugin_list")
						Expect(err).NotTo(HaveOccurred())

						err = os.Setenv("PATH", tmpPathDir+":"+tmpPathDir2)
						Expect(err).NotTo(HaveOccurred())

						pluginPath = CreateTestPluginInPath(KnTestPluginName, KnTestPluginScript, FileModeExecutable, tmpPathDir2)
						Expect(pluginPath).NotTo(BeEmpty())
					})

					AfterEach(func() {
						err = os.RemoveAll(tmpPathDir)
						Expect(err).NotTo(HaveOccurred())
					})

					It("warns user about second (in $PATH) plugin shadowing first", func() {
						rootCmd.SetArgs([]string{"plugin", "list"})
						err = rootCmd.Execute()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("error: one plugin warning was found"))
					})
				})

				Context("with plugins with name of existing command", func() {
					BeforeEach(func() {
						pluginPath = CreateTestPluginInPath("kn-service", KnTestPluginScript, FileModeExecutable, tmpPathDir)
						Expect(pluginPath).NotTo(BeEmpty())
					})

					It("warns user about overwritting exising command", func() {
						rootCmd.SetArgs([]string{"plugin", "list"})
						err = rootCmd.Execute()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("error: one plugin warning was found"))
					})
				})
			})
		})

		Context("when using pluginDir config variable", func() {
			BeforeEach(func() {})
			AfterEach(func() {})
			Context("plugins installed", func() {})
			Context("no plugins installed", func() {})
		})
	})
})
