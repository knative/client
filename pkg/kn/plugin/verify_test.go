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
	"bytes"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"testing"

	"knative.dev/client/pkg/util"

	"gotest.tools/v3/assert"
)

func TestPluginIsExecutableUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skip test for windows as the permission check are for Unix only")
		return
	}

	ctx := setup(t)
	defer cleanup(t, ctx)

	pluginPath := createTestPlugin(t, "kn-test", ctx)
	for _, uid := range getExecTestUids() {
		for _, gid := range getExecTestGids() {
			for _, userPerm := range []int{0, UserExecute} {
				for _, groupPerm := range []int{0, GroupExecute} {
					for _, otherPerm := range []int{0, OtherExecute} {
						perm := os.FileMode(userPerm | groupPerm | otherPerm + 0444)
						assert.NilError(t, prepareFile(pluginPath, uid, gid, perm), "prepare plugin file, uid: %d, gid: %d, perm: %03o", uid, gid, perm)

						eaw := ctx.pluginManager.Verify()

						if isExecutable(pluginPath) {
							assert.Assert(t, len(eaw.Warnings) == 0, "executable: perm %03o | uid %d | gid %d | %v", perm, uid, gid, eaw.Warnings)
						} else {
							assert.Assert(t, len(eaw.Warnings) == 1, "not executable: perm %03o | uid %d | gid %d | %v", perm, uid, gid, eaw.Warnings)
							assert.Assert(t, util.ContainsAll(eaw.Warnings[0], pluginPath))
						}
						assert.Assert(t, len(eaw.Errors) == 0)
					}
				}
			}
		}
	}
}

func TestPluginIsExecutableWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skip test for non-windows OS as this test checks for Windows Extensions")
		return
	}

	ctx := setup(t)
	defer cleanup(t, ctx)

	pluginPath := createTestPlugin(t, "kn-test.bat", ctx)
	eaw := ctx.pluginManager.Verify()
	assert.Equal(t, len(eaw.Warnings), 0)
	assert.Equal(t, len(eaw.Errors), 0)
	os.Remove(pluginPath)

	pluginPath = createTestPlugin(t, "kn-test", ctx)
	eaw = ctx.pluginManager.Verify()
	assert.Equal(t, len(eaw.Warnings), 1)
	assert.Assert(t, util.ContainsAll(eaw.Warnings[0], pluginPath))
	assert.Equal(t, len(eaw.Errors), 0)
}

func TestWarnIfPluginShadowsOtherPlugin(t *testing.T) {
	ctx := setupWithPathLookup(t, true)
	defer cleanup(t, ctx)

	pl1path := createTestPlugin(t, "kn-test", ctx)
	pathDir, cleanupFunc := preparePathDirectory(t)
	defer cleanupFunc()
	pl2path := createTestPluginInDirectory(t, "kn-test", pathDir)

	eaw := ctx.pluginManager.Verify()
	assert.Assert(t, !eaw.IsEmpty())
	assert.Equal(t, len(eaw.Errors), 0)
	assert.Equal(t, len(eaw.Warnings), 1)
	assert.Assert(t, util.ContainsAll(eaw.Warnings[0], "shadowed", "ignored", pl1path, pl2path))

	var buf bytes.Buffer
	eaw.PrintWarningsAndErrors(&buf)
	assert.Assert(t, util.ContainsAll(buf.String(), "WARNING", "shadowed", "ignored"))
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
