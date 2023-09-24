// Copyright © 2023 The Knative Authors
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

//go:build !windows

package plugin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

type testPluginWithManifest struct {
	testPlugin
	manifest    *Manifest
	contextData map[string]string
}

func (t testPluginWithManifest) GetManifest() *Manifest            { return t.manifest }
func (t testPluginWithManifest) GetContextData() map[string]string { return t.contextData }
func (t testPluginWithManifest) ExecuteWithContext(ctx context.Context, args []string) error {
	return nil
}

// Verify both interfaces are implemented
var _ Plugin = testPluginWithManifest{}
var _ PluginWithManifest = testPluginWithManifest{}

func TestFetchManifest(t *testing.T) {
	testCases := []struct {
		name             string
		cmdPart          []string
		hasManifest      bool
		expectedManifest *Manifest
	}{
		{
			name:        "Inlined with manifest",
			cmdPart:     []string{"cmd"},
			hasManifest: true,
			expectedManifest: &Manifest{
				Path:                    "",
				HasManifest:             true,
				ProducesContextDataKeys: []string{"service"},
				ConsumesContextDataKeys: []string{"service"},
			},
		},
		{
			name:        "Inlined no manifest",
			cmdPart:     []string{"no", "manifest"},
			hasManifest: false,
			expectedManifest: &Manifest{
				HasManifest: false,
			},
		},
		{
			name:             "No plugins",
			cmdPart:          []string{},
			hasManifest:      false,
			expectedManifest: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := setup(t)
			t.Cleanup(func() {
				cleanup(t, c)
				InternalPlugins = PluginList{}
				CtxManager = nil
			})

			if len(tc.cmdPart) == 0 {
				ctxManager, err := NewContextManager()
				assert.NilError(t, err)

				err = ctxManager.FetchManifests(c.pluginManager)
				assert.NilError(t, err)
				assert.Assert(t, len(ctxManager.Manifests) == 0)
				return
			}

			if tc.hasManifest {
				prepareInternalPlugins(testPluginWithManifest{
					testPlugin:  testPlugin{parts: tc.cmdPart},
					manifest:    tc.expectedManifest,
					contextData: map[string]string{},
				})
			} else {
				prepareInternalPlugins(testPlugin{parts: tc.cmdPart})
			}

			ctxManager, err := NewContextManager()
			assert.NilError(t, err)

			err = ctxManager.FetchManifests(c.pluginManager)
			assert.NilError(t, err)
			assert.Assert(t, len(ctxManager.Manifests) == 1)

			expectedKey := "kn-" + strings.Join(tc.cmdPart, "-")
			assert.DeepEqual(t, ctxManager.Manifests[expectedKey], *tc.expectedManifest)
		})
	}

}

func TestContextFetchExternalManifests(t *testing.T) {
	testCases := []struct {
		name             string
		testScript       string
		expectedManifest *Manifest
	}{
		{
			name: "manifest",
			testScript: `#!/bin/bash
echo '{"hasManifest":true,"consumesKeys":["service"]}'\n`,
			expectedManifest: &Manifest{
				HasManifest:             true,
				ConsumesContextDataKeys: []string{"service"},
			},
		},
		{
			name: "badjson",
			testScript: `#!/bin/bash
echo '{hasManifest:true,"consumesKeys":["service"]}'\n`,
			expectedManifest: nil,
		},
		{
			name: "badscript",
			testScript: `#!/bin/bash
exit 1\n`,
			expectedManifest: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ct := setup(t)
			defer cleanup(t, ct)

			fullPath := createTestPluginInDirectoryFromScript(t, "kn-"+tc.name, ct.pluginsDir, tc.testScript)
			// Set expected path
			if tc.expectedManifest != nil {
				tc.expectedManifest.Path = fullPath
			}

			testPlugin, err := ct.pluginManager.FindPlugin([]string{tc.name})
			assert.NilError(t, err)
			assert.Assert(t, testPlugin != nil)

			actual := fetchExternalManifest(testPlugin)
			assert.DeepEqual(t, actual, tc.expectedManifest)
		})

	}

}

// CreateTestPluginInPath with name, path, script, and fileMode and return the tmp random path
func createTestPluginInDirectoryFromScript(t *testing.T, name string, dir string, script string) string {
	fullPath := filepath.Join(dir, name)
	err := os.WriteFile(fullPath, []byte(script), 0777)
	assert.NilError(t, err)
	// Some extra files to feed the tests
	err = os.WriteFile(filepath.Join(dir, "non-plugin-prefix-"+name), []byte{}, 0555)
	assert.NilError(t, err)
	_, err = os.CreateTemp(dir, "bogus-dir")
	assert.NilError(t, err)

	return fullPath
}
