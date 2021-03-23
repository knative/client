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
	"knative.dev/client/pkg/sources/v1alpha2"
	"knative.dev/client/pkg/util"
)

// NewAPIServerUpdateCommand for managing source update
func NewAPIServerUpdateCommand(p *commands.KnParams) *cobra.Command {
	var updateFlags APIServerSourceUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update an api-server source",
		Example: `
  # Update an ApiServerSource 'k8sevents' with different service account and sink service
  kn source apiserver update k8sevents --service-account newsa --sink ksvc:newsvc`,

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

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			sourcesClient, err := newAPIServerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			source, err := sourcesClient.GetAPIServerSource(name)
			if err != nil {
				return err
			}
			if source.GetDeletionTimestamp() != nil {
				return fmt.Errorf("can't update apiserver source %s because it has been marked for deletion", name)
			}

			b := v1alpha2.NewAPIServerSourceBuilderFromExisting(source)
			if cmd.Flags().Changed("service-account") {
				b.ServiceAccount(updateFlags.ServiceAccountName)
			}

			if cmd.Flags().Changed("mode") {
				b.EventMode(updateFlags.Mode)
			}

			if cmd.Flags().Changed("resource") {
				updateExisting, err := updateFlags.updateExistingAPIVersionKindSelectorArray(source.Spec.Resources)
				if err != nil {
					return err
				}
				b.Resources(updateExisting)
			}

			if cmd.Flags().Changed("sink") {
				objectRef, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
				if err != nil {
					return err
				}
				b.Sink(*objectRef)
			}

			if cmd.Flags().Changed("ce-override") {
				ceOverridesMap, err := util.MapFromArrayAllowingSingles(updateFlags.ceOverrides, "=")
				if err != nil {
					return err
				}
				ceOverridesToRemove := util.ParseMinusSuffix(ceOverridesMap)
				b.CloudEventOverrides(ceOverridesMap, ceOverridesToRemove)
			}

			err = sourcesClient.UpdateAPIServerSource(b.Build())
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "ApiServer source '%s' updated in namespace '%s'.\n", args[0], namespace)
			}

			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	updateFlags.Add(cmd)
	sinkFlags.Add(cmd)
	return cmd
}
