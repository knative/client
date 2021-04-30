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
	v1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
)

// NewAPIServerCreateCommand for creating source
func NewAPIServerCreateCommand(p *commands.KnParams) *cobra.Command {
	var updateFlags APIServerSourceUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --resource RESOURCE --sink SINK",
		Short: "Create an api-server source",
		Example: `
  # Create an ApiServerSource 'k8sevents' which consumes Kubernetes events and sends message to service 'mysvc' as a cloudevent
  kn source apiserver create k8sevents --resource Event:v1 --service-account myaccountname --sink ksvc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the source to create as single argument")
			}
			name := args[0]

			// get client
			apiSourceClient, err := newAPIServerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			namespace := apiSourceClient.Namespace()

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}
			objectRef, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return fmt.Errorf(
					"cannot create ApiServerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			resources, err := updateFlags.getAPIServerVersionKindSelector()
			if err != nil {
				return err
			}

			ceOverridesMap, err := util.MapFromArrayAllowingSingles(updateFlags.ceOverrides, "=")
			if err != nil {
				return err
			}
			ceOverridesToRemove := util.ParseMinusSuffix(ceOverridesMap)

			b := v1.NewAPIServerSourceBuilder(name).
				ServiceAccount(updateFlags.ServiceAccountName).
				EventMode(updateFlags.Mode).
				Sink(*objectRef).
				Resources(resources).
				CloudEventOverrides(ceOverridesMap, ceOverridesToRemove)

			err = apiSourceClient.CreateAPIServerSource(cmd.Context(), b.Build())

			if err != nil {
				return fmt.Errorf(
					"cannot create ApiServerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "ApiServer source '%s' created in namespace '%s'.\n", args[0], namespace)
			}

			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	updateFlags.Add(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("resource")
	cmd.MarkFlagRequired("sink")
	return cmd
}
