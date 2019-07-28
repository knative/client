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
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/knative/client/pkg/util"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestPluginVerifier(t *testing.T) {

	var (
		pluginPath string
		rootCmd    *cobra.Command
		verifier   *pluginVerifier
	)

	setup := func(t *testing.T) {
		knParams := &commands.KnParams{}
		rootCmd, _, _ = commands.CreateTestKnCommand(NewPluginCommand(knParams), knParams)
		verifier = newPluginVerifier(rootCmd)
	}

	cleanup := func(t *testing.T) {
		if pluginPath != "" {
			DeleteTestPlugin(t, pluginPath)
		}
	}

	t.Run("with nil root command", func(t *testing.T) {
		t.Run("returns error verifying path", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			verifier.root = nil
			eaw := errorsAndWarnings{}
			eaw = verifier.verify(eaw, pluginPath)
			assert.Assert(t, len(eaw.errors) == 1)
			assert.Assert(t, len(eaw.warnings) == 0)
			assert.Assert(t, util.ContainsAll(eaw.errors[0], "nil root"))
		})
	})

	t.Run("with root command", func(t *testing.T) {
		t.Run("whether plugin in path is executable (unix only)", func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("Skip test for windows as the permission check are for Unix only")
				return
			}

			setup(t)
			defer cleanup(t)

			pluginDir, err := ioutil.TempDir("", "plugin")
			assert.NilError(t, err)
			defer os.RemoveAll(pluginDir)
			pluginPath := filepath.Join(pluginDir, "kn-execution-test")
			err = ioutil.WriteFile(pluginPath, []byte("#!/bin/sh\ntrue"), 0644)
			assert.NilError(t, err, "can't create test plugin")

			for _, uid := range getExecTestUids() {
				for _, gid := range getExecTestGids() {
					for _, userPerm := range []int{0, UserExecute} {
						for _, groupPerm := range []int{0, GroupExecute} {
							for _, otherPerm := range []int{0, OtherExecute} {
								perm := os.FileMode(userPerm | groupPerm | otherPerm + 0444)
								assert.NilError(t, prepareFile(pluginPath, uid, gid, perm), "prepare plugin file, uid: %d, gid: %d, perm: %03o", uid, gid, perm)

								eaw := errorsAndWarnings{}
								eaw = newPluginVerifier(rootCmd).verify(eaw, pluginPath)

								if isExecutable(pluginPath) {
									assert.Assert(t, len(eaw.warnings) == 0, "executable: perm %03o | uid %d | gid %d | %v", perm, uid, gid, eaw.warnings)
									assert.Assert(t, len(eaw.errors) == 0)
								} else {
									assert.Assert(t, len(eaw.warnings) == 1, "not executable: perm %03o | uid %d | gid %d | %v", perm, uid, gid, eaw.warnings)
									assert.Assert(t, len(eaw.errors) == 0)
									assert.Assert(t, util.ContainsAll(eaw.warnings[0], pluginPath))
								}

							}
						}
					}
				}
			}
		})

		t.Run("when kn plugin in path is executable", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			pluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable)

			t.Run("when kn plugin in path shadows another", func(t *testing.T) {
				var shadowPluginPath = CreateTestPlugin(t, KnTestPluginName, KnTestPluginScript, FileModeExecutable)
				verifier.seenPlugins[KnTestPluginName] = pluginPath
				defer DeleteTestPlugin(t, shadowPluginPath)

				t.Run("fails with overshadowed error", func(t *testing.T) {
					eaw := errorsAndWarnings{}
					eaw = verifier.verify(eaw, shadowPluginPath)
					assert.Assert(t, len(eaw.errors) == 0)
					assert.Assert(t, len(eaw.warnings) == 1)
					assert.Assert(t, util.ContainsAll(eaw.warnings[0], "shadowed", "ignored"))
				})
			})
		})

		t.Run("when kn plugin in path overwrites existing command", func(t *testing.T) {
			setup(t)
			defer cleanup(t)
			var overwritingPluginPath = CreateTestPlugin(t, "kn-plugin", KnTestPluginScript, FileModeExecutable)
			defer DeleteTestPlugin(t, overwritingPluginPath)

			t.Run("fails with overwrites error", func(t *testing.T) {
				eaw := errorsAndWarnings{}
				eaw = verifier.verify(eaw, overwritingPluginPath)
				assert.Assert(t, len(eaw.errors) == 1)
				assert.Assert(t, len(eaw.warnings) == 0)
				assert.Assert(t, util.ContainsAll(eaw.errors[0], "overwrite", "kn-plugin"))
			})
		})
	})
}

func isExecutable(plugin string) bool {
	_, err := exec.Command(plugin).Output()
	return err == nil
}

func getExecTestUids() []int {
	currentUser := os.Getuid()
	// Only root can switch ownership of a file
	if currentUser == 0 {
		foreignUser, err := lookupForeignUser()
		if err == nil {
			return []int{currentUser, foreignUser}
		}
	}
	return []int{currentUser}
}

func getExecTestGids() []int {
	currentUser := os.Getuid()
	currentGroup := os.Getgid()
	// Only root can switch group of a file
	if currentUser == 0 {
		foreignGroup, err := lookupForeignGroup()
		if err == nil {
			return []int{currentGroup, foreignGroup}
		}
	}
	return []int{currentGroup}
}

func lookupForeignUser() (int, error) {
	for _, probe := range []string{"daemon", "nobody", "_unknown"} {
		u, err := user.Lookup(probe)
		if err != nil {
			continue
		}
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			continue
		}
		if uid != os.Getuid() {
			return uid, nil
		}
	}
	return 0, errors.New("could not find foreign user")
}

func lookupForeignGroup() (int, error) {
	gids, err := os.Getgroups()
	if err != nil {
		return 0, err
	}
OUTER:
	for _, probe := range []string{"daemon", "wheel", "nobody", "nogroup", "admin"} {
		group, err := user.LookupGroup(probe)
		if err != nil {
			continue
		}
		gid, err := strconv.Atoi(group.Gid)
		if err != nil {
			continue
		}

		for _, g := range gids {
			if gid == g {
				continue OUTER
			}
		}
		return gid, nil
	}
	return 0, errors.New("could not find a foreign group")
}

func prepareFile(file string, uid int, gid int, perm os.FileMode) error {
	err := os.Chown(file, uid, gid)
	if err != nil {
		return err
	}
	return os.Chmod(file, perm)
}
