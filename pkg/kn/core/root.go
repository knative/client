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

package core

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/kn/commands/revision"
	"github.com/knative/client/pkg/kn/commands/route"
	"github.com/knative/client/pkg/kn/commands/service"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var cfgFile string
var kubeCfgFile string

// NewKnCommand creates new rootCmd represents the base command when called without any subcommands
func NewKnCommand(params ...commands.KnParams) *cobra.Command {
	var p *commands.KnParams
	if len(params) == 0 {
		p = &commands.KnParams{}
	} else if len(params) == 1 {
		p = &params[0]
	} else {
		panic("Too many params objects to NewKnCommand")
	}
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
	}
	if p.Output != nil {
		rootCmd.SetOutput(p.Output)
	}
	rootCmd.PersistentFlags().StringVar(&commands.CfgFile, "config", "", "config file (default is $HOME/.kn/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&p.KubeCfgPath, "kubeconfig", "", "kubectl config file (default is $HOME/.kube/config)")

	rootCmd.AddCommand(service.NewServiceCommand(p))
	rootCmd.AddCommand(revision.NewRevisionCommand(p))
	rootCmd.AddCommand(route.NewRouteCommand(p))
	rootCmd.AddCommand(commands.NewCompletionCommand(p))
	rootCmd.AddCommand(commands.NewVersionCommand(p))

	// Deal with empty and unknown sub command groups
	EmptyAndUnknownSubCommands(rootCmd)

	// For glog parse error.
	flag.CommandLine.Parse([]string{})
	return rootCmd
}

// EmptyAndUnknownSubCommands adds a RunE to all commands that are groups to
// deal with errors when called with empty or unknown sub command
func EmptyAndUnknownSubCommands(cmd *cobra.Command) {
	for _, childCmd := range cmd.Commands() {
		if childCmd.HasSubCommands() && childCmd.RunE == nil {
			childCmd.RunE = func(aCmd *cobra.Command, args []string) error {
				aCmd.Help()
				fmt.Println()

				if len(args) == 0 {
					return errors.New(fmt.Sprintf("please provide a valid sub-command for \"kn %s\"", aCmd.Name()))
				} else {
					return errors.New(fmt.Sprintf("unknown sub-command \"%s\" for \"kn %s\"", args[0], aCmd.Name()))
				}
			}
		}

		// recurse to deal with child commands that are themselves command groups
		EmptyAndUnknownSubCommands(childCmd)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if commands.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(commands.CfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kn" (without extension).
		viper.AddConfigPath(path.Join(home, ".kn"))
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
