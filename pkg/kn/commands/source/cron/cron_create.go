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

package cron

import (
	"errors"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/source/sink"
)

func NewCronCreateCommand(p *commands.KnParams) *cobra.Command {
	var cronUpdateFlags cronUpdateFlags
	var sinkFlags sink.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --crontab SCHEDULE --sink SERVICE --data DATA",
		Short: "Create an Crontab scheduler as importer",
		Example: `
  # Create a crontabs scheduler 'mycron' which fires every minute and sends 'ping'' to service 'mysvc' as a cloudevent
  kn importer cron create mycron --schedule "* * * * */1" --data "ping" --sink svc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'importer cron create' requires the name of the importer as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			servingClient, err := p.NewClient(namespace)
			if err != nil {
				return err
			}

			config, err := p.GetClientConfig()
			if err != nil {
				return err
			}
			cronSourceClient, err := NewCronSourceClient(config, namespace)

			objectRef, err := sinkFlags.ResolveSink(servingClient)
			if err != nil {
				return err
			}
			return cronSourceClient.CreateCronSource(name, cronUpdateFlags.schedule, cronUpdateFlags.data, objectRef)
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	cronUpdateFlags.addCronFlags(cmd)
	sinkFlags.AddSinkFlag(cmd)
	cmd.MarkFlagRequired("schedule")

	return cmd
}
