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

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string
var BuildDate string
var GitRevision string
var ServingVersion string

func NewVersionCommand(p *KnParams) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the client version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version:      %s\n", Version)
			fmt.Printf("Build Date:   %s\n", BuildDate)
			fmt.Printf("Git Revision: %s\n", GitRevision)
			fmt.Printf("Dependencies:\n")
			fmt.Printf("- serving:    %s\n", ServingVersion)
			return nil
		},
	}
	return versionCmd
}
