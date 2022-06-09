// Copyright © 2022 The Knative Authors
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

package configCmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"knative.dev/client/pkg/kn/commands"
)

// NewConfigSetCommand represents 'kn config show' command
func NewConfigShowCommand(p *commands.KnParams) *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show configuration settings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgSettings := viper.AllSettings()
			b, err := yaml.Marshal(cfgSettings)
			if err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), string(b))
			return nil
		},
	}
	return showCmd
}
