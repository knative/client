// Copyright © 2020 The Knative Authors
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
	"errors"
	"fmt"

	"sort"
	"strconv"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// NewServiceExportCommand returns a new command for exporting a service.
func NewServiceExportCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:   "export NAME",
		Short: "export a service",
		Example: `
  # Export a service in yaml format
  kn service export foo -n bar -o yaml
  # Export a service in json format
  kn service export foo -n bar -o json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn service export' requires name of the service as single argument")
			}
			if !machineReadablePrintFlags.OutputFlagSpecified() {
				return errors.New("'kn service export' requires output format")
			}
			serviceName := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			service, err := client.GetService(serviceName)
			if err != nil {
				return err
			}

			withRevisions, err := cmd.Flags().GetBool("with-revisions")
			if err != nil {
				return err
			}

			printer, err := machineReadablePrintFlags.ToPrinter()
			if err != nil {
				return err
			}

			if withRevisions {
				if svcList, err := exportServiceWithActiveRevisions(service, client); err != nil {
					return err
				} else {
					return printer.PrintObj(svcList, cmd.OutOrStdout())
				}
			}
			return printer.PrintObj(exportService(service), cmd.OutOrStdout())
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.Bool("with-revisions", false, "Export all routed revisions (experimental)")
	machineReadablePrintFlags.AddFlags(command)
	return command
}

func exportService(latestSvc *servingv1.Service) *servingv1.Service {

	exportedSvc := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   latestSvc.ObjectMeta.Name,
			Labels: latestSvc.ObjectMeta.Labels,
		},
		TypeMeta: latestSvc.TypeMeta,
	}

	exportedSvc.Spec.Template = servingv1.RevisionTemplateSpec{
		Spec:       latestSvc.Spec.ConfigurationSpec.Template.Spec,
		ObjectMeta: latestSvc.Spec.ConfigurationSpec.Template.ObjectMeta,
	}

	return &exportedSvc
}

func constructServicefromRevision(latestSvc *servingv1.Service, revision servingv1.Revision) servingv1.Service {

	exportedSvc := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   latestSvc.ObjectMeta.Name,
			Labels: latestSvc.ObjectMeta.Labels,
		},
		TypeMeta: latestSvc.TypeMeta,
	}

	exportedSvc.Spec.Template = servingv1.RevisionTemplateSpec{
		Spec:       revision.Spec,
		ObjectMeta: latestSvc.Spec.ConfigurationSpec.Template.ObjectMeta,
	}

	exportedSvc.Spec.ConfigurationSpec.Template.ObjectMeta.Name = revision.ObjectMeta.Name

	return exportedSvc
}

func exportServiceWithActiveRevisions(latestSvc *servingv1.Service, client clientservingv1.KnServingClient) (*servingv1.ServiceList, error) {
	var exportedSvcItems []servingv1.Service

	//get revisions to export from traffic
	revsMap := getRevisionstoExport(latestSvc)

	// Query for list with filters
	revisionList, err := client.ListRevisions(clientservingv1.WithService(latestSvc.ObjectMeta.Name))
	if err != nil {
		return nil, err
	}
	if len(revisionList.Items) == 0 {
		return nil, fmt.Errorf("no revisions found for the service %s", latestSvc.ObjectMeta.Name)
	}

	// sort revisions to main the order of generations
	sortRevisions(revisionList)

	for _, revision := range revisionList.Items {
		//construct service only for active revisions
		if revsMap[revision.ObjectMeta.Name] {
			exportedSvcItems = append(exportedSvcItems, constructServicefromRevision(latestSvc, revision))
		}
	}

	if len(exportedSvcItems) == 0 {
		return nil, fmt.Errorf("no revisions found for service %s", latestSvc.ObjectMeta.Name)
	}

	//set traffic in the latest revision
	exportedSvcItems[len(exportedSvcItems)-1] = setTrafficSplit(latestSvc, exportedSvcItems[len(exportedSvcItems)-1])

	typeMeta := metav1.TypeMeta{
		APIVersion: "v1",
		Kind:       "List",
	}
	exportedSvcList := &servingv1.ServiceList{
		TypeMeta: typeMeta,
		Items:    exportedSvcItems,
	}

	return exportedSvcList, nil
}

func setTrafficSplit(latestSvc *servingv1.Service, exportedSvc servingv1.Service) servingv1.Service {

	exportedSvc.Spec.RouteSpec = latestSvc.Spec.RouteSpec
	return exportedSvc
}

func getRevisionstoExport(latestSvc *servingv1.Service) map[string]bool {
	trafficList := latestSvc.Spec.RouteSpec.Traffic
	revsMap := make(map[string]bool)

	for _, traffic := range trafficList {
		if traffic.RevisionName == "" {
			revsMap[latestSvc.Spec.ConfigurationSpec.Template.ObjectMeta.Name] = true
		} else {
			revsMap[traffic.RevisionName] = true
		}
	}
	return revsMap
}

// sortRevisions sorts revisions by generation and name (in this order)
func sortRevisions(revisionList *servingv1.RevisionList) {
	// sort revisionList by configuration generation key
	sort.SliceStable(revisionList.Items, revisionListSortFunc(revisionList))
}

// revisionListSortFunc sorts by generation and name
func revisionListSortFunc(revisionList *servingv1.RevisionList) func(i int, j int) bool {
	return func(i, j int) bool {
		a := revisionList.Items[i]
		b := revisionList.Items[j]

		// By Generation
		// Convert configuration generation key from string to int for avoiding string comparison.
		agen, err := strconv.Atoi(a.Labels[serving.ConfigurationGenerationLabelKey])
		if err != nil {
			return a.Name > b.Name
		}
		bgen, err := strconv.Atoi(b.Labels[serving.ConfigurationGenerationLabelKey])
		if err != nil {
			return a.Name > b.Name
		}

		if agen != bgen {
			return agen < bgen
		}
		return a.Name > b.Name
	}
}
