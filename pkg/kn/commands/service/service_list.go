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

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewServiceListCommand represents 'kn service list' command
func NewServiceListCommand(p *commands.KnParams) *cobra.Command {
	serviceListFlags := NewServiceListFlags()

	serviceListCommand := &cobra.Command{
		Use:   "list",
		Short: "List available services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			service, err := client.Services(namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}
			if len(service.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No resources found.\n")
				return nil
			}
			service.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"})

			printer, err := serviceListFlags.ToPrinter()
			if err != nil {
				return err
			}

			err = printer.PrintObj(service, cmd.OutOrStdout())
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
