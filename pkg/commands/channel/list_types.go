/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package channel

import (
	"context"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	listfl "knative.dev/client-pkg/pkg/commands/flags/list"
	"knative.dev/client-pkg/pkg/dynamic"
	knerrors "knative.dev/client-pkg/pkg/errors"
	messagingv1 "knative.dev/client-pkg/pkg/messaging/v1"
	"knative.dev/client/pkg/commands"
)

// NewChannelListTypesCommand defines and processes `kn channel list-types`
func NewChannelListTypesCommand(p *commands.KnParams) *cobra.Command {
	listTypesFlags := listfl.NewPrintFlags(ListTypesHandlers)
	listTypesCommand := &cobra.Command{
		Use:   "list-types",
		Short: "List channel types",
		Example: `
  # List available channel types
  kn channel list-types

  # List available channel types in YAML format
  kn channel list-types -o yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}

			channelListTypes, err := dynamicClient.ListChannelsTypes(cmd.Context())
			switch {
			case knerrors.IsForbiddenError(err):
				if channelListTypes, err = listBuiltInChannelTypes(cmd.Context(), dynamicClient); err != nil {
					return knerrors.GetError(err)
				}
			case err != nil:
				return knerrors.GetError(err)
			}

			if channelListTypes == nil {
				channelListTypes = &unstructured.UnstructuredList{}
			}
			if !listTypesFlags.GenericPrintFlags.OutputFlagSpecified() && len(channelListTypes.Items) == 0 {
				return knerrors.NewInvalidCRD("Channels")
			}

			if channelListTypes.GroupVersionKind().Empty() {
				channelListTypes.SetAPIVersion("apiextensions.k8s.io/v1")
				channelListTypes.SetKind("CustomResourceDefinitionList")
			}

			printer, err := listTypesFlags.ToPrinter()
			if err != nil {
				return nil
			}

			err = printer.PrintObj(channelListTypes, cmd.OutOrStdout())
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

func listBuiltInChannelTypes(ctx context.Context, d dynamic.KnDynamicClient) (*unstructured.UnstructuredList, error) {
	var err error
	uList := unstructured.UnstructuredList{}
	gvks := messagingv1.BuiltInChannelGVKs()
	for _, gvk := range gvks {
		_, err = d.ListChannelsUsingGVKs(ctx, &[]schema.GroupVersionKind{gvk})
		if err != nil {
			continue
		}
		u := dynamic.UnstructuredCRDFromGVK(gvk)
		uList.Items = append(uList.Items, *u)
	}
	// if not even one channel is found
	if len(uList.Items) == 0 && err != nil {
		return nil, knerrors.GetError(err)
	}
	return &uList, nil
}
