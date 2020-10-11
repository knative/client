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

package version

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
	"sigs.k8s.io/yaml"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

var versionOutputTemplate = `Version:      {{.Version}}
Build Date:   {{.BuildDate}}
Git Revision: {{.GitRevision}}
Supported APIs:
* Serving{{range $apis := .SupportedAPIs.serving }}
  - {{$apis}}{{end}}
* Eventing{{range $apis := .SupportedAPIs.eventing }}
  - {{$apis}}{{end}}
`

const (
	fakeVersion     = "fake-version"
	fakeBuildDate   = "fake-build-date"
	fakeGitRevision = "fake-git-revision"
)

func TestVersion(t *testing.T) {
	var (
		versionCmd     *cobra.Command
		knParams       *commands.KnParams
		expectedOutput string
		knVersionObj   knVersion
		output         *bytes.Buffer
	)

	setup := func() {
		Version = fakeVersion
		BuildDate = fakeBuildDate
		GitRevision = fakeGitRevision
		knVersionObj = knVersion{fakeVersion, fakeBuildDate, fakeGitRevision, apiVersions}
		expectedOutput = genVersionOutput(t, knVersionObj)
		knParams = &commands.KnParams{}
		versionCmd = NewVersionCommand(knParams)
		output = new(bytes.Buffer)
		versionCmd.SetOutput(output)
	}

	runVersionCmd := func(args []string) error {
		setup()
		versionCmd.SetArgs(args)
		return versionCmd.Execute()
	}

	t.Run("creates a VersionCommand", func(t *testing.T) {
		setup()
		assert.Equal(t, versionCmd.Use, "version")
		assert.Assert(t, util.ContainsAll(versionCmd.Short, "version"))
		assert.Assert(t, versionCmd.RunE != nil)
	})

	t.Run("prints version, build date, git revision, supported APIs", func(t *testing.T) {
		err := runVersionCmd([]string{})
		assert.NilError(t, err)
		assert.Equal(t, output.String(), expectedOutput)
	})

	t.Run("print version command with machine readable output", func(t *testing.T) {
		t.Run("json", func(t *testing.T) {
			err := runVersionCmd([]string{"-oJSON"})
			assert.NilError(t, err)
			in := knVersion{}
			err = json.Unmarshal(output.Bytes(), &in)
			assert.NilError(t, err)
			assert.DeepEqual(t, in, knVersionObj)
		})

		t.Run("yaml", func(t *testing.T) {
			err := runVersionCmd([]string{"-oyaml"})
			assert.NilError(t, err)
			jsonData, err := yaml.YAMLToJSON(output.Bytes())
			assert.NilError(t, err)
			in := knVersion{}
			err = json.Unmarshal(jsonData, &in)
			assert.NilError(t, err)
			assert.DeepEqual(t, in, knVersionObj)
		})

		t.Run("invalid format", func(t *testing.T) {
			err := runVersionCmd([]string{"-o", "jsonpath"})
			assert.Assert(t, err != nil)
			assert.ErrorContains(t, err, "Invalid", "output", "flag", "choose", "among")
		})
	})

}

func genVersionOutput(t *testing.T, obj knVersion) string {
	tmpl, err := template.New("versionOutput").Parse(versionOutputTemplate)
	assert.NilError(t, err)
	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, obj)
	assert.NilError(t, err)
	return buf.String()
}
