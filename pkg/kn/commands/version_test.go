package commands_test

import (
	"bytes"
	"text/template"

	. "github.com/knative/client/pkg/kn/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

type VersionOutput struct {
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

var _ = Describe("Version", func() {
	const (
		fakeVersion        = "fake-version"
		fakeBuildDate      = "fake-build-date"
		fakeGitRevision    = "fake-git-revision"
		fakeServingVersion = "fake-serving-version"
	)

	var (
		versionCmd            *cobra.Command
		knParams              *KnParams
		expectedVersionOutput string
	)

	BeforeEach(func() {
		Version = fakeVersion
		BuildDate = fakeBuildDate
		GitRevision = fakeGitRevision
		ServingVersion = fakeServingVersion

		expectedVersionOutput = genVersionOuput(versionOutputTemplate,
			VersionOutput{
				Version:        fakeVersion,
				BuildDate:      fakeBuildDate,
				GitRevision:    fakeGitRevision,
				ServingVersion: fakeServingVersion})

		knParams = &KnParams{}
		versionCmd = NewVersionCommand(knParams)
	})

	Describe("NewVersionCommand", func() {
		It("Creates a VersionCommand", func() {
			Expect(versionCmd).NotTo(BeNil())
			Expect(versionCmd.Use).To(Equal("version"))
			Expect(versionCmd.Short).To(Equal("Prints the client version"))
			Expect(versionCmd.RunE).NotTo(BeNil())
		})

		It("Prints Version, Build date, GitRevision, and ServingVersion", func() {
			err := versionCmd.RunE(nil, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(ReadStdout()).To(Equal(expectedVersionOutput))
		})
	})
})

// Private

func genVersionOuput(templ string, vOutput VersionOutput) string {
	tmpl, err := template.New("versionOutput").Parse(versionOutputTemplate)
	Expect(err).NotTo(HaveOccurred())

	buf := bytes.Buffer{}
	err = tmpl.Execute(&buf, vOutput)
	Expect(err).NotTo(HaveOccurred())
	
	return buf.String()
}
