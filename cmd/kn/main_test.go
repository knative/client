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
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"knative.dev/client/pkg/kn/plugin"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"

	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/kn/config"
	"knative.dev/client/pkg/kn/root"
	"knative.dev/client/pkg/util"
)

func TestValidatePlugin(t *testing.T) {

	// Build up simple command hierarchy
	root := cobra.Command{}
	one := &cobra.Command{Use: "one"}
	one.AddCommand(&cobra.Command{Use: "eins"}, &cobra.Command{Use: "zwei"})
	two := &cobra.Command{Use: "two"}
	root.AddCommand(one, two)

	data := []struct {
		givenPluginCommandParts []string
		expectedErrors          []string
	}{
		{
			// Allowed because it add a new top-level plugin
			[]string{"test"},
			nil,
		},
		{
			// Allowed because it adds to an existing command-group
			[]string{"one", "drei"},
			nil,
		},
		{
			// Forbidden because it overrides an command-group
			[]string{"one"},
			[]string{"pluginPath", "one"},
		},
		{
			// Forbidden because it overrides a leaf-command
			[]string{"one", "zwei"},
			[]string{"pluginPath", "one", "zwei"},
		},
		{
			// Forbidden because it would mis-use a leaf-comman to a command-group
			[]string{"one", "zwei", "trois"},
			[]string{"pluginPath", "one", "zwei"},
		},
		{
			// Forbidden because it overrides a (top-level) leaf-command
			[]string{"two"},
			[]string{"pluginPath", "two"},
		},
		{
			// Forbidden because it would add to a leaf command
			[]string{"two", "deux", "and", "more"},
			[]string{"pluginPath", "two", "deux"},
		},
	}

	for i, d := range data {
		step := fmt.Sprintf("Check %d", i)
		err := validatePlugin(&root, commandPartsOnlyPlugin(d.givenPluginCommandParts))
		if len(d.expectedErrors) == 0 {
			assert.NilError(t, err, step)
		} else {
			assert.Assert(t, err != nil, step)
			assert.Assert(t, util.ContainsAll(err.Error(), d.expectedErrors...), step)
		}
	}

}

// Used above for wrapping the command part to check
type commandPartsOnlyPlugin []string

func (f commandPartsOnlyPlugin) CommandParts() []string       { return f }
func (f commandPartsOnlyPlugin) Name() string                 { return "" }
func (f commandPartsOnlyPlugin) Execute(args []string) error  { return nil }
func (f commandPartsOnlyPlugin) Description() (string, error) { return "", nil }
func (f commandPartsOnlyPlugin) Path() string                 { return "pluginPath" }

func TestArgsWithoutCommands(t *testing.T) {
	data := []struct {
		givenCmdArgs            []string
		givenPluginCommandParts []string
		expectedResult          []string
	}{
		{
			[]string{"--option", "val", "one", "second", "rest"},
			[]string{"one", "second"},
			[]string{"--option", "val", "rest"},
		},
		{
			[]string{"--option", "val", "one", "second", "rest"},
			[]string{"second", "one"},
			[]string{"--option", "val", "one", "rest"},
		},
		{
			[]string{"--option", "val", "one", "second", "third", "one", "rest"},
			[]string{"second", "one"},
			[]string{"--option", "val", "one", "third", "rest"},
		},
	}
	for _, d := range data {
		result := argsWithoutCommands(d.givenCmdArgs, d.givenPluginCommandParts)
		assert.DeepEqual(t, result, d.expectedResult)
	}
}

