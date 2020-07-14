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

package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/config"
	"knative.dev/client/pkg/kn/plugin"
	"knative.dev/client/pkg/kn/root"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	err := run(os.Args[1:])
	if err != nil && len(os.Args) > 1 {
		printError(err)
		// This is the only point from where to exit when an error occurs
		os.Exit(1)
	}
}

// runError is used when during the execution of a command/plugin an error occurs and
// so no extra usage message should be shown.
type runError struct {
	err error
}

// Error implements the error() interface
func (e *runError) Error() string {
	return e.err.Error()
}

// Run the main program. Args are the args as given on the command line (excluding the program name itself)
func run(args []string) error {
	// Parse config & plugin flags early to read in configuration file
	// and bind to viper. After that you can access all configuration and
	// global options via methods on config.GlobalConfig
	err := config.BootstrapConfig()
	if err != nil {
		return err
	}

	// Strip of all flags to get the non-flag commands only
	commands, err := stripFlags(args)
	if err != nil {
		return err
	}

	// Find plugin with the commands arguments
	pluginManager := plugin.NewManager(config.GlobalConfig.PluginsDir(), config.GlobalConfig.LookupPluginsInPath())
	plugin, err := pluginManager.FindPlugin(commands)
	if err != nil {
		return err
	}

	// Create kn root command and all sub-commands
	rootCmd, err := root.NewRootCommand(pluginManager.HelpTemplateFuncs())
	if err != nil {
		return err
	}

	if plugin != nil {
		// Validate & Execute plugin
		err = validatePlugin(rootCmd, plugin)
		if err != nil {
			return err
		}

		err := plugin.Execute(argsWithoutCommands(args, plugin.CommandParts()))
		if err != nil {
			return &runError{err: err}
		}
		return nil
	} else {
		// Validate args for root command
		err = validateRootCommand(rootCmd)
		if err != nil {
			return err
		}
		// Execute kn root command, args are taken from os.Args directly
		return rootCmd.Execute()
	}
}

// Get only the args provided but no options. The extraction
// process is a bit tricky as Cobra doesn't provide such
// functionality out of the box
func stripFlags(args []string) ([]string, error) {
	// Store all command
	commandsFound := &[]string{}

	// Use a canary command that allows all options and only extracts
	// commands. Doesn't work with arbitrary boolean flags but is good enough
	// for us here
	extractCommand := cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			for _, arg := range args {
				*commandsFound = append(*commandsFound, arg)
			}
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// No help and usage functions to prin
	extractCommand.SetHelpFunc(func(*cobra.Command, []string) {})
	extractCommand.SetUsageFunc(func(*cobra.Command) error { return nil })

	// Filter out --help and -h options to avoid special treatment which we don't
	// need here
	extractCommand.SetArgs(filterHelpOptions(args))

	// Adding all global flags here
	config.AddBootstrapFlags(extractCommand.Flags())

	// Allow all options
	extractCommand.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}

	// Execute to get to the command args
	err := extractCommand.Execute()
	if err != nil {
		return nil, err
	}
	return *commandsFound, nil
}

// Strip all plugin commands before calling out to the plugin
func argsWithoutCommands(cmdArgs []string, pluginCommandsParts []string) []string {
	var ret []string
	for _, arg := range cmdArgs {
		if len(pluginCommandsParts) > 0 && pluginCommandsParts[0] == arg {
			pluginCommandsParts = pluginCommandsParts[1:]
			continue
		}
		ret = append(ret, arg)
	}
	return ret
}

// Remove all help options
func filterHelpOptions(args []string) []string {
	var ret []string
	for _, arg := range args {
		if arg != "-h" && arg != "--help" {
			ret = append(ret, arg)
		}
	}
	return ret
}

// Check if the plugin collides with any command specified in the root command
func validatePlugin(root *cobra.Command, plugin plugin.Plugin) error {
	// Check if a given plugin can be identified as a command
	cmd, args, err := root.Find(plugin.CommandParts())

	if err == nil {
		if !cmd.HasSubCommands() || // a leaf command can't be overridden
			cmd.HasSubCommands() && len(args) == 0 { // a group can't be overridden either
			return errors.Errorf("plugin %s is overriding built-in command '%s' which is not allowed", plugin.Path(), strings.Join(plugin.CommandParts(), " "))
		}
	}
	return nil
}

// Check whether an unknown sub-command is addressed and return an error if this is the case
// Needs to be called after the plugin has been extracted (as a plugin name can also lead to
// an unknown sub command error otherwise)
func validateRootCommand(cmd *cobra.Command) error {
	foundCmd, innerArgs, err := cmd.Find(os.Args[1:])
	if err == nil && foundCmd.HasSubCommands() && len(innerArgs) > 0 {
		argsWithoutFlags, err := stripFlags(innerArgs)
		if len(argsWithoutFlags) > 0 || err != nil {
			return errors.Errorf("unknown sub-command '%s' for '%s'. Available sub-commands: %s", innerArgs[0], foundCmd.CommandPath(), strings.Join(root.ExtractSubCommandNames(foundCmd.Commands()), ", "))
		}
		// If no args where given (only flags), then fall through to execute the command itself, which leads to
		// a more appropriate error message
	}
	return nil
}

// printError prints out any given error
func printError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", cleanupErrorMessage(err.Error()))
	var runError *runError
	if !errors.As(err, &runError) {
		// Print help hint only if its not a runError occurred when executing a command
		fmt.Fprintf(os.Stderr, "Run '%s --help' for usage\n", extractCommandPathFromErrorMessage(err.Error(), os.Args[0]))
	}
}

// extractCommandPathFromErrorMessage tries to extract the command name from an error message
// by checking a pattern like 'kn service' in the error message. If not found, return the
// base command name.
func extractCommandPathFromErrorMessage(errorMsg string, arg0 string) string {
	extractPattern := regexp.MustCompile(fmt.Sprintf("'(%s\\s.+?)'", arg0))
	command := extractPattern.FindSubmatch([]byte(errorMsg))
	if command != nil {
		return string(command[1])
	}
	return arg0
}

// cleanupErrorMessage remove any redundance content of an error message
func cleanupErrorMessage(msg string) string {
	regexp := regexp.MustCompile("(?i)^error:\\s*")
	return string(regexp.ReplaceAll([]byte(msg), []byte("")))
}
