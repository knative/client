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

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/commands"
	"knative.dev/client/pkg/commands/flags"
	clientsourcesv1 "knative.dev/client/pkg/sources/v1"
	"knative.dev/client/pkg/util"
)

// NewPingCreateCommand is for creating Ping source COs
func NewPingCreateCommand(p *commands.KnParams) *cobra.Command {
	var updateFlags pingUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --sink SINK",
		Short: "Create a ping source",
		Example: `
  # Create a Ping source 'my-ping' which fires every two minutes and sends '{ value: "hello" }' to service 'mysvc' as a cloudevent
  kn source ping create my-ping --schedule "*/2 * * * *" --data '{ value: "hello" }' --sink ksvc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the Ping source to create as single argument")

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

			destination, err := sinkFlags.ResolveSink(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return err
			}

			ceOverridesMap, err := util.MapFromArrayAllowingSingles(updateFlags.ceOverrides, "=")
			if err != nil {
				return err
			}
			ceOverridesToRemove := util.ParseMinusSuffix(ceOverridesMap)

			data, dataBase64, err := getDataFields(&updateFlags)
			if err != nil {
				return fmt.Errorf("cannot create PingSource %q in namespace "+
					"%q because: %s", name, namespace, err)
			}

			err = pingSourceClient.CreatePingSource(cmd.Context(), clientsourcesv1.NewPingSourceBuilder(name).
				Schedule(updateFlags.schedule).
				Data(data).
				DataBase64(dataBase64).
				Sink(*destination).
				CloudEventOverrides(ceOverridesMap, ceOverridesToRemove).
				Build())
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Ping source '%s' created in namespace '%s'.\n", args[0], pingSourceClient.Namespace())
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	updateFlags.addFlags(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("sink")

	return cmd
}
