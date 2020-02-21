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

package route

import (
	"errors"
	"fmt"

	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"

	"github.com/spf13/cobra"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands/flags"
)

// NewrouteListCommand represents 'kn route list' command
func NewRouteListCommand(p *commands.KnParams) *cobra.Command {
	routeListFlags := flags.NewListPrintFlags(RouteListHandlers)
	routeListCommand := &cobra.Command{
		Use:   "list NAME",
		Short: "List available routes.",
		Example: `
  # List all routes
  kn route list

  # List route 'web' in namespace 'dev'
  kn route list web -n dev

  # List all routes in YAML format
  kn route list -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			var routeList *servingv1.RouteList
			switch len(args) {
			case 0:
				routeList, err = client.ListRoutes()
			case 1:
				routeList, err = client.ListRoutes(clientservingv1.WithName(args[0]))
			default:
				return errors.New("'kn route list' accepts only one additional argument")
			}
			if err != nil {
				return err
			}
			if len(routeList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No routes found.\n")
				return nil
			}
			printer, err := routeListFlags.ToPrinter()
			if err != nil {
				return err
			}
			err = printer.PrintObj(routeList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(routeListCommand.Flags(), true)
	routeListFlags.AddFlags(routeListCommand)
	return routeListCommand
}
