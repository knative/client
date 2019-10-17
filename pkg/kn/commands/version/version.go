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
	"fmt"
	"io"

	"knative.dev/client/pkg/kn/commands"

	"github.com/spf13/cobra"
)

var Version string
var BuildDate string
var GitRevision string
var ServingVersion string

type SupportedAPIs []string

// update this var as we increase the serving version in go.mod
var knServingDep = "v0.8.0"
var supportMatrix = map[string]*SupportedAPIs{
	knServingDep: {"serving.knative.dev/v1alpha1 (knative-serving v0.8.0)"},
}

func (s SupportedAPIs) print(out io.Writer) {
	for _, api := range s {
		fmt.Fprintf(out, "- %s\n", api)
	}
}

// NewVersionCommand implements 'kn version' command
func NewVersionCommand(p *commands.KnParams) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the client version",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Version:      %s\n", Version)
			fmt.Fprintf(out, "Build Date:   %s\n", BuildDate)
			fmt.Fprintf(out, "Git Revision: %s\n", GitRevision)
			fmt.Fprintf(out, "Supported APIs:\n")
			if apis, ok := supportMatrix[ServingVersion]; ok {
				apis.print(out)
			} else {
				// ensure the go build works when we update,
				// but version command tests fails to prevent shipping
				fmt.Fprintf(out, "- Serving: %s\n", ServingVersion)
			}
			return nil
		},
	}
	return versionCmd
}
