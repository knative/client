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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"knative.dev/client/pkg/kn/commands"
)

var ArgumentMissingValueErrorMsg = "Invalid argument %q: this is the correct structure key=value"
var NonScalarConfigPathErrorMsg = "The given key %q does not store a scalar value"

// NewConfigSetCommand represents 'kn config set' command
func NewConfigSetCommand(p *commands.KnParams) *cobra.Command {
	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration settings that are scalar",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				keyVal, err := splitToKeyValue(arg)
				if err != nil {
					return err
				}
				cfgPath := keyVal[0]
				newValue := keyVal[1]
				if !isValidScalarConfigKey(cfgPath) {
					return fmt.Errorf(NonScalarConfigPathErrorMsg, cfgPath)
				}
				viper.Set(cfgPath, newValue)
				if err := viper.WriteConfig(); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s=%s", cfgPath, newValue); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return setCmd
}

func splitToKeyValue(arg string) ([2]string, error) {
	keyVal := strings.Split(arg, "=")
	if len(keyVal) != 2 {
		return [2]string{}, fmt.Errorf(ArgumentMissingValueErrorMsg, arg)
	}
	return [2]string{keyVal[0], keyVal[1]}, nil
}

func isValidScalarConfigKey(configPath string) bool {
	return configPath == "plugins.directory"
}
