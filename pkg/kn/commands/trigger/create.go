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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// NewTriggerCreateCommand to create trigger create command
func NewTriggerCreateCommand(p *commands.KnParams) *cobra.Command {
	var triggerUpdateFlags TriggerUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --broker BROKER --filter KEY=VALUE --sink SINK",
		Short: "Create a trigger",
		Example: `
  # Create a trigger 'mytrigger' to declare a subscription to events with attribute 'type=dev.knative.foo' from default broker. The subscriber is service 'mysvc'
  kn trigger create mytrigger --broker default --filter type=dev.knative.foo --sink svc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("'trigger create' requires the name of the trigger")
			}
			name := args[0]
			sourceArgs := args[1:]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			eventingClient, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			objectRef, err := sinkFlags.ResolveSink(dynamicClient, namespace)
			if err != nil {
				return fmt.Errorf(
					"cannot create trigger '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			filters, err := triggerUpdateFlags.GetFilters()
			if err != nil {
				return fmt.Errorf(
					"cannot create trigger '%s' "+
						"because %s", name, err)
			}
			if filters == nil {
				return fmt.Errorf(
					"cannot create trigger '%s' "+
						"because filters are required", name)
			}

			trigger := constructTrigger(name, namespace, triggerUpdateFlags.Broker, filters)
			trigger.Spec.Subscriber = &duckv1.Destination{
				Ref: objectRef.Ref,
				URI: objectRef.URI,
			}
			if triggerUpdateFlags.Source != "" {
				fullSourceArgs := []string{
					"source", triggerUpdateFlags.Source, "create",
					"--sink", fmt.Sprintf("broker:%s", triggerUpdateFlags.Broker)}
				fullSourceArgs = append(fullSourceArgs, sourceArgs...)
				fullSourceArgs = append(fullSourceArgs, name)
				root := cmd
				for root.HasParent() {
					root = root.Parent()
				}
				root.SetArgs(fullSourceArgs)
				err = root.Execute()
				if err != nil {
					return err
				}
			}

			err = eventingClient.CreateTrigger(trigger)
			if err != nil {
				return fmt.Errorf(
					"cannot create trigger '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Trigger '%s' successfully created in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	triggerUpdateFlags.Add(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("sink")
	cmd.MarkFlagRequired("filter")

	return cmd
}

// constructTrigger is to create an instance of v1alpha1.Trigger
func constructTrigger(name string, namespace string, broker string, filters map[string]string) *v1alpha1.Trigger {
	trigger := v1alpha1.Trigger{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.TriggerSpec{
			Broker: broker,
		},
	}

	triggerFilterAttributes := v1alpha1.TriggerFilterAttributes(filters)
	trigger.Spec.Filter = &v1alpha1.TriggerFilter{
		Attributes: &triggerFilterAttributes,
	}

	return &trigger
}
