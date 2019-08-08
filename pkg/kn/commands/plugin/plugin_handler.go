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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mitchellh/go-homedir"
)

// PluginHandler is capable of parsing command line arguments
// and performing executable filename lookups to search
// for valid plugin files, and execute found plugins.
type PluginHandler interface {
	// exists at the given filename, or a boolean false.
	// Lookup will iterate over a list of given prefixes
	// in order to recognize valid plugin filenames.
	// The first filepath to match a prefix is returned.
	Lookup(name string) (string, bool)
	// Execute receives an executable's filepath, a slice
	// of arguments, and a slice of environment variables
	// to relay to the executable.
	Execute(executablePath string, cmdArgs, environment []string) error
}

// DefaultPluginHandler implements PluginHandler
type DefaultPluginHandler struct {
	ValidPrefixes       []string
	PluginsDir          string
	LookupPluginsInPath bool
}

// NewDefaultPluginHandler instantiates the DefaultPluginHandler with a list of
// given filename prefixes used to identify valid plugin filenames.
func NewDefaultPluginHandler(validPrefixes []string, pluginsDir string, lookupPluginsInPath bool) *DefaultPluginHandler {
	return &DefaultPluginHandler{
		ValidPrefixes:       validPrefixes,
		PluginsDir:          pluginsDir,
		LookupPluginsInPath: lookupPluginsInPath,
	}
}

// Lookup implements PluginHandler
func (h *DefaultPluginHandler) Lookup(name string) (string, bool) {
	for _, prefix := range h.ValidPrefixes {
		pluginPath := fmt.Sprintf("%s-%s", prefix, name)

		// Try to find plugin in pluginsDir
		pluginDir, err := homedir.Expand(h.PluginsDir)
		if err != nil {
			return "", false
		}

		pluginDirPluginPath := filepath.Join(pluginDir, pluginPath)
		_, err = os.Stat(pluginDirPluginPath)
		if !os.IsNotExist(err) {
			return pluginDirPluginPath, true
		}

		// No plugins found in pluginsDir, try in PATH of that's an option
		if h.LookupPluginsInPath {
			pluginPath, err = exec.LookPath(pluginPath)
			if err != nil {
				continue
			}

			if pluginPath != "" {
				return pluginPath, true
			}
		}
	}

	return "", false
}

// Execute implements PluginHandler
func (h *DefaultPluginHandler) Execute(executablePath string, cmdArgs, environment []string) error {
	return syscall.Exec(executablePath, cmdArgs, environment)
}

// HandlePluginCommand receives a pluginHandler and command-line arguments and attempts to find
// a plugin executable that satisfies the given arguments.
func HandlePluginCommand(pluginHandler PluginHandler, cmdArgs []string) error {
	remainingArgs := []string{}

	for idx := range cmdArgs {
		if strings.HasPrefix(cmdArgs[idx], "-") {
			continue
		}
		remainingArgs = append(remainingArgs, strings.Replace(cmdArgs[idx], "-", "_", -1))
	}

	foundBinaryPath := ""

	// attempt to find binary, starting at longest possible name with given cmdArgs
	for len(remainingArgs) > 0 {
		path, found := pluginHandler.Lookup(strings.Join(remainingArgs, "-"))
		if !found {
			remainingArgs = remainingArgs[:len(remainingArgs)-1]
			continue
		}

		foundBinaryPath = path
		break
	}

	if len(foundBinaryPath) == 0 {
		return errors.New("Could not find plugin to execute")
	}

	// invoke cmd binary relaying the current environment and args given
	// remainingArgs will always have at least one element.
	// execve will make remainingArgs[0] the "binary name".
	err := pluginHandler.Execute(foundBinaryPath, append([]string{foundBinaryPath}, cmdArgs[len(remainingArgs):]...), os.Environ())
	if err != nil {
		return err
	}

	return nil
}
