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

// NewConfigGetCommand represents 'kn config get' command
func NewConfigGetCommand(p *commands.KnParams) *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get configuration settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, cfgPath := range args {
				cfg := viper.Get(cfgPath)
				if cfg == nil {
					return fmt.Errorf(ConfigKeyNotFoundErrorMsg, cfgPath)
				}
				b, err := yaml.Marshal(cfg)
				if err != nil {
					return err
				}
				fmt.Fprint(cmd.OutOrStdout(), string(b))
			}
			return nil
		},
	}
	return getCmd
}
