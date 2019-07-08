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
	"fmt"

	"github.com/knative/client/pkg/kn/commands"
	v1alpha12 "github.com/knative/client/pkg/serving/v1alpha1"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
)

// NewrouteListCommand represents 'kn route list' command
func NewRouteListCommand(p *commands.KnParams) *cobra.Command {
	routeListFlags := NewRouteListFlags()
	routeListCommand := &cobra.Command{
		Use:   "list NAME",
		Short: "List available routes.",
		Example: `
  # List all routes
  kn route list

  # List route 'web' in namespace 'dev'
  kn route list web -n dev

  # List all routes in yaml format
  kn route list -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}

			var routeList *v1alpha1.RouteList
			switch len(args) {
			case 0:
				routeList, err = client.ListRoutes()
			case 1:
				routeList, err = client.ListRoutes(v1alpha12.WithName(args[0]))
			default:
				return fmt.Errorf("'kn route list' accepts maximum 1 argument.")
			}
			if err != nil {
				return err
			}
			if len(routeList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
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
