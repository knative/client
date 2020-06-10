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

package root

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/completion"
	"knative.dev/client/pkg/kn/commands/plugin"
	"knative.dev/client/pkg/kn/commands/revision"
	"knative.dev/client/pkg/kn/commands/route"
	"knative.dev/client/pkg/kn/commands/service"
	"knative.dev/client/pkg/kn/commands/source"
	"knative.dev/client/pkg/kn/commands/trigger"
	"knative.dev/client/pkg/kn/commands/version"
	"knative.dev/client/pkg/kn/config"
	"knative.dev/client/pkg/kn/flags"
)

// NewRootCommand creates the default `kn` command with a default plugin handler
func NewRootCommand() (*cobra.Command, error) {
	p := &commands.KnParams{}
	p.Initialize()

	rootCmd := &cobra.Command{
		Use:   "kn",
		Short: "Knative client",
		Long: `Manage your Knative building blocks:

* Serving: Manage your services and release new software to them.
* Eventing: Manage event subscriptions and channels. Connect up event sources.`,

		// Disable docs header
		DisableAutoGenTag: true,

		// Affects children as well
		SilenceUsage: true,

		// Prevents Cobra from dealing with errors as we deal with them in main.go
		SilenceErrors: true,

		// Validate our boolean configs
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return flags.ReconcileBoolFlags(cmd.Flags())
		},
	}
	if p.Output != nil {
		rootCmd.SetOut(p.Output)
	}

	// Bootstrap flags (rebinding to avoid errors when parsing the full commands)
	config.AddBootstrapFlags(rootCmd.PersistentFlags())

	// Global flags
	rootCmd.PersistentFlags().StringVar(&p.KubeCfgPath, "kubeconfig", "", "kubectl configuration file (default: ~/.kube/config)")
	flags.AddBothBoolFlags(rootCmd.PersistentFlags(), &p.LogHTTP, "log-http", "", false, "log http traffic")

	// root child commands
	rootCmd.AddCommand(service.NewServiceCommand(p))
	rootCmd.AddCommand(revision.NewRevisionCommand(p))
	rootCmd.AddCommand(plugin.NewPluginCommand(p))
	rootCmd.AddCommand(route.NewRouteCommand(p))
	rootCmd.AddCommand(completion.NewCompletionCommand(p))
	rootCmd.AddCommand(version.NewVersionCommand(p))
	rootCmd.AddCommand(source.NewSourceCommand(p))
	rootCmd.AddCommand(trigger.NewTriggerCommand(p))

	// Initialize default `help` cmd early to prevent unknown command errors
	rootCmd.InitDefaultHelpCmd()

	// Check that command groups can't execute and that leaf commands don't h
	err := validateCommandStructure(rootCmd)
	if err != nil {
		return nil, err
	}

	// Wrap usage.
	fitUsageMessageToTerminalWidth(rootCmd)

	// For glog parse error. TOO: Check why this is needed
	flag.CommandLine.Parse([]string{})
	return rootCmd, nil
}

// Verify that command groups are not executable and that leaf commands have a run function
func validateCommandStructure(cmd *cobra.Command) error {
	for _, childCmd := range cmd.Commands() {
		if childCmd.HasSubCommands() {
			if childCmd.RunE != nil || childCmd.Run != nil {
				return errors.Errorf("internal: command group '%s' must not enable any direct logic, only leaf commands are allowed to take actions", childCmd.Name())
			}

			subCommands := childCmd.Commands()
			name := childCmd.Name()
			childCmd.RunE = func(aCmd *cobra.Command, args []string) error {
				subText := fmt.Sprintf("Available sub-commands: %s", strings.Join(ExtractSubCommandNames(subCommands), ", "))
				if len(args) == 0 {
					return fmt.Errorf("no sub-command given for 'kn %s'. %s", name, subText)
				}
				return fmt.Errorf("unknown sub-command '%s' for 'kn %s'. %s", args[0], aCmd.Name(), subText)
			}
		}

		// recurse to deal with child commands that are themselves command groups
		err := validateCommandStructure(childCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func fitUsageMessageToTerminalWidth(rootCmd *cobra.Command) {
	width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err == nil {
		newUsage := strings.ReplaceAll(rootCmd.UsageTemplate(), "FlagUsages ",
			fmt.Sprintf("FlagUsagesWrapped %d ", width))
		rootCmd.SetUsageTemplate(newUsage)
	}
}

// ExtractSubCommandNames extracts the names of all sub commands of a given command
func ExtractSubCommandNames(cmds []*cobra.Command) []string {
	var ret []string
	for _, subCmd := range cmds {
		ret = append(ret, subCmd.Name())
	}
	return ret
}
