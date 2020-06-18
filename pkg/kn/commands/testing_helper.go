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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clienttesting "k8s.io/client-go/testing"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"

	"knative.dev/client/pkg/kn/flags"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/client/pkg/sources/v1alpha2"

	dynamicfake "k8s.io/client-go/dynamic/fake"
	sourcesv1alpha2fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2/fake"

	clientdynamic "knative.dev/client/pkg/dynamic"
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
	knCommand := NewTestCommand(cmd, knParams)
	return knCommand, fakeServing, buf
}

// CreateSourcesTestKnCommand helper for creating test commands
func CreateSourcesTestKnCommand(cmd *cobra.Command, knParams *KnParams) (*cobra.Command, *sourcesv1alpha2fake.FakeSourcesV1alpha2, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	// create fake serving client because the sink of source depends on serving client
	fakeServing := &servingv1fake.FakeServingV1{Fake: &clienttesting.Fake{}}
	knParams.NewServingClient = func(namespace string) (clientservingv1.KnServingClient, error) {
		return clientservingv1.NewKnServingClient(fakeServing, FakeNamespace), nil
	}
	// create fake sources client
	fakeEventing := &sourcesv1alpha2fake.FakeSourcesV1alpha2{Fake: &clienttesting.Fake{}}
	knParams.Output = buf
	knParams.NewSourcesClient = func(namespace string) (v1alpha2.KnSourcesClient, error) {
		return v1alpha2.NewKnSourcesClient(fakeEventing, FakeNamespace), nil
	}
	knParams.fixedCurrentNamespace = FakeNamespace
	knCommand := NewTestCommand(cmd, knParams)
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
	knCommand := NewTestCommand(cmd, knParams)
	return knCommand, fakeDynamic, buf

}

// NewTestCommand can be used by tes
func NewTestCommand(subCommand *cobra.Command, params *KnParams) *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "kn",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return flags.ReconcileBoolFlags(cmd.Flags())
		},
	}
	if params.Output != nil {
		rootCmd.SetOut(params.Output)
	}
	rootCmd.AddCommand(subCommand)
	return rootCmd
}
