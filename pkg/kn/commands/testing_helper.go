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

package commands

import (
	"bytes"
	"flag"
	"io"
	"os"
	"testing"

	"github.com/knative/client/pkg/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gotest.tools/assert"
	client_testing "k8s.io/client-go/testing"
)

const FakeNamespace = "current"

var (
	oldStdout *os.File
	stdout    *os.File
	output    string

	readFile, writeFile *os.File

	origArgs []string
)

// CreateTestKnCommand helper for creating test commands
func CreateTestKnCommand(cmd *cobra.Command, knParams *KnParams) (*cobra.Command, *fake.FakeServingV1alpha1, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	fakeServing := &fake.FakeServingV1alpha1{&client_testing.Fake{}}
	knParams.Output = buf
	knParams.NewClient = func(namespace string) (v1alpha1.KnClient, error) {
		return v1alpha1.NewKnServingClient(fakeServing, namespace), nil
	}
	knParams.fixedCurrentNamespace = FakeNamespace
	knCommand := newKnCommand(cmd, knParams)
	return knCommand, fakeServing, buf
}

// CaptureStdout collects the current content of os.Stdout
func CaptureStdout(t *testing.T) {
	oldStdout = os.Stdout
	var err error
	readFile, writeFile, err = os.Pipe()
	assert.Assert(t, err == nil)
	stdout = writeFile
	os.Stdout = writeFile
}

// ReleaseStdout releases the os.Stdout and restores to original
func ReleaseStdout(t *testing.T) {
	output = ReadStdout(t)
	os.Stdout = oldStdout
}

// ReadStdout returns the collected os.Stdout content
func ReadStdout(t *testing.T) string {
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, readFile)
		outC <- buf.String()
	}()
	writeFile.Close()
	output = <-outC

	CaptureStdout(t)

	return output
}

// Private

// newKnCommand needed since calling the one in core would cause a import cycle
func newKnCommand(subCommand *cobra.Command, params *KnParams) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kn",
		Short: "Knative client",
		Long: `Manage your Knative building blocks:

Serving: Manage your services and release new software to them.
Build: Create builds and keep track of their results.
Eventing: Manage event subscriptions and channels. Connect up event sources.`,

		// Disable docs header
		DisableAutoGenTag: true,

		// Affects children as well
		SilenceUsage: true,

		// Prevents Cobra from dealing with errors as we deal with them in main.go
		SilenceErrors: true,
	}
	if params.Output != nil {
		rootCmd.SetOutput(params.Output)
	}
	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file (default is $HOME/.kn.yaml)")
	rootCmd.PersistentFlags().StringVar(&params.KubeCfgPath, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	rootCmd.Flags().StringVar(&Cfg.PluginsDir, "plugins-dir", "~/.kn/plugins", "kn plugins directory")
	rootCmd.Flags().BoolVar(&Cfg.LookupPluginsInPath, "lookup-plugins-in-path", false, "look for kn plugins in $PATH")

	viper.BindPFlag("pluginsDir", rootCmd.Flags().Lookup("plugins-dir"))
	viper.BindPFlag("lookupPluginsInPath", rootCmd.Flags().Lookup("lookup-plugins-in-path"))

	viper.SetDefault("pluginsDir", "~/.kn/plugins")
	viper.SetDefault("lookupPluginsInPath", false)

	rootCmd.AddCommand(subCommand)

	// For glog parse error.
	flag.CommandLine.Parse([]string{})
	return rootCmd
}
