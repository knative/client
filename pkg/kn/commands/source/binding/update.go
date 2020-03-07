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

package binding

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	v1alpha12 "knative.dev/client/pkg/sources/v1alpha2"
)

// NewBindingUpdateCommand prepares the command for a sink binding update
func NewBindingUpdateCommand(p *commands.KnParams) *cobra.Command {
	var bindingFlags bindingUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME --subject SCHEDULE --sink SINK --ce-override OVERRIDE",
		Short: "Update a sink binding.",
		Example: `
  # Update the subject of a sink binding 'my-binding' to a new cronjob with label selector 'app=ping'  
  kn source binding update my-binding --subject cronjob:batch/v1beta1:app=ping"`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the sink binding to update as single argument")
			}
			name := args[0]

			sinkBindingClient, err := newSinkBindingClient(p, cmd)
			if err != nil {
				return err
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			source, err := sinkBindingClient.GetSinkBinding(name)
			if err != nil {
				return err
			}

			b := v1alpha12.NewSinkBindingBuilderFromExisting(source)
			if cmd.Flags().Changed("sink") {
				destination, err := sinkFlags.ResolveSink(dynamicClient, namespace)
				if err != nil {
					return err
				}
				b.Sink(destination)
			}
			if cmd.Flags().Changed("subject") {
				reference, err := toReference(bindingFlags.subject, namespace)
				if err != nil {
					return err
				}
				b.Subject(reference)
			}
			err = updateCeOverrides(bindingFlags, b)
			if err != nil {
				return err
			}

			binding, err := b.Build()
			if err != nil {
				return err
			}
			err = sinkBindingClient.UpdateSinkBinding(binding)
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Sink binding '%s' updated in namespace '%s'.\n", name, sinkBindingClient.Namespace())
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	bindingFlags.addBindingFlags(cmd)
	sinkFlags.Add(cmd)

	return cmd
}
