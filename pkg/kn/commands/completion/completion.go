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
	"errors"
	"os"

	"knative.dev/client/pkg/kn/commands"

	"github.com/spf13/cobra"
)

const (
	desc = `
This command prints shell completion code which needs to be evaluated
to provide interactive completion

Supported Shells:
 - bash
 - fish
 - powershell
 - zsh`
	eg = `
 # Generate completion code for bash
 source <(kn completion bash)

 # Generate completion code for fish
 kn completion fish | source

 # Generate completion code for powershell
 kn completion powershell | Out-String | Invoke-Expression

 # Generate completion code for zsh
 source <(kn completion zsh)
 compdef _kn kn`
)

// NewCompletionCommand implements shell auto-completion feature for Bash and Zsh
func NewCompletionCommand(p *commands.KnParams) *cobra.Command {
	return &cobra.Command{
		Use:       "completion SHELL",
		Short:     "Output shell completion code",
		Long:      desc,
		ValidArgs: []string{"bash", "fish", "powershell", "zsh"},
		Example:   eg,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				switch args[0] {
				case "bash":
					return cmd.Root().GenBashCompletion(os.Stdout)
				case "fish":
					return cmd.Root().GenFishCompletion(os.Stdout, true)
				case "powershell":
					return cmd.Root().GenPowerShellCompletion(os.Stdout)
				case "zsh":
					return cmd.Root().GenZshCompletion(os.Stdout)
				default:
					return errors.New("'bash', 'fish', 'powershell' or 'zsh' shell completion is supported")
				}
			} else {
				return errors.New("Only one argument can be provided, either 'bash', 'fish', 'powershell' or 'zsh'")
			}
		},
	}
}
