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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Collection of errors and warning collected during verifications
type VerificationErrorsAndWarnings struct {
	Errors   []string
	Warnings []string
}

// permission bits for execute
const (
	UserExecute  = 1 << 6
	GroupExecute = 1 << 3
	OtherExecute = 1 << 0
)

// Verification of a ll plugins. This method returns all errors and warnings
// for the verification. The following criteria are verified (for each plugin):
// * If the plugin is executable
// * If the plugin is overshadowed by a previous plugin
func (manager *Manager) Verify() VerificationErrorsAndWarnings {
	eaw := VerificationErrorsAndWarnings{}

	dirs, err := manager.pluginLookupDirectories()
	if err != nil {
		return eaw.AddError("cannot lookup plugin directories: %v", err)
	}

	// Examine all files in possible plugin directories

	seenPlugins := make(map[string]string)
	for _, dir := range dirs {
		files, err := ioutil.ReadDir(dir)

		// Ignore non-existing directories
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			eaw.AddError("unable to read directory '%s' from your plugin path: %v", dir, err)
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !strings.HasPrefix(f.Name(), "kn-") {
				continue
			}
			eaw = verifyPath(filepath.Join(dir, f.Name()), seenPlugins, eaw)
		}
	}
	return eaw
}

func verifyPath(path string, seenPlugins map[string]string, eaw VerificationErrorsAndWarnings) VerificationErrorsAndWarnings {

	// Verify that plugin actually exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if err == os.ErrNotExist {
			return eaw.AddError("cannot find plugin in %s", path)
		}
		return eaw.AddError("cannot stat %s: %v", path, err)
	}

	eaw = addWarningIfNotExecutable(eaw, path, fileInfo)
	eaw = addWarningIfAlreadySeen(eaw, seenPlugins, path)

	// Remember each verified plugin for duplicate check
	seenPlugins[filepath.Base(path)] = path

	return eaw
}

func addWarningIfNotExecutable(eaw VerificationErrorsAndWarnings, path string, fileInfo os.FileInfo) VerificationErrorsAndWarnings {
	if runtime.GOOS == "windows" {
		return checkForWindowsExecutable(eaw, fileInfo, path)
	}

	mode := fileInfo.Mode()
	if !mode.IsRegular() && !isSymlink(mode) {
		return eaw.AddWarning("%s is not a file", path)
	}
	perms := uint32(mode.Perm())

	uid, gid, err := statFileOwner(fileInfo)
	if err != nil {
		return eaw.AddWarning("%s", err.Error())
	}
	isOwner := checkIfUserIsFileOwner(uid)
	isInGroup, err := checkIfUserInGroup(gid)
	if err != nil {
		return eaw.AddError("cannot get group ids for checking executable status of file %s", path)
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

	return eaw.AddWarning("%s is not executable by current user", path)
}

func addWarningIfAlreadySeen(eaw VerificationErrorsAndWarnings, seenPlugins map[string]string, path string) VerificationErrorsAndWarnings {
	fileName := filepath.Base(path)
	if existingPath, ok := seenPlugins[fileName]; ok {
		return eaw.AddWarning("%s is ignored because it is shadowed by an equally named plugin: %s", path, existingPath)
	}
	return eaw
}

func checkForWindowsExecutable(eaw VerificationErrorsAndWarnings, fileInfo os.FileInfo, path string) VerificationErrorsAndWarnings {
	name := fileInfo.Name()
	nameWithoutExecExtension := stripWindowsExecExtensions(name)

	if name == nameWithoutExecExtension {
		return eaw.AddWarning("%s is not executable as it does not have a Windows exec extension (one of %s)", path, strings.Join(windowsExecExtensions, ", "))
	}
	return eaw
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

func (eaw *VerificationErrorsAndWarnings) AddError(format string, args ...interface{}) VerificationErrorsAndWarnings {
	eaw.Errors = append(eaw.Errors, fmt.Sprintf(format, args...))
	return *eaw
}

func (eaw *VerificationErrorsAndWarnings) AddWarning(format string, args ...interface{}) VerificationErrorsAndWarnings {
	eaw.Warnings = append(eaw.Warnings, fmt.Sprintf(format, args...))
	return *eaw
}

func (eaw *VerificationErrorsAndWarnings) PrintWarningsAndErrors(out io.Writer) {
	printSection(out, "ERROR", eaw.Errors)
	printSection(out, "WARNING", eaw.Warnings)
}

func (eaw *VerificationErrorsAndWarnings) HasErrors() bool {
	return len(eaw.Errors) > 0
}

func (eaw *VerificationErrorsAndWarnings) IsEmpty() bool {
	return len(eaw.Errors)+len(eaw.Warnings) == 0
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

func isSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}
