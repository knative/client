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
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/broker"
	"knative.dev/client/pkg/kn/commands/completion"
	"knative.dev/client/pkg/kn/commands/options"
	"knative.dev/client/pkg/kn/commands/plugin"
	"knative.dev/client/pkg/kn/commands/revision"
	"knative.dev/client/pkg/kn/commands/route"
	"knative.dev/client/pkg/kn/commands/service"
	"knative.dev/client/pkg/kn/commands/source"
	"knative.dev/client/pkg/kn/commands/trigger"
	"knative.dev/client/pkg/kn/commands/version"
	"knative.dev/client/pkg/kn/config"
	"knative.dev/client/pkg/kn/flags"
	"knative.dev/client/pkg/templates"
)

// NewRootCommand creates the default `kn` command with a default plugin handler
func NewRootCommand(helpFuncs *template.FuncMap) (*cobra.Command, error) {
	p := &commands.KnParams{}
	p.Initialize()

	rootCmd := &cobra.Command{
		Use:   "kn",
		Short: "kn manages Knative Serving and Eventing resources",
		Long: `kn is the command line interface for managing Knative Serving and Eventing resources

 Find more information about Knative at: https://knative.dev`,

		// Disable docs header
		DisableAutoGenTag: true,

		// Disable usage & error printing from cobra as we
		// are handling all error output on our own
		SilenceUsage:  true,
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

	// Grouped commands
	groups := templates.CommandGroups{
		{
			Header: "Serving Commands:",
			Commands: []*cobra.Command{
				service.NewServiceCommand(p),
				revision.NewRevisionCommand(p),
				route.NewRouteCommand(p),
			},
		},
		{
			Header: "Eventing Commands:",
			Commands: []*cobra.Command{
				source.NewSourceCommand(p),
				broker.NewBrokerCommand(p),
				trigger.NewTriggerCommand(p),
			},
		},
		{
			Header: "Other Commands:",
			Commands: []*cobra.Command{
				plugin.NewPluginCommand(p),
				completion.NewCompletionCommand(p),
				version.NewVersionCommand(p),
			},
		},
	}
	// Add all commands to the root command, flat
	groups.AddTo(rootCmd)

	// Initialize default `help` cmd early to prevent unknown command errors
	groups.SetRootUsage(rootCmd, helpFuncs)

	// Add the "options" commands for showing all global options
	rootCmd.AddCommand(options.NewOptionsCommand())

	// Check that command groups can't execute and that leaf commands don't h
	err := validateCommandStructure(rootCmd)
	if err != nil {
		return nil, err
	}

	// Add some command context when flags can not be parsed
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return errors.Errorf("%s for '%s'", err.Error(), c.CommandPath())
	})

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

// ExtractSubCommandNames extracts the names of all sub commands of a given command
func ExtractSubCommandNames(cmds []*cobra.Command) []string {
	var ret []string
	for _, subCmd := range cmds {
		ret = append(ret, subCmd.Name())
	}
	return ret
}
