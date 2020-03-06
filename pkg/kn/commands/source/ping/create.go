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

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	"knative.dev/client/pkg/sources/v1alpha2"
)

// NewPingCreateCommand is for creating Ping source COs
func NewPingCreateCommand(p *commands.KnParams) *cobra.Command {
	var pingUpdateFlags pingUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --schedule SCHEDULE --sink SINK --data DATA",
		Short: "Create a Ping source.",
		Example: `
  # Create a Ping source 'my-ping' which fires every two minutes and sends '{ value: "hello" }' to service 'mysvc' as a cloudevent
  kn source ping create my-ping --schedule "*/2 * * * *" --data '{ value: "hello" }' --sink svc:mysvc`,

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

			destination, err := sinkFlags.ResolveSink(dynamicClient, namespace)
			if err != nil {
				return err
			}

			err = pingSourceClient.CreatePingSource(
				v1alpha2.NewPingSourceBuilder(name).
					Schedule(pingUpdateFlags.schedule).
					JsonData(pingUpdateFlags.data).
					Sink(*destination).
					Build())
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Ping source '%s' created in namespace '%s'.\n", args[0], pingSourceClient.Namespace())
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	pingUpdateFlags.addPingFlags(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("sink")

	return cmd
}
