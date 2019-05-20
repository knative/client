package commands

import (
	"os"

	"github.com/spf13/cobra"
)

type CompletionFlags struct {
	Zsh bool
}

func NewCompletionCommand(p *KnParams) *cobra.Command {
	var completionFlags CompletionFlags

	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Output shell completion code (default Bash)",
		Run: func(cmd *cobra.Command, args []string) {
			if completionFlags.Zsh {
				cmd.Root().GenZshCompletion(os.Stdout)
			} else {
				cmd.Root().GenBashCompletion(os.Stdout)
			}
		},
	}

	completionCmd.Flags().BoolVar(&completionFlags.Zsh, "zsh", false, "Generates completion code for Zsh shell.")
	return completionCmd
}
