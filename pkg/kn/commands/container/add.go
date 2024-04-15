// Copyright © 2021 The Knative Authors
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

package container

import (
	"errors"
	cliflag "k8s.io/component-base/cli/flag"
	"os"

	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/pkg/kn/commands"
	knflags "knative.dev/client/pkg/kn/flags"
	sigyaml "sigs.k8s.io/yaml"
)

// NewContainerAddCommand to create event channels
func NewContainerAddCommand(p *commands.KnParams) *cobra.Command {
	var podSpecFlags knflags.PodSpecFlags
	cmd := &cobra.Command{
		Use:   "add NAME",
		Short: "Add a container",
		Example: `
  The command is Beta and may change in the future releases.

  The 'container add' represents utility command that prints YAML container spec to standard output. It's useful for
  multi-container use cases to create definition with help of standard 'kn' option flags. It accepts all container related
  flag available for 'service create'. The command can be chained through Unix pipes to create multiple containers at once.

  # Add a container 'sidecar' from image 'docker.io/example/sidecar' and print it to standard output (Beta)
  kn container add sidecar --image docker.io/example/sidecar

  # Add command can be chained by standard Unix pipe symbol '|' and passed to 'service add|update|apply' commands (Beta)
  kn container add sidecar --image docker.io/example/sidecar:first | \
  kn container add second --image docker.io/example/sidecar:second | \
  kn service create myksvc --image docker.io/example/my-app:latest --containers -`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'container add' requires the container name given as single argument")
			}
			name := args[0]
			if podSpecFlags.Image == "" {
				return errors.New("'container add'  requires the image name to run provided with the --image option")
			}

			// Detect pipe input from previous container command
			if IsPipeInput(os.Stdin) {
				podSpecFlags.ExtraContainers = "-"
			}
			podSpec := &corev1.PodSpec{}
			if err = podSpecFlags.ResolvePodSpec(podSpec, cmd.Flags(), os.Args); err != nil {
				return err
			}
			// Add container's name to current one
			podSpec.Containers[0].Name = name

			b, err := sigyaml.Marshal(podSpec)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(b)
			return err
		},
	}
	fss := cliflag.NamedFlagSets{}
	experimentalFlagSet := fss.FlagSet("experimental")
	podSpecFlags.AddFlags(cmd.Flags(), experimentalFlagSet)
	podSpecFlags.AddUpdateFlags(cmd.Flags())
	// Volume is not part of ContainerSpec
	cmd.Flag("volume").Hidden = true

	return cmd
}

// IsPipeInput determines the input is created using unix '|' pipe
func IsPipeInput(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}
