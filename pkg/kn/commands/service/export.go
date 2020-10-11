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
	"k8s.io/cli-runtime/pkg/printers"

	clientv1alpha1 "knative.dev/client/pkg/apis/client/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

var IGNORED_SERVICE_ANNOTATIONS = []string{
	"serving.knative.dev/creator",
	"serving.knative.dev/lastModifier",
	"kubectl.kubernetes.io/last-applied-configuration",
}
var IGNORED_REVISION_ANNOTATIONS = []string{
	"serving.knative.dev/lastPinned",
	"serving.knative.dev/creator",
	"serving.knative.dev/routingStateModified",
}

// NewServiceExportCommand returns a new command for exporting a service.
func NewServiceExportCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	command := &cobra.Command{
		Use:   "export NAME",
		Short: "Export a service and its revisions",
		Example: `
  # Export a service in YAML format
  kn service export foo -n bar -o yaml

  # Export a service in JSON format
  kn service export foo -n bar -o json

  # Export a service with revisions
  kn service export foo --with-revisions --mode=export -n bar -o json

  # Export services in kubectl friendly format, as a list kind, one service item for each revision
  kn service export foo --with-revisions --mode=replay -n bar -o json`,
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
			printer, err := machineReadablePrintFlags.ToPrinter()
			if err != nil {
				return err
			}
			return exportService(cmd, service, client, printer)
		},
	}
	flags := command.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.Bool("with-revisions", false, "Export all routed revisions (experimental)")
	flags.String("mode", "", "Format for exporting all routed revisions. One of replay|export (experimental)")
	machineReadablePrintFlags.AddFlags(command)
	return command
}

func exportService(cmd *cobra.Command, service *servingv1.Service, client clientservingv1.KnServingClient, printer printers.ResourcePrinter) error {
	withRevisions, err := cmd.Flags().GetBool("with-revisions")
	if err != nil {
		return err
	}

	if !withRevisions {
		return printer.PrintObj(exportLatestService(service.DeepCopy(), false), cmd.OutOrStdout())
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}

	switch mode {
	case "replay":
		svcList, err := exportServiceListForReplay(service.DeepCopy(), client)
		if err != nil {
			return err
		}
		return printer.PrintObj(svcList, cmd.OutOrStdout())
	case "export":
		knExport, err := exportForKNImport(service.DeepCopy(), client)
		if err != nil {
			return err
		}
		//print kn export
		if err := printer.PrintObj(knExport, cmd.OutOrStdout()); err != nil {
			return err
		}
	default:
		return errors.New("'kn service export --with-revisions' requires a mode, please specify one of replay|export")
	}
	return nil
}

func exportLatestService(latestSvc *servingv1.Service, withRoutes bool) *servingv1.Service {
	exportedSvc := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        latestSvc.ObjectMeta.Name,
			Labels:      latestSvc.ObjectMeta.Labels,
			Annotations: latestSvc.ObjectMeta.Annotations,
		},
		TypeMeta: latestSvc.TypeMeta,
	}

	exportedSvc.Spec.Template = servingv1.RevisionTemplateSpec{
		Spec:       latestSvc.Spec.Template.Spec,
		ObjectMeta: latestSvc.Spec.Template.ObjectMeta,
	}

	if withRoutes {
		exportedSvc.Spec.RouteSpec = latestSvc.Spec.RouteSpec
	}

	stripIgnoredAnnotationsFromService(&exportedSvc)

	return &exportedSvc
}

func exportRevision(revision *servingv1.Revision) servingv1.Revision {
	exportedRevision := servingv1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Name:        revision.ObjectMeta.Name,
			Labels:      revision.ObjectMeta.Labels,
			Annotations: revision.ObjectMeta.Annotations,
		},
		TypeMeta: revision.TypeMeta,
	}

	exportedRevision.Spec = revision.Spec
	stripIgnoredAnnotationsFromRevision(&exportedRevision)
	return exportedRevision
}

