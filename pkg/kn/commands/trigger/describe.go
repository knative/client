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

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// NewTriggerDescribeCommand returns a new command for describe a trigger
func NewTriggerDescribeCommand(p *commands.KnParams) *cobra.Command {

	triggerDescribe := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a trigger",
		Example: `
  # Describe a trigger with name 'my-trigger'
  kn trigger describe my-trigger`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn trigger describe' requires name of the trigger as single argument")
			}
			name := args[0]

			// get namespace
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			// get client
			eventingClient, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			trigger, err := eventingClient.GetTrigger(name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeTrigger(dw, trigger, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Revisions summary info
			writeSink(dw, &trigger.Spec.Subscriber)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, trigger.Status.Conditions, printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := triggerDescribe.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")

	return triggerDescribe
}

func writeSink(dw printers.PrefixWriter, sink *duckv1.Destination) {
	subWriter := dw.WriteAttribute("Sink", "")
	subWriter.WriteAttribute("Name", sink.Ref.Name)
	subWriter.WriteAttribute("Namespace", sink.Ref.Namespace)
	ref := sink.Ref
	if ref != nil {
		subWriter.WriteAttribute("Resource", fmt.Sprintf("%s (%s)", sink.Ref.Kind, sink.Ref.APIVersion))
	}
	uri := sink.URI
	if uri != nil {
		subWriter.WriteAttribute("URI", uri.String())
	}
}

func writeTrigger(dw printers.PrefixWriter, trigger *v1alpha1.Trigger, printDetails bool) {
	commands.WriteMetadata(dw, &trigger.ObjectMeta, printDetails)
	dw.WriteAttribute("Broker", trigger.Spec.Broker)
	subWriter := dw.WriteAttribute("Filter", "")
	for key, value := range *trigger.Spec.Filter.Attributes {
		subWriter.WriteAttribute(key, value)
	}
}
