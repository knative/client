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

func TestCompletionUsage(t *testing.T) {
	completionCmd := NewCompletionCommand(&commands.KnParams{})
	assert.Assert(t, util.ContainsAllIgnoreCase(completionCmd.Use, "completion"))
	assert.Assert(t, util.ContainsAllIgnoreCase(completionCmd.Short, "completion", "shell"))
	assert.Assert(t, completionCmd.RunE == nil)
}

func TestCompletionGeneration(t *testing.T) {
	for _, shell := range []string{"bash", "zsh"} {
		completionCmd := NewCompletionCommand(&commands.KnParams{})
		c := commands.CaptureStdout(t)
		completionCmd.Run(&cobra.Command{}, []string{shell})
		out := c.Close()
		assert.Assert(t, out != "")
	}
}

func TestCompletionNoArg(t *testing.T) {
	completionCmd := NewCompletionCommand(&commands.KnParams{})
	c := commands.CaptureStdout(t)
	completionCmd.Run(&cobra.Command{}, []string{})
	out := c.Close()
	assert.Assert(t, util.ContainsAll(out, "bash", "zsh", "one", "argument"))
}

func TestCompletionWrongArg(t *testing.T) {
	completionCmd := NewCompletionCommand(&commands.KnParams{})
	c := commands.CaptureStdout(t)
	completionCmd.Run(&cobra.Command{}, []string{"sh"})
	out := c.Close()
	assert.Assert(t, util.ContainsAll(out, "bash", "zsh", "only", "supports"))
}