func TestUnknownCommands(t *testing.T) {
	oldArgs := os.Args
	defer (func() {
		os.Args = oldArgs
	})()

	data := []struct {
		givenCmdArgs  []string
		commandPath   []string
		expectedError []string
	}{
		{
			[]string{"service", "unknown", "test", "--scale-min=0"},
			[]string{"service"},
			[]string{"unknown sub-command", "unknown"},
		},
		{
			[]string{"service", "--foo=bar"},
			[]string{"service"},
			[]string{},
		},
		{
			[]string{"source", "ping", "blub", "--foo=bar"},
			[]string{"source", "ping"},
			[]string{"unknown sub-command", "blub"},
		},
	}
	for _, d := range data {
		args := append([]string{"kn"}, d.givenCmdArgs...)
		rootCmd, err := root.NewRootCommand(nil)
		rootCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true} // wokeignore:rule=whitelist // TODO(#1031)
		os.Args = args
		assert.NilError(t, err)
		err = validateRootCommand(rootCmd)
		if len(d.expectedError) == 0 {
			assert.NilError(t, err)
			continue
		}
		assert.Assert(t, err != nil)
		assert.Assert(t, util.ContainsAll(err.Error(), d.expectedError...))
		cmd, _, e := rootCmd.Find(d.commandPath)
		assert.NilError(t, e)
		for _, sub := range cmd.Commands() {
			assert.ErrorContains(t, err, sub.Name())
		}
	}

}

func TestStripFlags(t *testing.T) {

	data := []struct {
		givenArgs        []string
		expectedCommands []string
		expectedError    string
	}{
		{
			[]string{"test", "-h", "second", "--bla"},
			[]string{"test", "second"},
			"",
		},
		{
			[]string{"--help", "test"},
			[]string{"test"},
			"",
		},
		{
			[]string{"--unknown-option", "bla", "test", "second"},
			[]string{"test", "second"},
			"",
		},
		{
			[]string{"--lookup-plugins", "bla", "test", "second"},
			[]string{"bla", "test", "second"},
			"",
		},
		{
			[]string{"--config-file", "bla", "test", "second"},
			[]string{"test", "second"},
			"",
		},
		{
			[]string{"test"},
			[]string{"test"},
			"",
		},
	}

	for i, f := range data {
		step := fmt.Sprintf("Check %d", i)
		cmd := &cobra.Command{FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true}} // wokeignore:rule=whitelist // TODO(#1031)
		config.AddBootstrapFlags(cmd.Flags())
		commands, err := stripFlags(cmd, f.givenArgs)
		assert.DeepEqual(t, commands, f.expectedCommands)
		if f.expectedError != "" {
			assert.ErrorContains(t, err, f.expectedError, step)
		} else {
			assert.NilError(t, err, step)
		}
	}

	t.Log("checking error case for stripFlags")
	cmd, err := root.NewRootCommand(nil)
	assert.NilError(t, err)
	_, err = stripFlags(cmd, []string{"--config"})
	assert.ErrorContains(t, err, "needs an argument")
}

func TestPrintError(t *testing.T) {
	data := []struct {
		given    string
		expected string
	}{
		{
			"unknown sub-command blub",
			"Error: unknown sub-command blub",
		},
		{
			"error: unknown type blub",
			"Error: unknown type blub",
		},
	}
	for _, d := range data {
		capture := test.CaptureOutput(t)
		printError(errors.New(d.given))
		stdOut, errOut := capture.Close()

		assert.Equal(t, stdOut, "")
		assert.Assert(t, strings.Contains(errOut, d.expected))
		assert.Assert(t, util.ContainsAll(errOut, "Run", "--help", "usage"))
	}
}

func TestRunWithPluginError(t *testing.T) {
	data := []struct {
		given    string
		expected string
	}{
		{
			"exit status 1",
			"Error: exit status 1",
		},
	}
	for _, d := range data {
		capture := test.CaptureOutput(t)
		// displayHelp argument is false for plugin error
		printError(&runError{errors.New(d.given)})
		stdOut, errOut := capture.Close()

		assert.Equal(t, stdOut, "")
		assert.Assert(t, strings.Contains(errOut, d.expected))
		// check that --help message isn't displayed
		assert.Assert(t, util.ContainsNone(errOut, "Run", "--help", "usage"))
	}
}

