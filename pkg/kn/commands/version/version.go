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

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"knative.dev/client/pkg/kn/commands"
)

var Version string
var BuildDate string
var GitRevision string

// update this var as we add more deps
var apiVersions = map[string][]string{
	"serving": {
		"serving.knative.dev/v1 (knative-serving v0.13.0)",
	},
	"eventing": {
		"sources.eventing.knative.dev/v1alpha1 (knative-eventing v0.13.1)",
		"sources.eventing.knative.dev/v1alpha2 (knative-eventing v0.13.1)",
		"eventing.knative.dev/v1alpha1 (knative-eventing v0.13.1)",
	},
}

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
		Short: "Prints the client version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("output") {
				return printVersionMachineReadable(cmd)
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

func printVersionMachineReadable(cmd *cobra.Command) error {
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
		return fmt.Errorf("Invalid value for output flag, choose one among 'json' or 'yaml'.")
	}
	return nil
}
