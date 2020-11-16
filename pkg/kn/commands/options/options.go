// Copyright © 2020 The Knative Authors
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

package options

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/templates"
)

// NewCmdOptions implements the options command
func NewOptionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options",
		Short: "Print the list of flags inherited by all commands",
		Long:  "Print the list of flags inherited by all commands",
		Example: `# Print flags inherited by all commands
kn options`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.Usage()
		},
		// Be quiet
		SilenceErrors: true,
		SilenceUsage:  true,
		// Allow all options
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true}, // wokeignore:rule=whitelist // TODO(#1031)
	}
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return errors.Errorf("%s for '%s'", err.Error(), c.CommandPath())
	})
	cmd.SetUsageFunc(templates.NewGlobalOptionsFunc())
	cmd.SetHelpFunc(func(command *cobra.Command, args []string) {
		templates.NewGlobalOptionsFunc()(command)
	})
	return cmd
}
