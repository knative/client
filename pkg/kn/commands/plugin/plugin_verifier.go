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

package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// pluginVerifier verifies that existing kn commands are not overriden
type pluginVerifier struct {
	root        *cobra.Command
	seenPlugins map[string]string
}

// collect errors and warnings on the way
type errorsAndWarnings struct {
	errors   []string
	warnings []string
}

// Create new verifier
func newPluginVerifier(root *cobra.Command) *pluginVerifier {
	return &pluginVerifier{
		root:        root,
		seenPlugins: make(map[string]string),
	}
}

// permission bits for execute
const (
	UserExecute  = 1 << 6
	GroupExecute = 1 << 3
	OtherExecute = 1 << 0
)

// Verify implements pathVerifier and determines if a given path
// is valid depending on whether or not it overwrites an existing
// kn command path, or a previously seen plugin.
// This method is not idempotent and must be called for each path only once.
func (v *pluginVerifier) verify(eaw errorsAndWarnings, path string) errorsAndWarnings {
	if v.root == nil {
		return eaw.addError("unable to verify path with nil root")
	}

	// Verify that plugin actually exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if err == os.ErrNotExist {
			return eaw.addError("cannot find plugin in %s", path)
		}
		return eaw.addError("cannot stat %s: %v", path, err)
	}

	eaw = v.addErrorIfWrongPrefix(eaw, path)
	eaw = v.addWarningIfNotExecutable(eaw, path, fileInfo)
	eaw = v.addWarningIfAlreadySeen(eaw, path)
	eaw = v.addErrorIfOverwritingExistingCommand(eaw, path)

	// Remember each verified plugin for duplicate check
	v.seenPlugins[filepath.Base(path)] = path

	return eaw
}

func (v *pluginVerifier) addWarningIfAlreadySeen(eaw errorsAndWarnings, path string) errorsAndWarnings {
	fileName := filepath.Base(path)
	if existingPath, ok := v.seenPlugins[fileName]; ok {
		return eaw.addWarning("%s is ignored because it is shadowed by a equally named plugin: %s.", path, existingPath)
	}
	return eaw
}

func (v *pluginVerifier) addErrorIfOverwritingExistingCommand(eaw errorsAndWarnings, path string) errorsAndWarnings {
	fileName := filepath.Base(path)
	cmds := strings.Split(fileName, "-")
	if len(cmds) < 2 {
		return eaw.addError("%s is not a valid plugin filename as its missing a prefix", fileName)
	}
	cmds = cmds[1:]

	// Check both, commands with underscore and with dash because plugins can be called with both
	overwrittenCommands := make(map[string]bool)
	for _, c := range [][]string{cmds, convertUnderscoresToDashes(cmds)} {
		cmd, _, err := v.root.Find(c)
		if err == nil {
			overwrittenCommands[cmd.CommandPath()] = true
		}
	}
	for command := range overwrittenCommands {
		eaw.addError("%s overwrites existing built-in command: %s", fileName, command)
	}
	return eaw
}

func (v *pluginVerifier) addErrorIfWrongPrefix(eaw errorsAndWarnings, path string) errorsAndWarnings {
	fileName := filepath.Base(path)
	// Only pick the first prefix as it is very like that it will be reduced to
	// a single prefix anyway (PR pending)
	prefix := ValidPluginFilenamePrefixes[0]
	if !strings.HasPrefix(fileName, prefix+"-") {
		eaw.addWarning("%s plugin doesn't start with plugin prefix %s", fileName, prefix)
	}
	return eaw
}

