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
	"knative.dev/client/pkg/util"
)

// NewBindingCreateCommand is for creating sink bindings
func NewBindingCreateCommand(p *commands.KnParams) *cobra.Command {
	var bindingFlags bindingUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --subject SUBJECT --sink SINK",
		Short: "Create a sink binding",
		Example: `
  # Create a sink binding which connects a deployment 'myapp' with a Knative service 'mysvc'
  kn source binding create my-binding --subject Deployment:apps/v1:myapp --sink ksvc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the sink binding to create as single argument")

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

			destination, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return err
			}

			reference, err := util.ToTrackerReference(bindingFlags.subject, namespace)
			if err != nil {
				return err
			}

			ceOverridesMap, err := util.MapFromArrayAllowingSingles(bindingFlags.ceOverrides, "=")
			if err != nil {
				return err
			}
			ceOverridesToRemove := util.ParseMinusSuffix(ceOverridesMap)

			bindingBuilder := v1alpha12.NewSinkBindingBuilder(name).
				Sink(destination).
				Subject(reference).
				Namespace(namespace).
				CloudEventOverrides(ceOverridesMap, ceOverridesToRemove)

			binding, err := bindingBuilder.Build()
			if err != nil {
				return err
			}
			err = sinkBindingClient.CreateSinkBinding(binding)
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Sink binding '%s' created in namespace '%s'.\n", args[0], sinkBindingClient.Namespace())
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	bindingFlags.addBindingFlags(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("subject")
	cmd.MarkFlagRequired("sink")

	return cmd
}
