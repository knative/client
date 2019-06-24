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
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			var listOptions v1.ListOptions
			switch len(args) {
			case 0:
				listOptions = v1.ListOptions{}
			case 1:
				listOptions.FieldSelector = fields.Set(map[string]string{"metadata.name": args[0]}).String()
			default:
				return fmt.Errorf("'kn route list' accepts maximum 1 argument.")
			}
			route, err := client.Routes(namespace).List(listOptions)
			if err != nil {
				return err
			}
			if len(route.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
				return nil
			}
			route.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "route",
			})
			printer, err := routeListFlags.ToPrinter()
			if err != nil {
				return err
			}
			err = printer.PrintObj(route, cmd.OutOrStdout())
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
