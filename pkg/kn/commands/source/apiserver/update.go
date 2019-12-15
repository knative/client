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

package apiserver

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

func NewApiServerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var apiServerUpdateFlags ApiServerSourceUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME --resource RESOURCE --service-account ACCOUNTNAME --sink SINK --mode MODE",
		Short: "Update an ApiServerSource, which watches for Kubernetes events and forwards them to a sink",
		Example: `
  # Update an ApiServerSource 'k8sevents' with different service account and sink service
  kn source apiserver update k8sevents --service-account newsa --sink svc:newsvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the source as single argument")
			}
			name := args[0]

			// get namespace
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			// get client
			sourcesClient, err := newApiServerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			source, err := sourcesClient.GetApiServerSource(name)
			if err != nil {
				return err
			}

			// resolve sink
			servingClient, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			objectRef, err := sinkFlags.ResolveSink(servingClient)
			if err != nil {
				return fmt.Errorf(
					"cannot update ApiServerSource '%s' in namespace '%s' "+
						"because %v", name, namespace, err)
			}

			source = source.DeepCopy()
			apiServerUpdateFlags.Apply(source, cmd)
			source.Spec.Sink = objectRef

			err = sourcesClient.UpdateApiServerSource(source)
			if err != nil {
				return fmt.Errorf(
					"cannot create ApiServerSource '%s' in namespace '%s' "+
						"because %s", name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ApiServerSource '%s' updated in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	apiServerUpdateFlags.Add(cmd)
	sinkFlags.Add(cmd)

	return cmd
}
