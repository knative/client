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

package completion

import (
	"testing"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestCompletion(t *testing.T) {
	var (
		fakeRootCmd, completionCmd *cobra.Command
		knParams                   *commands.KnParams
	)

	setup := func() {
		knParams = &commands.KnParams{}
		completionCmd = NewCompletionCommand(knParams)

		fakeRootCmd = &cobra.Command{}
		fakeRootCmd.AddCommand(completionCmd)
	}

	t.Run("creates a CompletionCommand", func(t *testing.T) {
		setup()
		assert.Equal(t, completionCmd.Use, "completion [SHELL]")
		assert.Equal(t, completionCmd.Short, "Output shell completion code")
		assert.Assert(t, completionCmd.RunE == nil)
	})

	t.Run("returns completion code for BASH", func(t *testing.T) {
		setup()
		commands.CaptureStdout(t)
		defer commands.ReleaseStdout(t)

		completionCmd.Run(fakeRootCmd, []string{"bash"})
		assert.Assert(t, commands.ReadStdout(t) != "")
	})

	t.Run("returns completion code for Zsh", func(t *testing.T) {
		setup()
		commands.CaptureStdout(t)
		defer commands.ReleaseStdout(t)

		completionCmd.Run(fakeRootCmd, []string{"zsh"})
		assert.Assert(t, commands.ReadStdout(t) != "")
	})

	t.Run("returns error on command without args", func(t *testing.T) {
		setup()
		commands.CaptureStdout(t)
		defer commands.ReleaseStdout(t)

		completionCmd.Run(fakeRootCmd, []string{})
		assert.Assert(t, commands.ReadStdout(t) == "accepts one argument either 'bash' or 'zsh'\n")
	})

	t.Run("returns error on command with invalid args", func(t *testing.T) {
		setup()
		commands.CaptureStdout(t)
		defer commands.ReleaseStdout(t)

		completionCmd.Run(fakeRootCmd, []string{"sh"})
		assert.Check(t, util.ContainsAll(commands.ReadStdout(t), "only supports 'bash' or 'zsh' shell completion"))
	})
}
