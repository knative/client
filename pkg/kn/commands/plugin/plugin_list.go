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

package plugin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	homedir "github.com/mitchellh/go-homedir"
)

// ValidPluginFilenamePrefixes controls the prefix for all kn plugins
var ValidPluginFilenamePrefixes = []string{"kn"}

// NewPluginListCommand creates a new `kn plugin list` command
func NewPluginListCommand(p *commands.KnParams) *cobra.Command {
	pluginFlags := PluginFlags{
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}

	pluginListCommand := &cobra.Command{
		Use:   "list",
		Short: "List all visible plugin executables",
		Long: `List all visible plugin executables.

Available plugin files are those that are:
- executable
- begin with "kn-
- anywhere on the path specified in Kn's config pluginDir variable, which:
  * can be overridden with the --plugin-dir flag`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := pluginFlags.complete(cmd)
			if err != nil {
				return err
			}

			err = pluginFlags.run()
			if err != nil {
				return err
			}

			return nil
		},
	}

	AddPluginFlags(pluginListCommand)
	BindPluginsFlagToViper(pluginListCommand)

	pluginFlags.AddPluginFlags(pluginListCommand)

	return pluginListCommand
}

// ExpandPath to a canonical path (need to see if Golang has a better option)
func ExpandPath(path string) (string, error) {
	if strings.Contains(path, "~") {
		var err error
		path, err = expandHomeDir(path)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

// Private

func (o *PluginFlags) complete(cmd *cobra.Command) error {
	o.Verifier = &CommandOverrideVerifier{
		Root:        cmd.Root(),
		SeenPlugins: make(map[string]string, 0),
	}

	pluginPath, err := ExpandPath(commands.Cfg.PluginsDir)
	if err != nil {
		return err
	}

	if commands.Cfg.LookupPluginsInPath {
		pluginPath = pluginPath + string(os.PathListSeparator) + os.Getenv("PATH")
	}

	o.PluginPaths = filepath.SplitList(pluginPath)

	return nil
}

func (o *PluginFlags) run() error {
	isFirstFile := true
	pluginErrors := []error{}
	pluginWarnings := 0

	for _, dir := range uniquePathsList(o.PluginPaths) {
		if dir == "" {
			continue
		}

		// Treat a missing dir as just something we skip - meaning
		// it's ok to include a path that might not exist yet or you
		// don't have access to - so just skip it
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		files, err := ioutil.ReadDir(dir)
		// However, error reading a dir that does exist is an issue
		if err != nil {
			if _, ok := err.(*os.PathError); ok {
				fmt.Fprintf(o.ErrOut, "Unable read directory '%s' from your plugins path: %v. Skipping...", dir, err)
				continue
			}

			pluginErrors = append(pluginErrors, fmt.Errorf("error: unable to read directory '%s' from your plugin path: %v", dir, err))
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !hasValidPrefix(f.Name(), ValidPluginFilenamePrefixes) {
				continue
			}

			if isFirstFile {
				fmt.Fprintf(o.ErrOut, "The following compatible plugins are available, using options:\n")
				fmt.Fprintf(o.ErrOut, "  - plugins dir: '%s'\n", commands.Cfg.PluginsDir)
				fmt.Fprintf(o.ErrOut, "  - lookup plugins in path: '%t'\n\n", commands.Cfg.LookupPluginsInPath)
				isFirstFile = false
			}

			pluginPath := f.Name()
			if !o.NameOnly {
				pluginPath = filepath.Join(dir, pluginPath)
			}

			fmt.Fprintf(o.Out, "%s\n", pluginPath)
			if errs := o.Verifier.Verify(filepath.Join(dir, f.Name())); len(errs) != 0 {
				for _, err := range errs {
					fmt.Fprintf(o.ErrOut, "  - %s\n", err)
					pluginWarnings++
				}
			}
		}
	}

	if pluginWarnings > 0 {
		if pluginWarnings == 1 {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: one plugin warning was found"))
		} else {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: %v plugin warnings were found", pluginWarnings))
		}
	}
	if len(pluginErrors) > 0 {
		fmt.Fprintln(o.ErrOut)
		errs := bytes.NewBuffer(nil)
		for _, e := range pluginErrors {
			fmt.Fprintln(errs, e)
		}
		return fmt.Errorf("%s", errs.String())
	}

	return nil
}

// Private

// expandHomeDir replaces the ~ with the home directory value
func expandHomeDir(path string) (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}

	return strings.Replace(path, "~", home, -1), nil
}

// uniquePathsList deduplicates a given slice of strings without
// sorting or otherwise altering its order in any way.
func uniquePathsList(paths []string) []string {
	seen := map[string]bool{}
	newPaths := []string{}
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		newPaths = append(newPaths, p)
	}
	return newPaths
}

func hasValidPrefix(filepath string, validPrefixes []string) bool {
	for _, prefix := range validPrefixes {
		if !strings.HasPrefix(filepath, prefix+"-") {
			continue
		}
		return true
	}
	return false
}
