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

package cronjob

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

func NewCronJobUpdateCommand(p *commands.KnParams) *cobra.Command {
	var cronUpdateFlags cronJobUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "update NAME --schedule SCHEDULE --sink SERVICE --data DATA",
		Short: "Update a Cronjob source.",
		Example: `
  # Update the schedule of a crontab source 'my-cron-trigger' to fire every minute
  kn source cronjob update my-cron-trigger --schedule "* * * * */1"`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("name of cronjob source required")
			}
			name := args[0]

			cronSourceClient, err := newCronJobSourceClient(p, cmd)
			if err != nil {
				return err
			}

			servingClient, err := p.NewServingClient(cronSourceClient.Namespace())
			if err != nil {
				return err
			}

			destination, err := sinkFlags.ResolveSink(servingClient)
			if err != nil {
				return err
			}
			err = cronSourceClient.UpdateCronJobSource(name, cronUpdateFlags.schedule, cronUpdateFlags.data, destination)
			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Cronjob source '%s' updated in namespace '%s'.\n", name, cronSourceClient.Namespace())
			}
			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	cronUpdateFlags.addCronJobFlags(cmd)
	sinkFlags.Add(cmd)

	return cmd
}
