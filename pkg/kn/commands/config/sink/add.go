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

const SinkPrefixExistsErrorMsg = "Sink prefix %q already exists in the configuration file"

// NewAddCommand defines and processes `kn config sink-mapping add`
func NewAddCommand() *cobra.Command {
	smf := SinkMappingFlags{}
	addSinkMapCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new sink mapping",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			items := []config.SinkMapping{}
			if err := viper.UnmarshalKey(config.KeySinkMappings, &items); err != nil {
				return err
			}
			sinkHash := map[string]bool{}
			for _, item := range items {
				sinkHash[item.Prefix] = true
			}
			for _, prefix := range args {
				if _, ok := sinkHash[prefix]; ok {
					return fmt.Errorf(SinkPrefixExistsErrorMsg, prefix)
				}
				smf.Prefix = prefix
			}
			newList := append(items, smf.SinkMapping)
			viper.Set(config.KeySinkMappings, newList)
			if err := viper.WriteConfig(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Sink mapping with prefix '%s' was added\n", args[0])
			return nil
		},
	}
	smf.AddFlags(addSinkMapCmd)
	return addSinkMapCmd
}
