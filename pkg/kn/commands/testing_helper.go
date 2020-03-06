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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"

	"knative.dev/client/pkg/kn/flags"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/sources/v1alpha2"

	dynamicfake "k8s.io/client-go/dynamic/fake"
	eventingv1alpha1fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1/fake"
	sourcesv1alpha2fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2/fake"

	clientdynamic "knative.dev/client/pkg/dynamic"
	eventingv1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
)

const FakeNamespace = "current"

var (
	oldStdout *os.File
	stdout    *os.File
	output    string

	readFile, writeFile *os.File
)

// CreateTestKnCommand helper for creating test commands
func CreateTestKnCommand(cmd *cobra.Command, knParams *KnParams) (*cobra.Command, *servingv1fake.FakeServingV1, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	fakeServing := &servingv1fake.FakeServingV1{&clienttesting.Fake{}}
	knParams.Output = buf
	knParams.NewServingClient = func(namespace string) (clientservingv1.KnServingClient, error) {
		return clientservingv1.NewKnServingClient(fakeServing, FakeNamespace), nil
	}
	knParams.fixedCurrentNamespace = FakeNamespace
	knCommand := NewKnTestCommand(cmd, knParams)
	return knCommand, fakeServing, buf
}

// CreateSourcesTestKnCommand helper for creating test commands
func CreateSourcesTestKnCommand(cmd *cobra.Command, knParams *KnParams) (*cobra.Command, *sourcesv1alpha2fake.FakeSourcesV1alpha2, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	// create fake serving client because the sink of source depends on serving client
	fakeServing := &servingv1fake.FakeServingV1{&clienttesting.Fake{}}
	knParams.NewServingClient = func(namespace string) (clientservingv1.KnServingClient, error) {
		return clientservingv1.NewKnServingClient(fakeServing, FakeNamespace), nil
	}
	// create fake sources client
	fakeEventing := &sourcesv1alpha2fake.FakeSourcesV1alpha2{&clienttesting.Fake{}}
	knParams.Output = buf
	knParams.NewSourcesClient = func(namespace string) (v1alpha2.KnSourcesClient, error) {
		return v1alpha2.NewKnSourcesClient(fakeEventing, FakeNamespace), nil
	}
	knParams.fixedCurrentNamespace = FakeNamespace
	knCommand := NewKnTestCommand(cmd, knParams)
	return knCommand, fakeEventing, buf
}

// CreateEventingTestKnCommand helper for creating test commands
func CreateEventingTestKnCommand(cmd *cobra.Command, knParams *KnParams) (*cobra.Command, *eventingv1alpha1fake.FakeEventingV1alpha1, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	// create fake serving client because the sink of source depends on serving client
	fakeServing := &servingv1fake.FakeServingV1{&clienttesting.Fake{}}
	knParams.NewServingClient = func(namespace string) (clientservingv1.KnServingClient, error) {
		return clientservingv1.NewKnServingClient(fakeServing, FakeNamespace), nil
	}
	// create fake sources client
	fakeEventing := &eventingv1alpha1fake.FakeEventingV1alpha1{&clienttesting.Fake{}}
	knParams.Output = buf
	knParams.NewEventingClient = func(namespace string) (eventingv1alpha1.KnEventingClient, error) {
		return eventingv1alpha1.NewKnEventingClient(fakeEventing, FakeNamespace), nil
	}
	knParams.fixedCurrentNamespace = FakeNamespace
	knCommand := NewKnTestCommand(cmd, knParams)
	return knCommand, fakeEventing, buf
}

// CreateDynamicTestKnCommand helper for creating test commands using dynamic client
func CreateDynamicTestKnCommand(cmd *cobra.Command, knParams *KnParams, objects ...runtime.Object) (*cobra.Command, *dynamicfake.FakeDynamicClient, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	fakeDynamic := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)
	knParams.Output = buf
	knParams.NewDynamicClient = func(namespace string) (clientdynamic.KnDynamicClient, error) {
		return clientdynamic.NewKnDynamicClient(fakeDynamic, FakeNamespace), nil
	}

	knParams.fixedCurrentNamespace = FakeNamespace
	knCommand := NewKnTestCommand(cmd, knParams)
	return knCommand, fakeDynamic, buf

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

// NewKnTestCommand needed since calling the one in core would cause a import cycle
func NewKnTestCommand(subCommand *cobra.Command, params *KnParams) *cobra.Command {
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

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return flags.ReconcileBoolFlags(cmd.Flags())
		},
	}
	if params.Output != nil {
		rootCmd.SetOutput(params.Output)
	}
	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file (default is ~/.config/kn/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&params.KubeCfgPath, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	rootCmd.Flags().StringVar(&Cfg.PluginsDir, "plugins-dir", "~/.config/kn/plugins", "kn plugins directory")
	rootCmd.Flags().BoolVar(Cfg.LookupPlugins, "lookup-plugins", false, "look for kn plugins in $PATH")

	viper.BindPFlag("plugins-dir", rootCmd.Flags().Lookup("plugins-dir"))
	viper.BindPFlag("lookup-plugins", rootCmd.Flags().Lookup("lookup-plugins"))

	viper.SetDefault("plugins-dir", "~/.config/kn/plugins")
	viper.SetDefault("lookup-plugins", false)

	rootCmd.AddCommand(subCommand)

	// For glog parse error.
	flag.CommandLine.Parse([]string{})
	return rootCmd
}
