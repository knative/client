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
	"fmt"

	"github.com/spf13/cobra"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewCronJobListCommand is for listing CronJob source COs
func NewCronJobListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(CronJobSourceListHandlers)

	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List CronJob sources.",
		Example: `
  # List all CronJob sources
  kn source cronjob list

  # List all CronJob sources in YAML format
  kn source cronjob list -o yaml`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// TODO: filter list by given source name

			cronSourceClient, err := newCronJobSourceClient(p, cmd)
			if err != nil {
				return err
			}

			sourceList, err := cronSourceClient.ListCronJobSource()
			if err != nil {
				return err
			}

			if len(sourceList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No CronJob source found.\n")
				return nil
			}

			if cronSourceClient.Namespace() == "" {
				listFlags.EnsureWithNamespace()
			}

			printer, err := listFlags.ToPrinter()
			if err != nil {
				return nil
			}

			err = printer.PrintObj(sourceList, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(listCommand.Flags(), true)
	listFlags.AddFlags(listCommand)
	return listCommand
}
