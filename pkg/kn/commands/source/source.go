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

package source

import (
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/source/apiserver"
	"knative.dev/client/pkg/kn/commands/source/binding"
	"knative.dev/client/pkg/kn/commands/source/ping"
)

func NewSourceCommand(p *commands.KnParams) *cobra.Command {
	sourceCmd := &cobra.Command{
		Use:   "source",
		Short: "Event source command group",
	}
	sourceCmd.AddCommand(NewListTypesCommand(p))
	sourceCmd.AddCommand(NewListCommand(p))
	sourceCmd.AddCommand(apiserver.NewAPIServerCommand(p))
	sourceCmd.AddCommand(ping.NewPingCommand(p))
	sourceCmd.AddCommand(binding.NewBindingCommand(p))
	return sourceCmd
}