func TestRunWithExit(t *testing.T) {
	oldArgs := os.Args
	defer (func() {
		os.Args = oldArgs
	})()
	testCases := []struct {
		args           []string
		expectedOut    []string
		expectedErrOut []string
		exitCode       int
	}{
		{
			[]string{"kn", "version"},
			[]string{"version", "build", "git"},
			[]string{""},
			0,
		},
		{
			[]string{"kn", "non-existing"},
			[]string{""},
			[]string{"unknown", "command"},
			1,
		},
		{
			[]string{"kn", "service", "foo"},
			[]string{""},
			[]string{"unknown", "sub-command"},
			1,
		},
		{
			[]string{"kn", "service", "create", "foo", "--foo"},
			[]string{""},
			[]string{"unknown", "flag"},
			1,
		},
	}
	for _, tc := range testCases {
		capture := test.CaptureOutput(t)
		os.Args = tc.args
		exitCode := runWithExit(tc.args[1:])
		out, errOut := capture.Close()
		assert.Equal(t, exitCode, tc.exitCode)
		assert.Assert(t, util.ContainsAllIgnoreCase(out, tc.expectedOut...))
		assert.Assert(t, util.ContainsAllIgnoreCase(errOut, tc.expectedErrOut...))
	}
}

func TestExtractCommandPathFromErrorMessage(t *testing.T) {
	for _, d := range []struct{ arg0, errMsg, expected string }{
		{"kn", "Invalid argument for 'kn service'", "kn service"},
		{"C:\\Users\\hudson.DESKTOP-T61GB27\\Documents\\foo-with-revisions", "Invalid argument for 'C:\\Users\\hudson.DESKTOP-T61GB27\\Documents\\foo-with-revisions test'", "C:\\Users\\hudson.DESKTOP-T61GB27\\Documents\\foo-with-revisions test"},
	} {
		assert.Equal(t, extractCommandPathFromErrorMessage(d.errMsg, d.arg0), d.expected)
	}
}

type internalPlugin struct {
	executeError func() error
	commandParts []string
}

func (p internalPlugin) CommandParts() []string       { return p.commandParts }
func (p internalPlugin) Name() string                 { return "" }
func (p internalPlugin) Execute(args []string) error  { return p.executeError() }
func (p internalPlugin) Description() (string, error) { return "", nil }
func (p internalPlugin) Path() string                 { return "" }

func TestRun(t *testing.T) {
	oldArgs := os.Args
	defer (func() {
		os.Args = oldArgs
	})()

	testCases := []struct {
		args           []string
		expectedOut    []string
		expectedErrOut []string
		plugin         plugin.Plugin
	}{
		{
			[]string{"kn", "version"},
			[]string{"version", "build", "git"},
			[]string{},
			nil,
		},
		{
			[]string{"kn", "non-existing"},
			[]string{},
			[]string{"unknown", "command"},
			nil,
		},
		{
			[]string{"kn", "service", "foo"},
			[]string{},
			[]string{"unknown", "sub-command"},
			nil,
		},
		{
			[]string{"kn", "service", "create", "foo", "--foo"},
			[]string{},
			[]string{"unknown", "flag"},
			nil,
		},
		// Internal plugins
		{
			[]string{"kn", "foo"},
			[]string{"OK", "plugin", "out"},
			[]string{},

			&internalPlugin{
				executeError: func() error {
					fmt.Println("OK plugin out")
					return nil
				},
				commandParts: []string{"foo"},
			},
		},
		{
			[]string{"kn", "service", "create"},
			[]string{},
			[]string{"plugin", "overriding", "'service create'"},
			&internalPlugin{
				executeError: nil,
				commandParts: []string{"service", "create"},
			},
		},
		{
			[]string{"kn", "foo", "bar"},
			[]string{},
			[]string{"internal", "plugin", "error"},
			&internalPlugin{
				executeError: func() error {
					return errors.New("internal plugin error")
				},
				commandParts: []string{"foo", "bar"},
			},
		},
	}
	for _, tc := range testCases {
		os.Args = tc.args
		if tc.plugin != nil {
			plugin.InternalPlugins = plugin.PluginList{}
			plugin.InternalPlugins = append(plugin.InternalPlugins, tc.plugin)
		}
		capture := test.CaptureOutput(t)
		err := run(tc.args[1:])
		out, _ := capture.Close()
		if len(tc.expectedErrOut) > 0 {
			assert.Assert(t, err != nil)
			assert.Assert(t, util.ContainsAllIgnoreCase(err.Error(), tc.expectedErrOut...))
		} else {
			assert.NilError(t, err)
			assert.Assert(t, util.ContainsAllIgnoreCase(out, tc.expectedOut...))
		}

	}
}
