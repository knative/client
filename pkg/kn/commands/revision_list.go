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

	knserving "github.com/knative/client/pkg/serving"
	util "github.com/knative/client/pkg/util"
	printers "github.com/knative/client/pkg/util/printers"
	v1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRevisionListCommand represent the 'revision list' command
func NewRevisionListCommand(p *KnParams) *cobra.Command {
	revisionListCmd := &cobra.Command{
		Use:   "list",
		Short: "List available revisions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace, err := GetNamespace(cmd)
			if err != nil {
				return err
			}
			revisions, err := client.Revisions(namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}

			routes, err := client.Routes(namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}

			printer := printers.GetNewTabWriter(cmd.OutOrStdout())
			// make sure the printer is flushed to stdout before returning
			defer printer.Flush()

			if err := printRevisionList(printer, *revisions, *routes); err != nil {
				return err
			}
			return nil
		},
	}
	AddNamespaceFlags(revisionListCmd.Flags(), true)
	return revisionListCmd
}

// printRevisionList takes care of printing revisions
func printRevisionList(
	printer *tabwriter.Writer,
	revisions v1alpha1.RevisionList,
	routes v1alpha1.RouteList) error {
	// case where no revisions are present
	if len(revisions.Items) < 1 {
		fmt.Fprintln(printer, "No resources found.")
		return nil
	}
	columnNames := []string{"NAME", "SERVICE", "AGE", "TRAFFIC"}
	if _, err := fmt.Fprintf(printer, "%s\n", strings.Join(columnNames, "\t")); err != nil {
		return err
	}
	for _, rev := range revisions.Items {
		row := []string{
			rev.Name,
			rev.Labels[knserving.ConfigurationLabelKey],
			util.CalculateAge(rev.CreationTimestamp.Time),
			// RouteTrafficValue returns comma separated traffic string
			RouteTrafficValue(rev, routes.Items),
		}
		if _, err := fmt.Fprintf(printer, "%s\n", strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return nil
}

// RouteTrafficValue returns a string with comma separated traffic for revision
func RouteTrafficValue(rev v1alpha1.Revision, routes []v1alpha1.Route) string {
	var traffic []string
	for _, route := range routes {
		for _, target := range route.Status.Traffic {
			if target.RevisionName == rev.Name {
				traffic = append(traffic, fmt.Sprintf("%d%% -> %s", target.Percent, route.Status.Domain))
			}
		}
	}
	return strings.Join(traffic, ", ")
}
