// Copyright © 2019 The Knative Authors
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

package trigger

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	client_v1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewTriggerUpdateCommand prepares the command for a CronJobSource update
func NewTriggerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var triggerUpdateFlags TriggerUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME --broker BROKER --filter KEY=VALUE --sink SINK",
		Short: "Update a trigger",
		Example: `
  # Update the broker of a trigger 'mytrigger' to 'new-broker'
  kn trigger update mytrigger --broker new-broker

  # Update the filter of a trigger 'mytrigger' to 'type=knative.dev.bar'
  kn trigger update mytrigger --filter type=knative.dev.bar

  # Update the sink of a trigger 'mytrigger' to 'svc:new-service'
  kn trigger update mytrigger --sink svc:new-service
  `,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("name of trigger required")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			eventingClient, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			servingClient, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			trigger, err := eventingClient.GetTrigger(name)
			if err != nil {
				return err
			}

			b := client_v1alpha1.NewTriggerBuilderFromExisting(trigger)

			if cmd.Flags().Changed("broker") {
				b.Broker(triggerUpdateFlags.Broker)
			}
			if cmd.Flags().Changed("filter") {
				b.Filter(triggerUpdateFlags.GetFilters())
			}
			if cmd.Flags().Changed("sink") {
				destination, err := sinkFlags.ResolveSink(servingClient)
				if err != nil {
					return err
				}
				b.Sink(destination)
			}
			err = eventingClient.UpdateTrigger(b.Build())
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Trigger '%s' updated in namespace '%s'.\n", name, namespace)
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	triggerUpdateFlags.Add(cmd)
	sinkFlags.Add(cmd)

	return cmd
}
