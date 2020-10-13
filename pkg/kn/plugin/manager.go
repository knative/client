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

package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Allow plugins to register to this slice for inlining
var InternalPlugins PluginList

// Interface describing a plugin
type Plugin interface {
	// Get the name of the plugin (the file name without extensions)
	Name() string

	// Execute the plugin with the given arguments
	Execute(args []string) error

	// Return a description of the plugin (if support by the plugin binary)
	Description() (string, error)

	// The command path leading to this plugin.
	// Eg. for a plugin "kn source github" this will be [ "source", "github" ]
	CommandParts() []string

	// Location of the plugin where it is stored in the filesystem
	Path() string
}

type Manager struct {
	// Dedicated plugin directory as configured
	pluginsDir string

	// Whether to check the OS path or not
	lookupInPath bool
}

type plugin struct {
	// Path to the plugin to execute
	path string

	// Name of the plugin
	name string

	// Commands leading to the execution of this plugin (e.g. "service","log" for a plugin kn-service-log)
	commandParts []string
}

// All extensions that are supposed to be windows executable
var windowsExecExtensions = []string{".bat", ".cmd", ".com", ".exe", ".ps1"}

// Used for sorting a list of plugins
type PluginList []Plugin

func (p PluginList) Len() int           { return len(p) }
func (p PluginList) Less(i, j int) bool { return p[i].Name() < p[j].Name() }
func (p PluginList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// === PluginManager =======================================================================

// NewManager creates a new manager for looking up plugins on the file system
func NewManager(pluginDir string, lookupInPath bool) *Manager {
	return &Manager{
		pluginsDir:   pluginDir,
		lookupInPath: lookupInPath,
	}
}

// FindPlugin checks if a plugin for the given parts exist and return it.
// The args given must not contain any options and contain only
// the commands (like in [ "source", "github" ] for a plugin called 'kn-source-github'
// The plugin with the most specific (longest) name is returned or nil if non is found.
// An error is returned if the lookup fails for some reason like an io error
func (manager *Manager) FindPlugin(parts []string) (Plugin, error) {
	if len(parts) == 0 {
		// No command given
		return nil, nil
	}

	// Try to find internal plugin fist
	plugin := lookupInternalPlugin(parts)
	if plugin != nil {
		return plugin, nil
	}

	// Try to find plugin in pluginsDir
	pluginDir, err := homedir.Expand(manager.pluginsDir)
	if err != nil {
		return nil, err
	}

	return findMostSpecificPluginInPath(pluginDir, parts, manager.lookupInPath)
}

// ListPlugins lists all plugins that can be found in the plugin directory or in the path (if configured)
func (manager *Manager) ListPlugins() (PluginList, error) {
	return manager.ListPluginsForCommandGroup(nil)
}

// ListPluginsForCommandGroup lists all plugins that can be found in the plugin directory or in the path (if configured),
// and which fits to a command group
func (manager *Manager) ListPluginsForCommandGroup(commandGroupParts []string) (PluginList, error) {

	// Initialize with list of internal plugins
	var plugins = append([]Plugin{}, filterPluginsByCommandGroup(InternalPlugins, commandGroupParts)...)

	dirs, err := manager.pluginLookupDirectories()
	if err != nil {
		return nil, err
	}

	// Examine all files in possible plugin directories
	hasSeen := make(map[string]bool)
	for _, pl := range plugins {
		hasSeen[pl.Name()] = true
	}
	for _, dir := range dirs {
		files, err := ioutil.ReadDir(dir)

		// Ignore non-existing directories
		if os.IsNotExist(err) {
			continue
		}

		// Check for plugins within given directory
		for _, f := range files {
			name := f.Name()
			if f.IsDir() {
				continue
			}
			if !strings.HasPrefix(name, "kn-") {
				continue
			}

			// Check if plugin matches a command group
			if !isPluginFileNamePartOfCommandGroup(commandGroupParts, f.Name()) {
				continue
			}

			// Ignore all plugins that are shadowed
			if seen, ok := hasSeen[name]; !ok || !seen {
				plugins = append(plugins, &plugin{
					path:         filepath.Join(dir, f.Name()),
					name:         stripWindowsExecExtensions(f.Name()),
					commandParts: extractPluginCommandFromFileName(f.Name()),
				})
				hasSeen[name] = true
			}
		}
	}

	// Sort according to name
	sort.Sort(PluginList(plugins))
	return plugins, nil
}

func filterPluginsByCommandGroup(plugins PluginList, commandGroupParts []string) PluginList {
	ret := PluginList{}
	for _, pl := range plugins {
		if isPartOfCommandGroup(commandGroupParts, pl.CommandParts()) {
			ret = append(ret, pl)
		}
	}
	return ret
}

func isPartOfCommandGroup(commandGroupParts []string, commandParts []string) bool {
	if len(commandParts) != len(commandGroupParts)+1 {
		return false
	}
	for i := range commandGroupParts {
		if commandParts[i] != commandGroupParts[i] {
			return false
		}
	}
	return true
}

func isPluginFileNamePartOfCommandGroup(commandGroupParts []string, pluginFileName string) bool {
	if commandGroupParts == nil {
		return true
	}

	commandParts := extractPluginCommandFromFileName(pluginFileName)
	if len(commandParts) != len(commandGroupParts)+1 {
		return false
	}
	for i := range commandGroupParts {
		if commandParts[i] != commandGroupParts[i] {
			return false
		}
	}
	return true
}

// PluginsDir returns the configured directory holding plugins
func (manager *Manager) PluginsDir() string {
	return manager.pluginsDir
}

// LookupInPath returns true if plugins should be also looked up within the path
func (manager *Manager) LookupInPath() bool {
	return manager.lookupInPath
}

// === Plugin ==============================================================================

// Execute the plugin with the given arguments
func (plugin *plugin) Execute(args []string) error {
	cmd := exec.Command(plugin.path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()
	return cmd.Run()
}

// Return a description of the plugin (if support by the plugin binary)
func (plugin *plugin) Description() (string, error) {
	// TODO: Call out to the plugin to find a description.
	// For now just use the path to the plugin
	return plugin.path, nil
	// return strings.Join(plugin.commandParts, "-"), nil
}

// The the command path leading to this plugin.
// Eg. for a plugin "kn source github" this will be [ "source", "github" ]
func (plugin *plugin) CommandParts() []string {
	return plugin.commandParts
}

// Return the path to the plugin
func (plugin *plugin) Path() string {
	return plugin.path
}

// Name of the plugin
func (plugin *plugin) Name() string {
	return plugin.name
}

// =========================================================================================

// Find out all directories that might hold a plugin
func (manager *Manager) pluginLookupDirectories() ([]string, error) {
	pluginPath, err := homedir.Expand(manager.pluginsDir)
	if err != nil {
		return nil, err
	}
	dirs := []string{pluginPath}
	if manager.lookupInPath {
		dirs = append(dirs, filepath.SplitList(os.Getenv("PATH"))...)
	}
	dirs = uniquePathsList(dirs)
	return dirs, nil
}

// HelpTemplateFuncs returns a function map which can be used in templates for resolving
// plugin related help messages
func (manager *Manager) HelpTemplateFuncs() *template.FuncMap {
	ret := template.FuncMap{
		"listPlugins": manager.listPluginsHelpMessage(),
	}

	return &ret
}

// listPluginsHelpMessage returns a function which returns all plugins that are directly below the given
// command as a properly formatted string
func (manager *Manager) listPluginsHelpMessage() func(cmd *cobra.Command) string {
	return func(cmd *cobra.Command) string {
		if !cmd.HasSubCommands() {
			return ""
		}
		list, err := manager.ListPluginsForCommandGroup(extractCommandGroup(cmd, []string{}))
		if err != nil || len(list) == 0 {
			// We don't show plugins if there is an error
			return ""
		}
		var plugins []string
		for _, pl := range list {
			t := fmt.Sprintf("  %%-%ds %%s", cmd.NamePadding())
			desc, _ := pl.Description()
			command := (pl.CommandParts())[len(pl.CommandParts())-1]
			help := fmt.Sprintf(t, command, desc)
			plugins = append(plugins, help)
		}
		return strings.Join(plugins, "\n")
	}
}

// extractCommandGroup constructs the command path as array of strings
func extractCommandGroup(cmd *cobra.Command, parts []string) []string {
	if cmd.HasParent() {
		parts = extractCommandGroup(cmd.Parent(), parts)
		parts = append(parts, cmd.Name())
	}
	return parts
}

// uniquePathsList deduplicates a given slice of strings without
// sorting or otherwise altering its order in any way.
func uniquePathsList(paths []string) []string {
	seen := map[string]bool{}
	newPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		newPaths = append(newPaths, p)
	}
	return newPaths
}

// Split up a command name, discard the initial prefix ("kn-") and convert
// parts to command syntax (i.e. replace _ with -)
func extractPluginCommandFromFileName(name string) []string {
	// Remove extension on windows
	name = stripWindowsExecExtensions(name)
	parts := strings.Split(name, "-")
	if len(parts) < 1 {
		return []string{}
	}
	ret := make([]string, 0, len(parts)-1)
	for _, p := range parts[1:] {
		ret = append(ret, convertUnderscoreToDash(p))
	}
	return ret
}

// Strip any extension that indicates an EXE on Windows
func stripWindowsExecExtensions(name string) string {
	if runtime.GOOS == "windows" {
		ext := filepath.Ext(name)
		if len(ext) > 0 {
			for _, e := range windowsExecExtensions {
				if ext == e {
					name = name[:len(name)-len(ext)]
					break
				}
			}
		}
	}
	return name
}

// Return the path and the parts building the most specific plugin in the given directory
// If lookupInPath is true, then also the OS PATH is checked.
// An error returned if any IO operation fails
func findMostSpecificPluginInPath(dir string, parts []string, lookupInPath bool) (Plugin, error) {
	for i := len(parts); i > 0; i-- {

		// Construct plugin name to lookup
		var nameParts []string
		var commandParts []string
		for _, p := range parts[0:i] {
			// Subcommands with "-" are translated to "_"
			// (e.g. a command  "kn log-all" is translated to a plugin "kn-log_all")
			nameParts = append(nameParts, convertDashToUnderscore(p))
			commandParts = append(commandParts, p)
		}
		name := fmt.Sprintf("kn-%s", strings.Join(nameParts, "-"))

		// Check for the name in plugin directory and PATH (if requested)
		path, err := findInDirOrPath(name, dir, lookupInPath)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot lookup plugin %s in directory %s (lookup in path: %t)", name, dir, lookupInPath))
		}

		// Found, return it
		if path != "" {
			return &plugin{
				path:         path,
				commandParts: commandParts,
				name:         name,
			}, nil
		}
	}

	// Nothing found
	return nil, nil
}

