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

const ChannelTypeAliasNotFoundErrorMsg = "Channel type alias %q was not found in the configuration file"

// NewUpdateCommand defines and processes `kn config channel-type-mapping update`
func NewUpdateCommand() *cobra.Command {
	ctmf := ChannelTypeMappingFlags{}
	updateChannelTypeMapCmd := &cobra.Command{
		Use:   "update",
		Short: "Update channel type mapping by its alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items := []config.ChannelTypeMapping{}
			if err := viper.UnmarshalKey(config.KeyChannelTypeMappings, &items); err != nil {
				return err
			}
			aliasHash := map[string]int{}
			for _, alias := range args {
				aliasHash[alias] = -1
			}

			for ix, item := range items {
				if _, ok := aliasHash[item.Alias]; ok {
					aliasHash[item.Alias] = ix
				}
			}
			for prefix, index := range aliasHash {
				if index == -1 {
					return fmt.Errorf(ChannelTypeAliasNotFoundErrorMsg, prefix)
				}
				if cmd.Flags().Changed("kind") {
					items[index].Kind = ctmf.Kind
				}
				if cmd.Flags().Changed("group") {
					items[index].Group = ctmf.Group
				}
				if cmd.Flags().Changed("version") {
					items[index].Version = ctmf.Version
				}
			}
			viper.Set(config.KeyChannelTypeMappings, items)
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Channel type mapping with alias '%s' was updated\n", args[0])
			return nil
		},
	}
	ctmf.AddFlags(updateChannelTypeMapCmd)
	return updateChannelTypeMapCmd
}
