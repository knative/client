package commands

import (
	"os"

	"github.com/spf13/cobra"
)

func NewCompletionCommand(p *KnParams) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Output bash completion code",
		Run:   completionAction,
	}
	return completionCmd
}

func completionAction(cmd *cobra.Command, args []string) {
	cmd.Root().GenBashCompletion(os.Stdout)
}
