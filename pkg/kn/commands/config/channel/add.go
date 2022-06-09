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

package channel

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"knative.dev/client/pkg/kn/config"
)

const ChannelTypeAliasExistsErrorMsg = "Channel type alias %q already exists in the configuration file"

// NewAddCommand defines and processes `kn config channel-type-mapping add`
func NewAddCommand() *cobra.Command {
	ctmf := ChannelTypeMappingFlags{}
	addChannelTypeMapCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new channel type mapping",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items := []config.ChannelTypeMapping{}
			if err := viper.UnmarshalKey(config.KeyChannelTypeMappings, &items); err != nil {
				return err
			}
			channelHash := map[string]bool{}
			for _, item := range items {
				channelHash[item.Alias] = true
			}
			for _, alias := range args {
				if _, ok := channelHash[alias]; ok {
					return fmt.Errorf(ChannelTypeAliasExistsErrorMsg, alias)
				}
				ctmf.Alias = alias
			}
			newList := append(items, ctmf.ChannelTypeMapping)
			viper.Set(config.KeyChannelTypeMappings, newList)
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Channel type mapping with alias '%s' was added\n", args[0])
			return nil
		},
	}
	ctmf.AddFlags(addChannelTypeMapCmd)
	return addChannelTypeMapCmd
}
