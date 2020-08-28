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

package channel

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
	hprinters "knative.dev/client/pkg/printers"

	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
)

var channelName = "channels.messaging.knative.dev"

// ListHandlers handles printing human readable table for `kn channel list` command's output
func ListHandlers(h hprinters.PrintHandler) {
	channelColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Channel", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Channel", Priority: 1},
		{Name: "Type", Type: "string", Description: "Type of the Channel", Priority: 1},
		{Name: "URL", Type: "string", Description: "URL of the Channel", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the Channel", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the Channel", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason for non ready channel", Priority: 1},
	}
	h.TableHandler(channelColumnDefinitions, printChannel)
	h.TableHandler(channelColumnDefinitions, printChannelList)
}

// printChannel populates a single row of Channel list
func printChannel(channel *messagingv1beta1.Channel, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: channel},
	}

	name := channel.Name
	ctype := channel.Spec.ChannelTemplate.Kind
	url := ""
	if channel.Status.Address != nil {
		url = channel.Status.Address.URL.String()
	}
	age := commands.TranslateTimestampSince(channel.CreationTimestamp)
	ready := commands.ReadyCondition(channel.Status.Conditions)
	reason := commands.NonReadyConditionReason(channel.Status.Conditions)

	if options.AllNamespaces {
		row.Cells = append(row.Cells, channel.Namespace)
	}

	row.Cells = append(row.Cells, name, ctype, url, age, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

// printChannelList populates the Channel list table rows
func printChannelList(channelList *messagingv1beta1.ChannelList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	if options.AllNamespaces {
		return printChannelListWithNamespace(channelList, options)
	}

	rows := make([]metav1beta1.TableRow, 0, len(channelList.Items))

	sort.SliceStable(channelList.Items, func(i, j int) bool {
		return channelList.Items[i].GetName() < channelList.Items[j].GetName()
	})

	for _, item := range channelList.Items {
		row, err := printChannel(&item, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}

// printChannelListWithNamespace populates the knative service table rows with namespace column
func printChannelListWithNamespace(channelList *messagingv1beta1.ChannelList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(channelList.Items))

	// temporary slice for sorting services in non-default namespace
	others := make([]metav1beta1.TableRow, 0, len(rows))

	for _, channel := range channelList.Items {
		// Fill in with services in `default` namespace at first
		if channel.Namespace == "default" {
			r, err := printChannel(&channel, options)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
			continue
		}
		// put other services in temporary slice
		r, err := printChannel(&channel, options)
		if err != nil {
			return nil, err
		}
		others = append(others, r...)
	}

	// sort other services list alphabetically by namespace
	sort.SliceStable(others, func(i, j int) bool {
		return others[i].Cells[0].(string) < others[j].Cells[0].(string)
	})

	return append(rows, others...), nil
}

// ListTypesHandlers handles printing human readable table for `kn channel list-types`
func ListTypesHandlers(h printers.PrintHandler) {
	channelTypesColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Type", Type: "string", Description: "Kind / Type of the channel type", Priority: 1},
		{Name: "Group", Type: "string", Description: "Group of the channel type", Priority: 1},
		{Name: "Version", Type: "string", Description: "Version of the channel type", Priority: 1},
	}
	h.TableHandler(channelTypesColumnDefinitions, printChannelTypes)
	h.TableHandler(channelTypesColumnDefinitions, printChannelTypesList)
}

// printChannelTypes populates a single row of source types list table
func printChannelTypes(channelType unstructured.Unstructured, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := channelType.GetName()
	content := channelType.UnstructuredContent()
	version := channelType.GroupVersionKind().Version
	kind, found, err := unstructured.NestedString(content, "spec", "names", "kind")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("can't find specs.names.kind for %s", name)
	}

	group, found, err := unstructured.NestedString(content, "spec", "group")
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("can't find specs.group for %s", name)
	}

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: &channelType},
	}
	row.Cells = append(row.Cells, kind, group, version)

	return []metav1beta1.TableRow{row}, nil
}

// printChannelTypesList populates the source types list table rows
func printChannelTypesList(channelTypesList *unstructured.UnstructuredList, options printers.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(channelTypesList.Items))

	sort.SliceStable(channelTypesList.Items, func(i, j int) bool {
		return channelTypesList.Items[i].GetName() < channelTypesList.Items[j].GetName()
	})
	for _, item := range channelTypesList.Items {
		if item.GetName() != channelName {
			row, err := printChannelTypes(item, options)
			if err != nil {
				return nil, err
			}
			rows = append(rows, row...)
		}

	}
	return rows, nil
}
