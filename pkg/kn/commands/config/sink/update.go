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

package sink

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"knative.dev/client/pkg/kn/config"
)

const SinkPrefixNotFoundErrorMsg = "Sink prefix %q was not found in the configuration file"

// NewUpdateCommand defines and processes `kn config sink-mapping update`
func NewUpdateCommand() *cobra.Command {
	smf := SinkMappingFlags{}
	updateSinkMapCmd := &cobra.Command{
		Use:   "update",
		Short: "Update sink mapping by its prefix",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items := []config.SinkMapping{}
			if err := viper.UnmarshalKey(config.KeySinkMappings, &items); err != nil {
				return err
			}
			prefixHash := map[string]int{}
			for _, prefix := range args {
				prefixHash[prefix] = -1
			}

			for ix, item := range items {
				if _, ok := prefixHash[item.Prefix]; ok {
					prefixHash[item.Prefix] = ix
				}
			}
			for prefix, index := range prefixHash {
				if index == -1 {
					return fmt.Errorf(SinkPrefixNotFoundErrorMsg, prefix)
				}
				if cmd.Flags().Changed("resource") {
					items[index].Resource = smf.Resource
				}
				if cmd.Flags().Changed("group") {
					items[index].Group = smf.Group
				}
				if cmd.Flags().Changed("version") {
					items[index].Version = smf.Version
				}
			}
			viper.Set(config.KeySinkMappings, items)
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Sink mapping with prefix '%s' was updated\n", args[0])
			return nil
		},
	}
	smf.AddFlags(updateSinkMapCmd)
	return updateSinkMapCmd
}
