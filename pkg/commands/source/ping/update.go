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

package ping

import (
	"errors"
	"fmt"

	"knative.dev/client/pkg/config"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/commands"
	"knative.dev/client/pkg/commands/flags"
	sourcesv1beta2 "knative.dev/client/pkg/sources/v1beta2"
	eventingsourcesv1beta2 "knative.dev/eventing/pkg/apis/sources/v1beta2"

	"knative.dev/client/pkg/util"
)

// NewPingUpdateCommand prepares the command for a PingSource update
func NewPingUpdateCommand(p *commands.KnParams) *cobra.Command {
	var updateFlags pingUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a ping source",
		Example: `
  # Update the schedule of a Ping source 'my-ping' to fire every minute
  kn source ping update my-ping --schedule "* * * * *"`,

		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("name of Ping source required")
			}
			name := args[0]

			pingSourceClient, err := newPingSourceClient(p, cmd)
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

			updateFunc := func(origSource *eventingsourcesv1beta2.PingSource) (*eventingsourcesv1beta2.PingSource, error) {
				b := sourcesv1beta2.NewPingSourceBuilderFromExisting(origSource)
				if cmd.Flags().Changed("schedule") {
					b.Schedule(updateFlags.schedule)
				}

				data, dataBase64, err := getDataFields(&updateFlags)
				if err != nil {
					return nil, fmt.Errorf("cannot update PingSource %q in namespace "+
						"%q because: %s", name, namespace, err)
				}
				if cmd.Flags().Changed("data") {
					b.Data(data).DataBase64(dataBase64)
				}
				if cmd.Flags().Changed("sink") {
					destination, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
					if err != nil {
						return nil, err
					}
					b.Sink(*destination)
				}
				if cmd.Flags().Changed("ce-override") {
					ceOverridesMap, err := util.MapFromArrayAllowingSingles(updateFlags.ceOverrides, "=")
					if err != nil {
						return nil, err
					}
					ceOverridesToRemove := util.ParseMinusSuffix(ceOverridesMap)
					b.CloudEventOverrides(ceOverridesMap, ceOverridesToRemove)
				}
				updatedSource := b.Build()
				return updatedSource, nil
			}

			err = pingSourceClient.UpdatePingSourceWithRetry(cmd.Context(), name, updateFunc, config.DefaultRetry.Steps)

			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Ping source '%s' updated in namespace '%s'.\n", name, pingSourceClient.Namespace())
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	updateFlags.addFlags(cmd)
	sinkFlags.Add(cmd)

	return cmd
}
