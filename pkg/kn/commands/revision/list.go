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

package revision

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"knative.dev/serving/pkg/apis/serving"

	"github.com/spf13/cobra"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	clientservingv1alpha "knative.dev/client/pkg/serving/v1"
)

// Service name filter, used with "-s"
var serviceNameFilter string

// NewRevisionListCommand represents 'kn revision list' command
func NewRevisionListCommand(p *commands.KnParams) *cobra.Command {
	revisionListFlags := flags.NewListPrintFlags(RevisionListHandlers)

	revisionListCommand := &cobra.Command{
		Use:   "list [name]",
		Short: "List available revisions.",
		Long:  "List revisions for a given service.",
		Example: `
  # List all revisions
  kn revision list

  # List revisions for a service 'svc1' in namespace 'myapp'
  kn revision list -s svc1 -n myapp

  # List all revisions in JSON output format
  kn revision list -o json

  # List revision 'web'
  kn revision list web`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			// Create list filters
			var params []clientservingv1alpha.ListConfig
			params, err = appendServiceFilter(params, client, cmd)
			if err != nil {
				return err
			}
			params, err = appendRevisionNameFilter(params, client, args)
			if err != nil {
				return err
			}

			// Query for list with filters
			revisionList, err := client.ListRevisions(params...)
			if err != nil {
				return err
			}

			// Stop if nothing found
			if len(revisionList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No revisions found.\n")
				return nil
			}

			// Add namespace column if no namespace is given (i.e. "--all-namespaces" option is given)
			if namespace == "" {
				revisionListFlags.EnsureWithNamespace()
			}

			// Only add temporary annotations if human readable output is requested
			if !revisionListFlags.GenericPrintFlags.OutputFlagSpecified() {
				err = enrichRevisionAnnotationsWithServiceData(p.NewServingClient, revisionList)
				if err != nil {
					return err
				}
			}

			// Sort revisions by namespace, service, generation (in this order)
			sortRevisions(revisionList)

			// Print out infos via printer framework
			printer, err := revisionListFlags.ToPrinter()
			if err != nil {
				return err
			}

			return printer.PrintObj(revisionList, cmd.OutOrStdout())
		},
	}
	commands.AddNamespaceFlags(revisionListCommand.Flags(), true)
	revisionListFlags.AddFlags(revisionListCommand)
	revisionListCommand.Flags().StringVarP(&serviceNameFilter, "service", "s", "", "Service name")

	return revisionListCommand
}

// If a service option is given append a filter to the list of filters
func appendServiceFilter(lConfig []clientservingv1alpha.ListConfig, client clientservingv1alpha.KnServingClient, cmd *cobra.Command) ([]clientservingv1alpha.ListConfig, error) {
	if !cmd.Flags().Changed("service") {
		return lConfig, nil
	}

	serviceName := cmd.Flag("service").Value.String()

	// Verify that service exists first
	_, err := client.GetService(serviceName)
	if err != nil {
		return nil, err
	}
	return append(lConfig, clientservingv1alpha.WithService(serviceName)), nil
}

// If an additional name is given append this as a revision name filter to the given list
func appendRevisionNameFilter(lConfigs []clientservingv1alpha.ListConfig, client clientservingv1alpha.KnServingClient, args []string) ([]clientservingv1alpha.ListConfig, error) {

	switch len(args) {
	case 0:
		// No revision name given
		return lConfigs, nil
	case 1:
		// Exactly one name given
		return append(lConfigs, clientservingv1alpha.WithName(args[0])), nil
	default:
		return nil, fmt.Errorf("'kn revision list' accepts maximum 1 argument, not %d arguments as given", len(args))
	}
}

// sortRevisions sorts revisions by namespace, service, generation and name (in this order)
func sortRevisions(revisionList *servingv1.RevisionList) {
	// sort revisionList by configuration generation key
	sort.SliceStable(revisionList.Items, revisionListSortFunc(revisionList))
}

// revisionListSortFunc sorts by namespace, service,  generation and name
func revisionListSortFunc(revisionList *servingv1.RevisionList) func(i int, j int) bool {
	return func(i, j int) bool {
		a := revisionList.Items[i]
		b := revisionList.Items[j]

		// By Namespace
		aNamespace := a.Namespace
		bNamespace := b.Namespace
		if aNamespace != bNamespace {
			return aNamespace < bNamespace
		}

		// By Service
		aService := a.Labels[serving.ServiceLabelKey]
		bService := b.Labels[serving.ServiceLabelKey]

		if aService != bService {
			return aService < bService
		}

		// By Generation
		// Convert configuration generation key from string to int for avoiding string comparison.
		agen, err := strconv.Atoi(a.Labels[serving.ConfigurationGenerationLabelKey])
		if err != nil {
			return a.Name < b.Name
		}
		bgen, err := strconv.Atoi(b.Labels[serving.ConfigurationGenerationLabelKey])
		if err != nil {
			return a.Name < b.Name
		}

		if agen != bgen {
			return agen > bgen
		}
		return a.Name < b.Name
	}
}

// Service factory function for a namespace
type serviceFactoryFunc func(namespace string) (clientservingv1alpha.KnServingClient, error)

// A function which looks up a service by name
type serviceGetFunc func(namespace, serviceName string) (*servingv1.Service, error)

// Create revision info with traffic and tag information (if present)
func enrichRevisionAnnotationsWithServiceData(serviceFactory serviceFactoryFunc, revisionList *servingv1.RevisionList) error {
	serviceLookup := serviceLookup(serviceFactory)

	for _, revision := range revisionList.Items {
		serviceName := revision.Labels[serving.ServiceLabelKey]
		if serviceName == "" {
			continue
		}
		service, err := serviceLookup(revision.Namespace, serviceName)
		if err != nil {
			return err
		}

		traffic, tags := trafficAndTagsForRevision(revision.Name, service)
		if traffic != 0 {
			revision.Annotations[RevisionTrafficAnnotation] = fmt.Sprintf("%d%%", traffic)
		}
		if len(tags) > 0 {
			revision.Annotations[RevisionTagsAnnotation] = strings.Join(tags, ",")
		}
	}
	return nil

}

// Create a function for being able to lookup a service for an arbitrary namespace
func serviceLookup(serviceFactory serviceFactoryFunc) serviceGetFunc {

	// Two caches: For service & clients (clients might not be necessary though)
	serviceCache := make(map[string]*servingv1.Service)
	clientCache := make(map[string]clientservingv1alpha.KnServingClient)

	return func(namespace, serviceName string) (*servingv1.Service, error) {
		if service, exists := serviceCache[serviceName]; exists {
			return service, nil
		}

		client := clientCache[namespace]
		if client == nil {
			var err error
			client, err = serviceFactory(namespace)
			if err != nil {
				return nil, err
			}
			clientCache[namespace] = client
		}

		service, err := client.GetService(serviceName)
		if err != nil {
			return nil, err
		}
		serviceCache[serviceName] = service
		return service, nil
	}
}
