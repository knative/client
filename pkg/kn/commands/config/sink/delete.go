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

// NewDeleteCommand represents 'kn config sink-mapping delete' command
func NewDeleteCommand() *cobra.Command {
	deleteSinkMapCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a sink mapping by its prefix",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sinkMappings := []config.SinkMapping{}
			if err := viper.UnmarshalKey(config.KeySinkMappings, &sinkMappings); err != nil {
				return err
			}

			prefixHash := map[string]bool{}
			for _, prefix := range args {
				prefixHash[prefix] = true
			}

			newList := []config.SinkMapping{}
			for _, sinkMap := range sinkMappings {
				if _, ok := prefixHash[sinkMap.Prefix]; !ok {
					newList = append(newList, sinkMap)
				}
			}
			viper.Set(config.KeySinkMappings, newList)
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Sink mapping with prefix '%s' was deleted\n", args[0])
			return nil
		},
	}
	return deleteSinkMapCmd
}
