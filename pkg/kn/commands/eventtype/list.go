/*
Copyright 2022 The Knative Authors

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

package eventtype

import (
	"fmt"

	"github.com/spf13/cobra"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	hprinters "knative.dev/client/pkg/printers"
)

var listExample = `
  # List all eventtypes
  kn eventtype list

  # List all eventtypes in JSON output format
  kn eventtype list -o json`

// NewEventtypeListCommand represents command to list all eventtypes
func NewEventtypeListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(ListHandlers)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List eventtypes",
		Aliases: []string{"ls"},
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			eventingV1Beta1Client, err := p.NewEventingV1beta1Client(namespace)
			if err != nil {
				return err
			}

			eventTypeList, err := eventingV1Beta1Client.ListEventtypes(cmd.Context())
			if err != nil {
				return err
			}
			if !listFlags.GenericPrintFlags.OutputFlagSpecified() && len(eventTypeList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No eventtypes found.\n")
				return nil
			}

			// empty namespace indicates all-namespaces flag is specified
			if namespace == "" {
				listFlags.EnsureWithNamespace()
			}

			err = listFlags.Print(eventTypeList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), true)
	listFlags.AddFlags(cmd)
	return cmd
}

// ListHandlers handles printing human readable table for `kn eventtype list` command's output
func ListHandlers(h hprinters.PrintHandler) {
	eventTypeColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the EventType instance", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the EventType instance", Priority: 1},
		{Name: "Type", Type: "string", Description: "Type of the EventType instance", Priority: 1},
		{Name: "Source", Type: "string", Description: "Source of the EventType instance", Priority: 1},
		{Name: "Broker", Type: "string", Description: "Broker of the EventType instance", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the EventType instance", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the EventType instance", Priority: 1},
	}
	h.TableHandler(eventTypeColumnDefinitions, printEventType)
	h.TableHandler(eventTypeColumnDefinitions, printEventTypeList)
}

// printEventTypeList populates the eventtype list table rows
func printEventTypeList(eventTypeList *eventingv1beta1.EventTypeList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(eventTypeList.Items))

	for i := range eventTypeList.Items {
		eventType := &eventTypeList.Items[i]
		r, err := printEventType(eventType, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printEventType populates the eventtype table rows
func printEventType(eventType *eventingv1beta1.EventType, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := eventType.Name
	age := commands.TranslateTimestampSince(eventType.CreationTimestamp)
	cetype := eventType.Spec.Type
	source := eventType.Spec.Source
	broker := eventType.Spec.Broker
	ready := commands.ReadyCondition(eventType.Status.Conditions)

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: eventType},
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, eventType.Namespace)
	}

	row.Cells = append(row.Cells, name, cetype, source, broker, age, ready)
	return []metav1beta1.TableRow{row}, nil
}
