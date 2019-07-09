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

package plugin

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// PluginFlags contains all PLugin commands flags
type PluginFlags struct {
	NameOnly bool

	Verifier    PathVerifier
	PluginPaths []string

	genericclioptions.IOStreams
}

// AddPluginFlags adds the various flags to plugin command
func (p *PluginFlags) AddPluginFlags(command *cobra.Command) {
	command.Flags().BoolVar(&p.NameOnly, "name-only", false, "If true, display only the binary name of each plugin, rather than its full path")
}