func (v *pluginVerifier) addWarningIfNotExecutable(eaw errorsAndWarnings, path string, fileInfo os.FileInfo) errorsAndWarnings {
	if runtime.GOOS == "windows" {
		return checkForWindowsExecutable(eaw, fileInfo, path)
	}

	mode := fileInfo.Mode()
	if !mode.IsRegular() && !isSymlink(mode) {
		return eaw.addWarning("%s is not a file", path)
	}
	perms := uint32(mode.Perm())

	var sys *syscall.Stat_t
	var ok bool
	if sys, ok = fileInfo.Sys().(*syscall.Stat_t); !ok {
		// We can check the files' owner/group
		return eaw.addWarning("cannot check owner/group of file %s", path)
	}

	isOwner := checkIfUserIsFileOwner(sys.Uid)
	isInGroup, err := checkIfUserInGroup(sys.Gid)
	if err != nil {
		return eaw.addError("cannot get group ids for checking executable status of file %s", path)
	}

	// User is owner and owner can execute
	if canOwnerExecute(perms, isOwner) {
		return eaw
	}

	// User is in group which can execute, but user is not file owner
	if canGroupExecute(perms, isOwner, isInGroup) {
		return eaw
	}

	// All can execute, and the user is not file owner and not in the file's perm group
	if canOtherExecute(perms, isOwner, isInGroup) {
		return eaw
	}

	return eaw.addWarning("%s is not executable by current user", path)
}

func checkForWindowsExecutable(eaw errorsAndWarnings, fileInfo os.FileInfo, path string) errorsAndWarnings {
	fileExt := strings.ToLower(filepath.Ext(fileInfo.Name()))

	switch fileExt {
	case ".bat", ".cmd", ".com", ".exe", ".ps1":
		return eaw
	}
	return eaw.addWarning("%s is not executable as it does not have the proper extension", path)
}

func checkIfUserInGroup(gid uint32) (bool, error) {
	groups, err := os.Getgroups()
	if err != nil {
		return false, err
	}
	for _, g := range groups {
		if int(gid) == g {
			return true, nil
		}
	}
	return false, nil
}

func checkIfUserIsFileOwner(uid uint32) bool {
	if int(uid) == os.Getuid() {
		return true
	}
	return false
}

// Check if all can execute, and the user is not file owner and not in the file's perm group
func canOtherExecute(perms uint32, isOwner bool, isInGroup bool) bool {
	if perms&OtherExecute != 0 {
		if os.Getuid() == 0 {
			return true
		}
		if !isOwner && !isInGroup {
			return true
		}
	}
	return false
}

// Check if user is owner and owner can execute
func canOwnerExecute(perms uint32, isOwner bool) bool {
	if perms&UserExecute != 0 {
		if os.Getuid() == 0 {
			return true
		}
		if isOwner {
			return true
		}
	}
	return false
}

// Check if user is in group which can execute, but user is not file owner
func canGroupExecute(perms uint32, isOwner bool, isInGroup bool) bool {
	if perms&GroupExecute != 0 {
		if os.Getuid() == 0 {
			return true
		}
		if !isOwner && isInGroup {
			return true
		}
	}
	return false
}

func (eaw *errorsAndWarnings) addError(format string, args ...interface{}) errorsAndWarnings {
	eaw.errors = append(eaw.errors, fmt.Sprintf(format, args...))
	return *eaw
}

func (eaw *errorsAndWarnings) addWarning(format string, args ...interface{}) errorsAndWarnings {
	eaw.warnings = append(eaw.warnings, fmt.Sprintf(format, args...))
	return *eaw
}

func (eaw *errorsAndWarnings) printWarningsAndErrors(out io.Writer) {
	printSection(out, "ERROR", eaw.errors)
	printSection(out, "WARNING", eaw.warnings)
}

func (eaw *errorsAndWarnings) combinedError() error {
	if len(eaw.errors) == 0 {
		return nil
	}
	return errors.New(strings.Join(eaw.errors, ","))
}

func printSection(out io.Writer, label string, values []string) {
	if len(values) > 0 {
		printLabelWithConditionalPluralS(out, label, len(values))
		for _, value := range values {
			fmt.Fprintf(out, "  - %s\n", value)
		}
	}
}

func printLabelWithConditionalPluralS(out io.Writer, label string, nr int) {
	if nr == 1 {
		fmt.Fprintf(out, "%s:\n", label)
	} else {
		fmt.Fprintf(out, "%ss:\n", label)
	}
}

func convertUnderscoresToDashes(cmds []string) []string {
	ret := make([]string, len(cmds))
	for i := range cmds {
		ret[i] = strings.ReplaceAll(cmds[i], "_", "-")
	}
	return ret
}

func isSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}