// convertDashToUnderscore converts from the command name to the file name
func convertDashToUnderscore(p string) string {
	return strings.Replace(p, "-", "_", -1)
}

// convertUnderscoreToDash converts from the filename to the command name
func convertUnderscoreToDash(p string) string {
	return strings.Replace(p, "_", "-", -1)
}

// Find a command with name in the given directory or on the execution PATH (if lookupInPath is true)
// On Windows, also check well known extensions for executables
// Return the full path found or "" if none has found
// Return an error on any IO error.
func findInDirOrPath(name string, dir string, lookupInPath bool) (string, error) {

	exts := []string{""}
	if runtime.GOOS == "windows" {
		// Add also well known extensions for windows
		exts = append(exts, windowsExecExtensions...)
	}

	for _, ext := range exts {
		nameExt := name + ext

		// Check plugin dir first
		path := filepath.Join(dir, nameExt)
		_, err := os.Stat(path)
		if err == nil {
			// Found in dir
			return path, nil
		}
		if !os.IsNotExist(err) {
			return "", errors.Wrap(err, fmt.Sprintf("i/o error while reading %s", path))
		}

		// Check in PATH if requested
		if lookupInPath {
			path, err = exec.LookPath(name)
			if err == nil {
				// Found in path
				return path, nil
			}
			if execErr, ok := err.(*exec.Error); !ok || execErr.Unwrap() != exec.ErrNotFound {
				return "", errors.Wrap(err, fmt.Sprintf("error for looking up %s in path", name))
			}
		}
	}

	// Not found
	return "", nil
}

// lookupInternalPlugin looks up internally registered plugins. Return nil if none is found.
// Start with longest argument path first to find the most specific match
func lookupInternalPlugin(parts []string) Plugin {
	for i := len(parts); i > 0; i-- {
		checkParts := parts[0:i]
		for _, plugin := range InternalPlugins {
			if equalsSlice(plugin.CommandParts(), checkParts) {
				return plugin
			}
		}
	}
	return nil
}

// equalsSlice return true if two string slices contain the same elements
func equalsSlice(a, b []string) bool {
	if len(a) != len(b) || len(a) == 0 {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
