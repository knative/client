// Copyright © 2019 The Knative Authors
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
	"fmt"
	"os"

	"knative.dev/client/pkg/kn/commands"

	"github.com/spf13/cobra"
)

const (
	desc = `
This command prints shell completion code which need to be evaluated
to provide interactive completion

Supported Shells:
 - bash
 - zsh`
	eg = `
 # Generate completion code for bash
 source <(kn completion bash)

 # Generate completion code for zsh
 source <(kn completion zsh)`
)

// NewCompletionCommand implements shell auto-completion feature for Bash and Zsh
func NewCompletionCommand(p *commands.KnParams) *cobra.Command {
	return &cobra.Command{
		Use:       "completion [SHELL]",
		Short:     "Output shell completion code",
		Long:      desc,
		ValidArgs: []string{"bash", "zsh"},
		Example:   eg,
		Hidden:    true, // Don't show this in help listing
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 1 {
				switch args[0] {
				case "bash":
					cmd.Root().GenBashCompletion(os.Stdout)
				case "zsh":
					cmd.Root().GenZshCompletion(os.Stdout)
				default:
					fmt.Println("only supports 'bash' or 'zsh' shell completion")
				}
			} else {
				fmt.Println("accepts one argument either 'bash' or 'zsh'")
			}
		},
	}
}
