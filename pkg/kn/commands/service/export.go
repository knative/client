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

			history, err := cmd.Flags().GetBool("history")
			if err != nil {
				return err
			}

			// Print out machine readable output if requested
			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				if history {
					svcList, err := exportServicewithActiveRevisions(service, client)
					if err != nil {
						return err
					}
					return printer.PrintObj(svcList, cmd.OutOrStdout())
				}
				return printer.PrintObj(exportService(service), cmd.OutOrStdout())
			}

			return nil
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("history", "r", false, "Export all active revisions")
	machineReadablePrintFlags.AddFlags(command)
	return command
}

func exportService(latestService *servingv1.Service) *servingv1.Service {
	return constructServiceTemplate(latestService)
}

func constructServiceTemplate(latestSvc *servingv1.Service) *servingv1.Service {

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

func exportServicewithActiveRevisions(latestSvc *servingv1.Service, client clientservingv1.KnServingClient) (*servingv1.ServiceList, error) {
	var exportedSvcItems []servingv1.Service

	//get revisions to export from traffic
	revsMap := getRevisionstoExport(latestSvc)

	var params []clientservingv1.ListConfig
	params = append(params, clientservingv1.WithService(latestSvc.ObjectMeta.Name))

	// Query for list with filters
	revisionList, err := client.ListRevisions(params...)
	if err != nil {
		return nil, err
	}

	sortRevisions(revisionList)

	for _, revision := range revisionList.Items {
		//construct service only for active revisions
		if revsMap[revision.ObjectMeta.Name] != nil {
			exportedSvcItems = append(exportedSvcItems, constructServicefromRevision(latestSvc, revision))
		}
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

func getRevisionstoExport(latestSvc *servingv1.Service) map[string]*int64 {
	trafficList := latestSvc.Spec.RouteSpec.Traffic
	revsMap := make(map[string]*int64)

	for _, traffic := range trafficList {
		if traffic.RevisionName == "" {
			revsMap[latestSvc.Spec.ConfigurationSpec.Template.ObjectMeta.Name] = traffic.Percent
		} else {
			revsMap[traffic.RevisionName] = traffic.Percent
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
