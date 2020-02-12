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
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	client_v1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	"knative.dev/client/pkg/util"
)

// NewTriggerUpdateCommand prepares the command for a tigger update
func NewTriggerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var triggerUpdateFlags TriggerUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME --filter KEY=VALUE --sink SINK",
		Short: "Update a trigger",
		Example: `
  # Update the filter which key is 'type' to value 'knative.dev.bar' in a trigger 'mytrigger'
  kn trigger update mytrigger --filter type=knative.dev.bar

  # Remove the filter which key is 'type' from a trigger 'mytrigger' 
  kn trigger update mytrigger --filter type-

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
			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			var retries = 0
			for {
				trigger, err := eventingClient.GetTrigger(name)
				if err != nil {
					return err
				}

				b := client_v1alpha1.NewTriggerBuilderFromExisting(trigger)

				if cmd.Flags().Changed("broker") {
					return fmt.Errorf(
						"cannot update trigger '%s' because broker is immutable", name)
				}
				if cmd.Flags().Changed("filter") {
					updated, removed, err := triggerUpdateFlags.GetUpdateFilters()
					if err != nil {
						return fmt.Errorf(
							"cannot update trigger '%s' because %s", name, err)
					}
					existing := extractFilters(trigger)
					b.Filters(existing.Merge(updated).Remove(removed))
				}
				if cmd.Flags().Changed("sink") {
					destination, err := sinkFlags.ResolveSink(dynamicClient, namespace)
					if err != nil {
						return err
					}
					b.Subscriber(&duckv1.Destination{
						Ref: destination.Ref,
						URI: destination.URI,
					})
				}
				err = eventingClient.UpdateTrigger(b.Build())
				if err != nil {
					if apierrors.IsConflict(err) && retries < MaxUpdateRetries {
						retries++
						continue
					}
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Trigger '%s' updated in namespace '%s'.\n", name, namespace)
				return nil
			}
		},
	}

	commands.AddNamespaceFlags(cmd.Flags(), false)
	triggerUpdateFlags.Add(cmd)
	sinkFlags.Add(cmd)

	return cmd
}

func extractFilters(trigger *v1alpha1.Trigger) util.StringMap {
	attributes := make(util.StringMap)
	if trigger.Spec.Filter != nil && trigger.Spec.Filter.Attributes != nil {
		for k, v := range *trigger.Spec.Filter.Attributes {
			attributes[k] = v
		}
	}
	return attributes
}
