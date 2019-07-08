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
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestCompletion(t *testing.T) {
	var (
		fakeRootCmd, completionCmd *cobra.Command
		knParams                   *KnParams
	)

	setup := func() {
		knParams = &KnParams{}
		completionCmd = NewCompletionCommand(knParams)

		fakeRootCmd = &cobra.Command{}
		fakeRootCmd.AddCommand(completionCmd)
	}

	t.Run("creates a CompletionCommand", func(t *testing.T) {
		setup()
		assert.Equal(t, completionCmd.Use, "completion")
		assert.Equal(t, completionCmd.Short, "Output shell completion code (default Bash)")
		assert.Assert(t, completionCmd.RunE == nil)
	})

	t.Run("returns completion code for BASH", func(t *testing.T) {
		setup()
		CaptureStdout(t)
		defer ReleaseStdout(t)

		completionCmd.Run(fakeRootCmd, []string{})
		assert.Assert(t, ReadStdout(t) != "")
	})

	t.Run("returns completion code for ZSH", func(t *testing.T) {
		setup()
		CaptureStdout(t)
		defer ReleaseStdout(t)

		completionCmd.Run(fakeRootCmd, []string{"--zsh"})
		assert.Assert(t, ReadStdout(t) != "")
	})
}
