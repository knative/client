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

package version

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
	"knative.dev/client/pkg/commands"
	"sigs.k8s.io/yaml"
)

var Version string
var BuildDate string
var GitRevision string
var VersionServing string
var VersionEventing string

type knVersion struct {
	Version       string
	BuildDate     string
	GitRevision   string
	SupportedAPIs map[string][]string
}

// NewVersionCommand implements 'kn version' command
func NewVersionCommand(p *commands.KnParams) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of this client",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, ok := debug.ReadBuildInfo()
			var apiVersions map[string][]string
			if ok {
				// Retrieve Serving & Eventing go.mod versions, format: v0.y.z
				for _, dep := range info.Deps {
					if strings.Contains(dep.Path, "knative.dev/serving") {
						VersionServing = dep.Version
					}
					if strings.Contains(dep.Path, "knative.dev/eventing") {
						VersionEventing = dep.Version
					}
				}
				// For valid {Version} of kn take MajorMinor + patch version of dependency,
				// e.g for client's release version=v1.3.0 + serving, eventing modules v0.30.0
				// the resulting displayed version should be also 'v1.3.0'.
				// Those versions should represent production releases tagged and published.
				// Otherwise keep plain unmodified version from go.mod,
				// that should cover local or nightly builds etc.
				if semver.IsValid(Version) && semver.IsValid(VersionServing) && semver.IsValid(VersionEventing) {
					VersionServing = semver.MajorMinor(Version) + VersionServing[5:]
					VersionEventing = semver.MajorMinor(Version) + VersionEventing[5:]
				}
			}
			apiVersions = map[string][]string{
				"serving": {
					fmt.Sprintf("serving.knative.dev/v1 (knative-serving %s)", VersionServing),
				},
				"eventing": {
					fmt.Sprintf("sources.knative.dev/v1 (knative-eventing %s)", VersionEventing),
					fmt.Sprintf("eventing.knative.dev/v1 (knative-eventing %s)", VersionEventing),
				},
			}

			if cmd.Flags().Changed("output") {
				return printVersionMachineReadable(cmd, apiVersions)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Version:      %s\n", Version)
			fmt.Fprintf(out, "Build Date:   %s\n", BuildDate)
			fmt.Fprintf(out, "Git Revision: %s\n", GitRevision)
			fmt.Fprintf(out, "Supported APIs:\n")
			fmt.Fprintf(out, "* Serving\n")
			for _, api := range apiVersions["serving"] {
				fmt.Fprintf(out, "  - %s\n", api)
			}
			fmt.Fprintf(out, "* Eventing\n")
			for _, api := range apiVersions["eventing"] {
				fmt.Fprintf(out, "  - %s\n", api)
			}
			return nil
		},
	}
	versionCmd.Flags().StringP(
		"output",
		"o",
		"",
		"Output format. One of: json|yaml.",
	)
	return versionCmd
}

func printVersionMachineReadable(cmd *cobra.Command, apiVersions map[string][]string) error {
	out := cmd.OutOrStdout()
	v := knVersion{Version, BuildDate, GitRevision, apiVersions}
	format := cmd.Flag("output").Value.String()
	switch format {
	case "JSON", "json":
		b, err := json.MarshalIndent(v, "", "\t")
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(b))
	case "YAML", "yaml":
		b, err := yaml.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(b))
	default:
		return fmt.Errorf("invalid value for output flag, choose one among 'json' or 'yaml'")
	}
	return nil
}
