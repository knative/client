// Copyright © 2018 The Knative Authors
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

package commands

import (
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var serviceListPrintFlags *genericclioptions.PrintFlags

// NewServiceListCommand represents the service list command
func NewServiceListCommand(p *KnParams) *cobra.Command {
	serviceListPrintFlags := genericclioptions.NewPrintFlags("").WithDefaultOutput(
		"jsonpath={range .items[*]}{.metadata.namespace}{\"\\t\"}{.metadata.name}{\"\\n\"}{end}")
	serviceListCommand := &cobra.Command{
		Use:   "list",
		Short: "List available services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace := cmd.Flag("namespace").Value.String()
			service, err := client.Services(namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}

			printer, err := serviceListPrintFlags.ToPrinter()
			if err != nil {
				return err
			}
			service.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"})
			return printer.PrintObj(service, cmd.OutOrStdout())
		},
	}
	serviceListPrintFlags.AddFlags(serviceListCommand)
	return serviceListCommand
}
