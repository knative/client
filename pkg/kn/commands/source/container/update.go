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

package container

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands/flags"
	knflags "knative.dev/client/pkg/kn/flags"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/sources/v1alpha2"
)

// NewContainerUpdateCommand for managing source update
func NewContainerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var podFlags knflags.PodSpecFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME --image IMAGE",
		Short: "Update a container source",
		Example: `
  # Update a ContainerSource 'src' with a different image uri 'docker.io/sample/newimage'
  kn source container update src --image docker.io/sample/newimage`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the source as single argument")
			}
			name := args[0]

			// get client
			srcClient, err := newContainerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			namespace := srcClient.Namespace()

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			source, err := srcClient.GetContainerSource(name)
			if err != nil {
				return err
			}
			if source.GetDeletionTimestamp() != nil {
				return fmt.Errorf("can't update container source %s because it has been marked for deletion", name)
			}

			b := v1alpha2.NewContainerSourceBuilderFromExisting(source)
			podSpec := b.Build().Spec.Template.Spec
			err = podFlags.ResolvePodSpec(&podSpec, cmd.Flags())
			if err != nil {
				return fmt.Errorf(
					"cannot update ContainerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}
			b.PodSpec(podSpec)

			if cmd.Flags().Changed("sink") {
				objectRef, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
				if err != nil {
					return fmt.Errorf(
						"cannot update ContainerSource '%s' in namespace '%s' "+
							"because: %s", name, namespace, err)
				}
				b.Sink(*objectRef)
			}

			err = srcClient.UpdateContainerSource(b.Build())
			if err != nil {
				return fmt.Errorf(
					"cannot update ContainerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Container source '%s' updated in namespace '%s'.\n", args[0], namespace)
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	podFlags.AddFlags(cmd.Flags())
	sinkFlags.Add(cmd)
	return cmd
}
