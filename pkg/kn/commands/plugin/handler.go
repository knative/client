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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	homedir "github.com/mitchellh/go-homedir"
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
// TODO: The current error handling is not optimal, and some errors may be lost. We should refactor the code in the future.
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
		if err == nil {
			return pluginDirPluginPath, true
		}

		// Try to match well-known file extensions on Windows
		if runtime.GOOS == "windows" {
			for _, ext := range []string{".bat", ".cmd", ".com", ".exe", ".ps1"} {
				pathWithExt := pluginDirPluginPath + ext
				_, err = os.Stat(pathWithExt)
				if err == nil {
					return pathWithExt, true
				}
			}
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
	if runtime.GOOS == "windows" {
		cmd := exec.Command(executablePath, cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = environment
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
	return syscall.Exec(executablePath, cmdArgs, environment)
}

// HandlePluginCommand receives a pluginHandler and command-line arguments and attempts to find
// a plugin executable that satisfies the given arguments.
func HandlePluginCommand(pluginHandler PluginHandler, cmdArgs []string) error {
	remainingArgs := []string{}

	for idx := range cmdArgs {
		cmdArg := cmdArgs[idx]
		if strings.HasPrefix(cmdArg, "-") {
			continue
		}
		remainingArgs = append(remainingArgs, strings.Replace(cmdArg, "-", "_", -1))
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

	// invoke cmd binary relaying the current environment and args given
	// remainingArgs will always have at least one element.
	// execve will make remainingArgs[0] the "binary name".
	if len(foundBinaryPath) != 0 {
		err := pluginHandler.Execute(foundBinaryPath, append([]string{foundBinaryPath}, cmdArgs[len(remainingArgs):]...), os.Environ())
		if err != nil {
			return err
		}
	}

	return nil
}
