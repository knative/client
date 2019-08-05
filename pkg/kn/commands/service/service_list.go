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
	"fmt"
	"sort"

	"github.com/knative/client/pkg/kn/commands"
	v1alpha12 "github.com/knative/client/pkg/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
)

// NewServiceListCommand represents 'kn service list' command
func NewServiceListCommand(p *commands.KnParams) *cobra.Command {
	serviceListFlags := NewServiceListFlags()

	serviceListCommand := &cobra.Command{
		Use:   "list [name]",
		Short: "List available services.",
		Example: `
  # List all services
  kn service list

  # List all services in JSON output format
  kn service list -o json

  # List service 'web'
  kn service list web`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewClient(namespace)
			if err != nil {
				return err
			}
			serviceList, err := getServiceInfo(args, client)
			if err != nil {
				return err
			}
			if len(serviceList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
				return nil
			}
			printer, err := serviceListFlags.ToPrinter()
			if err != nil {
				return err
			}

			// Sort serviceList by name
			sort.SliceStable(serviceList.Items, func(i, j int) bool {
				return serviceList.Items[i].ObjectMeta.Name < serviceList.Items[j].ObjectMeta.Name
			})

			err = printer.PrintObj(serviceList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(serviceListCommand.Flags(), true)
	serviceListFlags.AddFlags(serviceListCommand)
	return serviceListCommand
}

func getServiceInfo(args []string, client v1alpha12.KnClient) (*v1alpha1.ServiceList, error) {
	var (
		serviceList *v1alpha1.ServiceList
		err         error
	)
	switch len(args) {
	case 0:
		serviceList, err = client.ListServices()
	case 1:
		serviceList, err = client.ListServices(v1alpha12.WithName(args[0]))
	default:
		return nil, fmt.Errorf("'kn service list' accepts maximum 1 argument")
	}
	return serviceList, err
}
