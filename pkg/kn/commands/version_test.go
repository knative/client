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

package commands

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

type versionOutput struct {
	Version        string
	BuildDate      string
	GitRevision    string
	ServingVersion string
}

var versionOutputTemplate = `Version:      {{.Version}}
Build Date:   {{.BuildDate}}
Git Revision: {{.GitRevision}}
Dependencies:
- serving:    {{.ServingVersion}}
`

const (
	fakeVersion        = "fake-version"
	fakeBuildDate      = "fake-build-date"
	fakeGitRevision    = "fake-git-revision"
	fakeServingVersion = "fake-serving-version"
)

func TestVersion(t *testing.T) {
	var (
		versionCmd            *cobra.Command
		knParams              *KnParams
		expectedVersionOutput string
	)

	setup := func() {
		Version = fakeVersion
		BuildDate = fakeBuildDate
		GitRevision = fakeGitRevision
		ServingVersion = fakeServingVersion

		expectedVersionOutput = genVersionOuput(t, versionOutputTemplate,
			versionOutput{
				Version:        fakeVersion,
				BuildDate:      fakeBuildDate,
				GitRevision:    fakeGitRevision,
				ServingVersion: fakeServingVersion})

		knParams = &KnParams{}
		versionCmd = NewVersionCommand(knParams)
	}

	t.Run("creates a VersionCommand", func(t *testing.T) {
		setup()
		CaptureStdout(t); defer ReleaseStdout(t)

		assert.Equal(t, versionCmd.Use, "version")
		assert.Equal(t, versionCmd.Short, "Prints the client version")
		assert.Assert(t, versionCmd.RunE != nil)
	})

	t.Run("prints version, build date, git revision, and serving version string", func(t *testing.T) {
		setup()
		CaptureStdout(t); defer ReleaseStdout(t)

		err := versionCmd.RunE(nil, []string{})
		assert.Assert(t, err == nil)
		assert.Equal(t, ReadStdout(t), expectedVersionOutput)
	})
}

// Private

func genVersionOuput(t *testing.T, templ string, vOutput versionOutput) string {
	tmpl, err := template.New("versionOutput").Parse(versionOutputTemplate)
	assert.Assert(t, err == nil)

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, vOutput)
	assert.Assert(t, err == nil)

	return buf.String()
}
