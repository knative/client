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

	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"github.com/spf13/cobra"
	client_testing "k8s.io/client-go/testing"
)

func CreateTestKnCommand(cmd *cobra.Command, knParams *KnParams) (*cobra.Command, *fake.FakeServingV1alpha1, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	fakeServing := &fake.FakeServingV1alpha1{&client_testing.Fake{}}
	knParams.Output = buf
	knParams.ServingFactory = func() (serving.ServingV1alpha1Interface, error) { return fakeServing, nil }
	knCommand := newKnCommand(cmd, knParams)
	return knCommand, fakeServing, buf
}

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
	rootCmd.PersistentFlags().StringVar(&KubeCfgFile, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	rootCmd.AddCommand(subCommand)

	// For glog parse error.
	flag.CommandLine.Parse([]string{})
	return rootCmd
}
