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
)

// NewCronJobDeleteCommand is for deleting a CronJob source
func NewCronJobDeleteCommand(p *commands.KnParams) *cobra.Command {
	CronJobDeleteCommand := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a CronJob source.",
		Example: `
  # Delete a CronJob source 'my-cron-trigger'
  kn source cronjob delete my-cron-trigger`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'requires the name of the cronjob source to delete as single argument")
			}
			name := args[0]

			cronSourceClient, err := newCronJobSourceClient(p, cmd)
			if err != nil {
				return err
			}

			err = cronSourceClient.DeleteCronJobSource(name)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "CronJob source '%s' deleted in namespace '%s'.\n", name, cronSourceClient.Namespace())
			return nil
		},
	}
	commands.AddNamespaceFlags(CronJobDeleteCommand.Flags(), false)
	return CronJobDeleteCommand
}
