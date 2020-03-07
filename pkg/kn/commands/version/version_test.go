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
	"testing"
	"text/template"

	"knative.dev/client/pkg/kn/commands"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

type versionOutput struct {
	Version     string
	BuildDate   string
	GitRevision string
}

var versionOutputTemplate = `Version:      {{.Version}}
Build Date:   {{.BuildDate}}
Git Revision: {{.GitRevision}}
Supported APIs:
* Serving
  - serving.knative.dev/v1 (knative-serving v0.13.0)
* Eventing
  - sources.eventing.knative.dev/v1alpha1 (knative-eventing v0.13.1)
  - sources.eventing.knative.dev/v1alpha2 (knative-eventing v0.13.1)
  - eventing.knative.dev/v1alpha1 (knative-eventing v0.13.1)
`

const (
	fakeVersion     = "fake-version"
	fakeBuildDate   = "fake-build-date"
	fakeGitRevision = "fake-git-revision"
)

func TestVersion(t *testing.T) {
	var (
		versionCmd            *cobra.Command
		knParams              *commands.KnParams
		expectedVersionOutput string
		output                *bytes.Buffer
	)

	setup := func() {
		Version = fakeVersion
		BuildDate = fakeBuildDate
		GitRevision = fakeGitRevision

		expectedVersionOutput = genVersionOuput(t, versionOutputTemplate,
			versionOutput{
				fakeVersion,
				fakeBuildDate,
				fakeGitRevision})

		knParams = &commands.KnParams{}
		versionCmd = NewVersionCommand(knParams)
		output = new(bytes.Buffer)
		versionCmd.SetOutput(output)
	}

	t.Run("creates a VersionCommand", func(t *testing.T) {
		setup()

		assert.Equal(t, versionCmd.Use, "version")
		assert.Equal(t, versionCmd.Short, "Prints the client version")
		assert.Assert(t, versionCmd.Run != nil)
	})

	t.Run("prints version, build date, git revision, supported APIs", func(t *testing.T) {
		setup()

		versionCmd.Run(versionCmd, []string{})
		assert.Equal(t, output.String(), expectedVersionOutput)
	})

}

func genVersionOuput(t *testing.T, templ string, vOutput versionOutput) string {
	tmpl, err := template.New("versionOutput").Parse(versionOutputTemplate)
	assert.NilError(t, err)

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, vOutput)
	assert.NilError(t, err)

	return buf.String()
}
