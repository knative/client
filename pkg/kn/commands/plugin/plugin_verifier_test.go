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
	"fmt"
	"io/ioutil"
	"os"
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
			pluginPath := filepath.Join(pluginDir,"kn-execution-test")
			err = ioutil.WriteFile(pluginPath, []byte{}, 0644)
			assert.NilError(t, err)

			t.Run("fails with not executable error", func(t *testing.T) {
				for _, data := range getExecutionCheckTestParams() {
					eaw := errorsAndWarnings{}
					assert.NilError(t,prepareFile(pluginPath, data.userId, data.groupId, data.mode))
					eaw = newPluginVerifier(rootCmd).verify(eaw, pluginPath)

					if data.isExecutable {
						assert.Assert(t, len(eaw.warnings) == 0, "executable: %s | %v", data.string(), eaw.warnings)
						assert.Assert(t, len(eaw.errors) == 0)
					} else {
						assert.Assert(t, len(eaw.warnings) == 1,  "not executable: %s | %v", data.string(), eaw.warnings)
						assert.Assert(t, len(eaw.errors) == 0)
						assert.Assert(t, util.ContainsAll(eaw.warnings[0], pluginPath))
					}
				}
			})
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

type executionCheckTestParams struct {
	mode         os.FileMode
	isExecutable bool
	userId       int
	groupId      int
}

func (d executionCheckTestParams) string() string {
	return fmt.Sprintf("mode: %03o, isExecutable: %t, uid: %d, gid: %d", d.mode, d.isExecutable, d.userId, d.groupId)
}

func getExecutionCheckTestParams() []executionCheckTestParams {
	currentUser := os.Getuid()
	currentGroup := os.Getgid()

	ret := []executionCheckTestParams {
		{0000, false, currentUser, currentGroup},
		{0100, true, currentUser, currentGroup},
		{0010, false, currentUser, currentGroup},
		{0001, false, currentUser, currentGroup},
		{0110, true, currentUser, currentGroup},
		{0011, false, currentUser, currentGroup},
		{0101, true, currentUser, currentGroup},
		{0111, true, currentUser, currentGroup},
	}

	// The following parameters only work when running under root
	// because otherwise you can't change file permissions to other users
	// or groups you are not belonging to
	if currentUser != 0 {
		return ret
	}

	foreignGroup, err := lookupForeignGroup()
	if err == nil {
		for _, param := range []executionCheckTestParams{
			{0000, false, currentUser, foreignGroup},
			{0100, true, currentUser, foreignGroup},
			{0010, false, currentUser, foreignGroup},
			{0001, false, currentUser, foreignGroup},
			{0110, true, currentUser, foreignGroup},
			{0011, false, currentUser, foreignGroup},
			{0101, true, currentUser, foreignGroup},
			{0111, true, currentUser, foreignGroup},
		} {
			ret = append(ret, param)
		}
	}

	foreignUser, err := lookupForeignUser()
	if err != nil {
		return ret
	}

	for _, param := range []executionCheckTestParams{
		{0000, false, foreignUser, foreignGroup},
		{0100, false, foreignUser, foreignGroup},
		{0010, false, foreignUser, foreignGroup},
		{0001, true, foreignUser, foreignGroup},
		{0110, false, foreignUser, foreignGroup},
		{0011, true, foreignUser, foreignGroup},
		{0101, true, foreignUser, foreignGroup},
		{0111, true, foreignUser, foreignGroup},
	} {
		ret = append(ret, param)
	}

	return ret
}

func lookupForeignUser() (int, error) {
	for _, probe := range []string { "daemon", "nobody", "_unknown" } {
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
	gids, err :=  os.Getgroups()
	if err != nil {
		return 0, err
	}
	OUTER:
	for _, probe := range []string { "daemon", "wheel", "nobody", "nogroup", "admin" } {
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
	err := os.Chmod(file, perm)
	if err != nil {
		return err
	}
	return os.Chown(file, uid, gid)
}

