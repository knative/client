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

package source

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/client/pkg/dynamic"
	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	sourcesv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"
)

// NewListTypesCommand defines and processes `kn source list-types`
func NewListTypesCommand(p *commands.KnParams) *cobra.Command {
	listTypesFlags := flags.NewListPrintFlags(ListTypesHandlers)
	listTypesCommand := &cobra.Command{
		Use:   "list-types",
		Short: "List event source types",
		Example: `
  # List available event source types
  kn source list-types

  # List available event source types in YAML format
  kn source list-types -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			sourceListTypes, err := dynamicClient.ListSourcesTypes()
			if err != nil {
				if strings.HasPrefix(knerrors.GetError(err).Error(), "403") {
					sourcesClient, err := p.NewSourcesClient(namespace)
					if err != nil {
						return err
					}

					sourceListTypes, err = listBuiltInSourceTypes(sourcesClient)
					if err != nil {
						return knerrors.GetError(err)
					}
				}
			}

			if len(sourceListTypes.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "404: no sources found on the backend, please verify the installation\n")
				return nil
			}

			printer, err := listTypesFlags.ToPrinter()
			if err != nil {
				return nil
			}

			err = printer.PrintObj(sourceListTypes, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			return nil
		},
	}
	commands.AddNamespaceFlags(listTypesCommand.Flags(), false)
	listTypesFlags.AddFlags(listTypesCommand)
	return listTypesCommand
}

func listBuiltInSourceTypes(c sourcesv1alpha2.KnSourcesClient) (*unstructured.UnstructuredList, error) {
	_, err := c.APIServerSourcesClient().ListAPIServerSource()
	if err != nil {
		if strings.HasPrefix(err.Error(), "404") {
			return nil, err
		}
	}

	uList := unstructured.UnstructuredList{}
	gvks := sourcesv1alpha2.BuiltInSourcesGVKs()
	for _, gvk := range gvks {
		u := dynamic.UnstructuredCRDFromGVK(gvk)
		uList.Items = append(uList.Items, *u)
	}
	return &uList, nil
}
