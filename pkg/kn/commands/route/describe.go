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
	"io"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

// NewRouteDescribeCommand represents 'kn route describe' command
func NewRouteDescribeCommand(p *commands.KnParams) *cobra.Command {
	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")
	command := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a route",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn route describe' requires name of the route as single argument")
			}
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			route, err := client.GetRoute(args[0])
			if err != nil {
				return err
			}

			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(route, cmd.OutOrStdout())
			}
			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}
			return describe(cmd.OutOrStdout(), route, printDetails)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	machineReadablePrintFlags.AddFlags(command)
	flags.BoolP("verbose", "v", false, "More output.")
	return command
}

func describe(w io.Writer, route *servingv1.Route, printDetails bool) error {
	dw := printers.NewPrefixWriter(w)
	commands.WriteMetadata(dw, &route.ObjectMeta, printDetails)
	dw.WriteAttribute("URL", route.Status.URL.String())
	writeService(dw, route, printDetails)
	dw.WriteLine()
	writeTraffic(dw, route)
	dw.WriteLine()
	commands.WriteConditions(dw, route.Status.Conditions, printDetails)
	if err := dw.Flush(); err != nil {
		return err
	}
	return nil
}

func writeService(dw printers.PrefixWriter, route *servingv1.Route, printDetails bool) {
	svcName := ""
	for _, owner := range route.ObjectMeta.OwnerReferences {
		if owner.Kind == "Service" {
			svcName = owner.Name
			if printDetails {
				svcName = fmt.Sprintf("%s (%s)", svcName, owner.APIVersion)
			}
		}

		dw.WriteAttribute("Service", svcName)
	}
}

func writeTraffic(dw printers.PrefixWriter, route *servingv1.Route) {
	trafficSection := dw.WriteAttribute("Traffic Targets", "")
	dw.Flush()
	for _, target := range route.Status.Traffic {
		section := trafficSection.WriteColsLn(fmt.Sprintf("%3d%%", *target.Percent), formatTarget(target))
		if target.Tag != "" {
			section.WriteAttribute("URL", target.URL.String())
		}
	}
}

func formatTarget(target servingv1.TrafficTarget) string {
	targetHeader := target.RevisionName
	if target.LatestRevision != nil && *target.LatestRevision {
		targetHeader = fmt.Sprintf("@latest (%s)", target.RevisionName)
	}
	if target.Tag != "" {
		targetHeader = fmt.Sprintf("%s #%s", targetHeader, target.Tag)
	}
	return targetHeader
}
