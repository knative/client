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

	"github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1/fake"
	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestCreateTestKnCommand(t *testing.T) {
	var (
		knCmd   *cobra.Command
		serving *fake.FakeServingV1alpha1
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
