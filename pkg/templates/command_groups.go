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

package templates

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// A command group is for grouping together commands
type CommandGroup struct {

	// Title for command group shown in help/usage messages
	Header string

	// List of commans for this group
	Commands []*cobra.Command
}

type CommandGroups []CommandGroup

// Add all commands from this group slice to the given command
func (g CommandGroups) AddTo(cmd *cobra.Command) {
	for _, group := range g {
		for _, sub := range group.Commands {
			cmd.AddCommand(sub)
		}
	}
}

// SetRootUsage sets our own help and usage function messages to the root command
func (g CommandGroups) SetRootUsage(rootCmd *cobra.Command) {
	engine := &templateEngine{
		RootCmd:       rootCmd,
		CommandGroups: g,
	}
	setHelpFlagsToSubCommands(rootCmd)
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return errors.Errorf("%s for '%s'", err.Error(), c.CommandPath())
	})
	rootCmd.SetUsageFunc(engine.usageFunc())
	rootCmd.SetHelpFunc(engine.helpFunc())
}

func setHelpFlagsToSubCommands(parent *cobra.Command) {
	for _, cmd := range parent.Commands() {
		if cmd.HasSubCommands() {
			setHelpFlagsToSubCommands(cmd)
		}
		cmd.DisableFlagsInUseLine = true
	}
}
