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

// NewDeleteCommand defines and processes `kn config channel-type-mapping delete`
func NewDeleteCommand() *cobra.Command {
	deleteChannelTypeMapCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a channel type mapping by its alias",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			channelTypeMappings := []config.ChannelTypeMapping{}
			if err := viper.UnmarshalKey(config.KeyChannelTypeMappings, &channelTypeMappings); err != nil {
				return err
			}

			aliasHash := map[string]bool{}
			for _, alias := range args {
				aliasHash[alias] = true
			}

			newList := []config.ChannelTypeMapping{}
			for _, channelTypeMap := range channelTypeMappings {
				if _, ok := aliasHash[channelTypeMap.Alias]; !ok {
					newList = append(newList, channelTypeMap)
				}
			}
			viper.Set(config.KeyChannelTypeMappings, newList)
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Channel type mapping with alias '%s' was deleted\n", args[0])
			return nil
		},
	}
	return deleteChannelTypeMapCmd
}
