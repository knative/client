/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package broker

import (
	"github.com/spf13/cobra"

	"knative.dev/client/pkg/commands"
)

// NewBrokerCommand represents broker management commands
func NewBrokerCommand(p *commands.KnParams) *cobra.Command {
	brokerCmd := &cobra.Command{
		Use:     "broker",
		Short:   "Manage message brokers",
		Aliases: []string{"brokers"},
	}
	brokerCmd.AddCommand(NewBrokerCreateCommand(p))
	brokerCmd.AddCommand(NewBrokerDescribeCommand(p))
	brokerCmd.AddCommand(NewBrokerDeleteCommand(p))
	brokerCmd.AddCommand(NewBrokerListCommand(p))
	brokerCmd.AddCommand(NewBrokerUpdateCommand(p))
	return brokerCmd
}
