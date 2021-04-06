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

package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

// NewServiceListCommand represents 'kn service list' command
func NewServiceListCommand(p *commands.KnParams) *cobra.Command {
	serviceListFlags := flags.NewListPrintFlags(ServiceListHandlers)

	serviceListCommand := &cobra.Command{
		Use:     "list",
		Short:   "List services",
		Aliases: []string{"ls"},
		Example: `
  # List all services
  kn service list

  # List all services in JSON output format
  kn service list -o json

  # List service 'web'
  kn service list web

  # List the services in offline mode instead of kubernetes cluster
  kn service list --target=/user/knfiles
  kn service list --target=/user/knfiles/test.json
  kn service list --target=/user/knfiles/test.yaml
  kn service list -n test-ns --target=/user/knfiles`,

		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := newServingClient(p, namespace, cmd.Flag("target").Value.String())
			if err != nil {
				return err
			}
			serviceList, err := getServiceInfo(cmd.Context(), args, client)
			if err != nil {
				return err
			}
			if len(serviceList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No services found.\n")
				return nil
			}

			// empty namespace indicates all-namespaces flag is specified
			if namespace == "" {
				serviceListFlags.EnsureWithNamespace()
			}

			// Sort serviceList by namespace and name (in this order)
			sort.SliceStable(serviceList.Items, func(i, j int) bool {
				a := serviceList.Items[i]
				b := serviceList.Items[j]

				if a.Namespace != b.Namespace {
					return a.Namespace < b.Namespace
				}
				return a.ObjectMeta.Name < b.ObjectMeta.Name
			})

			return serviceListFlags.Print(serviceList, cmd.OutOrStdout())
		},
	}
	commands.AddNamespaceFlags(serviceListCommand.Flags(), true)
	commands.AddGitOpsFlags(serviceListCommand.Flags())
	serviceListFlags.AddFlags(serviceListCommand)
	return serviceListCommand
}

func getServiceInfo(ctx context.Context, args []string, client clientservingv1.KnServingClient) (*servingv1.ServiceList, error) {
	var (
		serviceList *servingv1.ServiceList
		err         error
	)
	switch len(args) {
	case 0:
		serviceList, err = client.ListServices(ctx)
	case 1:
		serviceList, err = client.ListServices(ctx, clientservingv1.WithName(args[0]))
	default:
		return nil, fmt.Errorf("'kn service list' accepts maximum 1 argument")
	}
	return serviceList, err
}
