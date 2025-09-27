// Copyright © 2021 The Knative Authors
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

package quickstart

import (
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/commands"
)

// Version contains the version string of the quickstart command
var Version string

// BuildDate contains the build date of the quickstart command
var BuildDate string

// GitRevision contains the git revision of the quickstart command
var GitRevision string

// NewVersionCommand implements 'kn quickstart version' command
func NewVersionCommand(p *commands.KnParams) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the quickstart version",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Version:      %s\n", Version)
			fmt.Fprintf(out, "Build Date:   %s\n", BuildDate)
			fmt.Fprintf(out, "Git Revision: %s\n", GitRevision)
			return nil
		},
	}
}