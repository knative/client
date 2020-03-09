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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	sourcesv1alpha2fake "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2/fake"
	servingv1fake "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1/fake"
)

func TestCreateTestKnCommand(t *testing.T) {
	var (
		knCmd   *cobra.Command
		serving *servingv1fake.FakeServingV1
		buffer  *bytes.Buffer
	)

	setup := func(t *testing.T) {
		knParams := &KnParams{}
		knCmd, serving, buffer = CreateTestKnCommand(&cobra.Command{Use: "fake"}, knParams)
		assert.Assert(t, knCmd != nil)
		assert.Assert(t, len(knCmd.Commands()) == 1)
		assert.Assert(t, knCmd.Commands()[0].Use == "fake")
		assert.Assert(t, serving != nil)
		assert.Assert(t, buffer != nil)
	}

	t.Run("creates a new kn cobra.Command", func(t *testing.T) {
		setup(t)

		assert.Assert(t, knCmd != nil)
		assert.Assert(t, knCmd.Use == "kn")
		assert.Assert(t, knCmd.Short == "Knative client")
		assert.Assert(t, strings.Contains(knCmd.Long, "Manage your Knative building blocks:"))
		assert.Assert(t, knCmd.RunE == nil)
		assert.Assert(t, knCmd.DisableAutoGenTag == true)
		assert.Assert(t, knCmd.SilenceUsage == true)
		assert.Assert(t, knCmd.SilenceErrors == true)
	})
}

func TestCreateSourcesTestKnCommand(t *testing.T) {
	var (
		knCmd   *cobra.Command
		sources *sourcesv1alpha2fake.FakeSourcesV1alpha2
		buffer  *bytes.Buffer
	)

	setup := func(t *testing.T) {
		knParams := &KnParams{}
		knCmd, sources, buffer = CreateSourcesTestKnCommand(&cobra.Command{Use: "fake"}, knParams)
		assert.Assert(t, knCmd != nil)
		assert.Assert(t, len(knCmd.Commands()) == 1)
		assert.Assert(t, knCmd.Commands()[0].Use == "fake")
		assert.Assert(t, sources != nil)
		assert.Assert(t, buffer != nil)
	}

	t.Run("creates a new kn cobra.Command", func(t *testing.T) {
		setup(t)

		assert.Assert(t, knCmd != nil)
		assert.Assert(t, knCmd.Use == "kn")
		assert.Assert(t, knCmd.Short == "Knative client")
		assert.Assert(t, strings.Contains(knCmd.Long, "Manage your Knative building blocks:"))
		assert.Assert(t, knCmd.RunE == nil)
		assert.Assert(t, knCmd.DisableAutoGenTag == true)
		assert.Assert(t, knCmd.SilenceUsage == true)
		assert.Assert(t, knCmd.SilenceErrors == true)
	})
}

func TestCreateDynamicTestKnCommand(t *testing.T) {
	var (
		knCmd   *cobra.Command
		dynamic *dynamicfake.FakeDynamicClient
		buffer  *bytes.Buffer
	)

	setup := func(t *testing.T) {
		knParams := &KnParams{}
		knCmd, dynamic, buffer = CreateDynamicTestKnCommand(&cobra.Command{Use: "fake"}, knParams)
		assert.Assert(t, knCmd != nil)
		assert.Assert(t, len(knCmd.Commands()) == 1)
		assert.Assert(t, knCmd.Commands()[0].Use == "fake")
		assert.Assert(t, dynamic != nil)
		assert.Assert(t, buffer != nil)
		client, err := knParams.NewDynamicClient("foo")
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
	}

	t.Run("creates a new kn cobra.Command", func(t *testing.T) {
		setup(t)

		assert.Assert(t, knCmd != nil)
		assert.Assert(t, knCmd.Use == "kn")
		assert.Assert(t, knCmd.Short == "Knative client")
		assert.Assert(t, strings.Contains(knCmd.Long, "Manage your Knative building blocks:"))
		assert.Assert(t, knCmd.RunE == nil)
		assert.Assert(t, knCmd.DisableAutoGenTag == true)
		assert.Assert(t, knCmd.SilenceUsage == true)
		assert.Assert(t, knCmd.SilenceErrors == true)
	})

}
