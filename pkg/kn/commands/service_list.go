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
	"fmt"
	"strings"
	"text/tabwriter"

	printers "github.com/knative/client/pkg/util/printers"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NewServiceListCommand represents the list command
func NewServiceListCommand(p *KnParams) *cobra.Command {
	serviceListCommand := &cobra.Command{
		Use:   "list",
		Short: "List available services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := GetNamespace(cmd)
			if err != nil {
				return err
			}
			service, err := client.Services(namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}

			service.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"})

			printer := printers.NewTabWriter(cmd.OutOrStdout())
			defer printer.Flush()

			if err := printServiceList(printer, *service); err != nil {
				return err
			}
			return nil
		},
	}
	AddNamespaceFlags(serviceListCommand.Flags(), true)
	return serviceListCommand
}

func printServiceList(printer *tabwriter.Writer, services servingv1alpha1.ServiceList) error {
	// case where no services are present
	if len(services.Items) < 1 {
		fmt.Fprintln(printer, "No resources found.")
		return nil
	}
	columnNames := []string{"NAME", "DOMAIN", "LATESTCREATED", "LATESTREADY", "AGE"}
	if _, err := fmt.Fprintf(printer, "%s\n", strings.Join(columnNames, "\t")); err != nil {
		return err
	}
	for _, ksvc := range services.Items {
		_, err := fmt.Fprintf(printer, "%s\n", strings.Join([]string{ksvc.Name, ksvc.Status.Domain,
			ksvc.Status.LatestCreatedRevisionName, ksvc.Status.LatestReadyRevisionName,
			printers.CalculateAge(ksvc.CreationTimestamp.Time)}, "\t"))
		if err != nil {
			return err
		}
	}
	return nil
}
