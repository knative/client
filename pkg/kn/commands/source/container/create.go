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

	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/sources/v1alpha2"
)

// NewContainerCreateCommand for creating source
func NewContainerCreateCommand(p *commands.KnParams) *cobra.Command {
	var podFlags knflags.PodSpecFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --image IMAGE --sink SINK",
		Short: "Create a container source",
		Example: `
  # Create a ContainerSource 'src' to start a container with image 'docker.io/sample/image' and send messages to service 'mysvc'
  kn source container create src --image docker.io/sample/image --sink ksvc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the source to create as single argument")
			}
			name := args[0]

			srcClient, err := newContainerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			namespace := srcClient.Namespace()

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			objectRef, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return fmt.Errorf(
					"cannot create ContainerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			podSpec := &corev1.PodSpec{Containers: []corev1.Container{{}}}
			err = podFlags.ResolvePodSpec(podSpec, cmd.Flags())
			if err != nil {
				return fmt.Errorf(
					"cannot create ContainerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			b := v1alpha2.NewContainerSourceBuilder(name).Sink(*objectRef).PodSpec(*podSpec)
			err = srcClient.CreateContainerSource(b.Build())
			if err != nil {
				return fmt.Errorf(
					"cannot create ContainerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "ContainerSource '%s' created in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	podFlags.AddFlags(cmd.Flags())
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("image")
	cmd.MarkFlagRequired("sink")
	return cmd
}
