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

	"knative.dev/client/pkg/config"
	clientv1beta1 "knative.dev/client/pkg/eventing/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/spf13/cobra"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1"

	"knative.dev/client/pkg/commands"
	"knative.dev/client/pkg/commands/flags"
	"knative.dev/client/pkg/util"
)

// NewTriggerUpdateCommand prepares the command for a tigger update
func NewTriggerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var triggerUpdateFlags TriggerUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a trigger",
		Example: `
  # Update the filter which key is 'type' to value 'knative.dev.bar' in a trigger 'mytrigger'
  kn trigger update mytrigger --filter type=knative.dev.bar

  # Remove the filter which key is 'type' from a trigger 'mytrigger'
  kn trigger update mytrigger --filter type-

  # Update the sink of a trigger 'mytrigger' to 'ksvc:new-service'
  kn trigger update mytrigger --sink ksvc:new-service
  `,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
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
			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			updateFunc := func(trigger *v1beta1.Trigger) (*v1beta1.Trigger, error) {
				b := clientv1beta1.NewTriggerBuilderFromExisting(trigger)

				if cmd.Flags().Changed("broker") {
					return nil, fmt.Errorf(
						"cannot update trigger '%s' because broker is immutable", name)
				}
				if cmd.Flags().Changed("filter") {
					updated, removed, err := triggerUpdateFlags.GetUpdateFilters()
					if err != nil {
						return nil, fmt.Errorf(
							"cannot update trigger '%s' because %w", name, err)
					}
					existing := extractFilters(trigger)
					b.Filters(existing.Merge(updated).Remove(removed))
				}
				if cmd.Flags().Changed("sink") {
					destination, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
					if err != nil {
						return nil, err
					}
					b.Subscriber(&duckv1.Destination{
						Ref: destination.Ref,
						URI: destination.URI,
					})
				}
				return b.Build(), nil
			}
			err = eventingClient.UpdateTriggerWithRetry(cmd.Context(), name, updateFunc, config.DefaultRetry.Steps)
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

func extractFilters(trigger *v1beta1.Trigger) util.StringMap {
	attributes := make(util.StringMap)
	if trigger.Spec.Filter != nil && trigger.Spec.Filter.Attributes != nil {
		for k, v := range trigger.Spec.Filter.Attributes {
			attributes[k] = v
		}
	}
	return attributes
}
