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

	"knative.dev/client/pkg/kn/commands"

	"github.com/spf13/cobra"
)

var Version string
var BuildDate string
var GitRevision string

// update this var as we add more deps
var apiVersions = map[string][]string{
	"serving": {
		"serving.knative.dev/v1 (knative-serving v0.12.1-0.20200206201132-525b15d87dc1)",
	},
	"eventing": {
		"sources.eventing.knative.dev/v1alpha1 (knative-eventing v0.12.1-0.20200206203632-b0a7d8a77cc7)",
		"eventing.knative.dev/v1alpha1 (knative-eventing v0.12.1-0.20200206203632-b0a7d8a77cc7)",
	},
}

// NewVersionCommand implements 'kn version' command
func NewVersionCommand(p *commands.KnParams) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the client version",
		Run: func(cmd *cobra.Command, args []string) {
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
		},
	}
	return versionCmd
}