func constructServiceFromRevision(latestSvc *servingv1.Service, revision *servingv1.Revision) servingv1.Service {
	exportedSvc := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        latestSvc.ObjectMeta.Name,
			Labels:      latestSvc.ObjectMeta.Labels,
			Annotations: latestSvc.ObjectMeta.Annotations,
		},
		TypeMeta: latestSvc.TypeMeta,
	}
	exportedSvc.Spec.Template = servingv1.RevisionTemplateSpec{
		Spec:       revision.Spec,
		ObjectMeta: latestSvc.Spec.Template.ObjectMeta,
	}

	//overriding revision template annotations with revision annotations
	stripIgnoredAnnotationsFromRevision(revision)
	exportedSvc.Spec.Template.ObjectMeta.Annotations = revision.ObjectMeta.Annotations

	exportedSvc.Spec.Template.ObjectMeta.Name = revision.ObjectMeta.Name
	stripIgnoredAnnotationsFromService(&exportedSvc)
	return exportedSvc
}

func exportServiceListForReplay(latestSvc *servingv1.Service, client clientservingv1.KnServingClient) (*servingv1.ServiceList, error) {
	revisionList, revsMap, err := getRevisionsToExport(latestSvc, client)
	if err != nil {
		return nil, err
	}

	var exportedSvcItems []servingv1.Service

	for _, revision := range revisionList.Items {
		//construct service only for active revisions
		if revsMap[revision.ObjectMeta.Name] && revision.ObjectMeta.Name != latestSvc.Spec.Template.ObjectMeta.Name {
			exportedSvcItems = append(exportedSvcItems, constructServiceFromRevision(latestSvc, revision.DeepCopy()))
		}
	}

	//add latest service, add traffic if more than one revision exist
	exportedSvcItems = append(exportedSvcItems, *(exportLatestService(latestSvc, len(revisionList.Items) > 1)))

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

func exportForKNImport(latestSvc *servingv1.Service, client clientservingv1.KnServingClient) (*clientv1alpha1.Export, error) {
	revisionList, revsMap, err := getRevisionsToExport(latestSvc, client)
	if err != nil {
		return nil, err
	}

	var exportedRevItems []servingv1.Revision

	for _, revision := range revisionList.Items {
		//append only active revisions, no latest revision
		if revsMap[revision.ObjectMeta.Name] && revision.ObjectMeta.Name != latestSvc.Spec.Template.ObjectMeta.Name {
			exportedRevItems = append(exportedRevItems, exportRevision(revision.DeepCopy()))
		}
	}

	typeMeta := metav1.TypeMeta{
		APIVersion: "client.knative.dev/v1alpha1",
		Kind:       "Export",
	}
	knExport := &clientv1alpha1.Export{
		TypeMeta: typeMeta,
		Spec: clientv1alpha1.ExportSpec{
			Service:   *(exportLatestService(latestSvc, len(revisionList.Items) > 1)),
			Revisions: exportedRevItems,
		},
	}

	return knExport, nil
}

func getRevisionsToExport(latestSvc *servingv1.Service, client clientservingv1.KnServingClient) (*servingv1.RevisionList, map[string]bool, error) {
	//get revisions to export from traffic
	revsMap := getRoutedRevisions(latestSvc)

	// Query for list with filters
	revisionList, err := client.ListRevisions(clientservingv1.WithService(latestSvc.ObjectMeta.Name))
	if err != nil {
		return nil, nil, err
	}
	if len(revisionList.Items) == 0 {
		return nil, nil, fmt.Errorf("no revisions found for the service %s", latestSvc.ObjectMeta.Name)
	}
	// sort revisions to maintain the order of generations
	sortRevisions(revisionList)
	return revisionList, revsMap, nil
}

func getRoutedRevisions(latestSvc *servingv1.Service) map[string]bool {
	trafficList := latestSvc.Spec.RouteSpec.Traffic
	revsMap := make(map[string]bool)

	for _, traffic := range trafficList {
		if traffic.RevisionName != "" {
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

func stripIgnoredAnnotationsFromService(svc *servingv1.Service) {
	for _, annotation := range IGNORED_SERVICE_ANNOTATIONS {
		delete(svc.ObjectMeta.Annotations, annotation)
	}
}

func stripIgnoredAnnotationsFromRevision(revision *servingv1.Revision) {
	for _, annotation := range IGNORED_REVISION_ANNOTATIONS {
		delete(revision.ObjectMeta.Annotations, annotation)
	}
}
