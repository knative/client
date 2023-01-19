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

package broker

import (
	"fmt"

	"knative.dev/pkg/apis"

	"github.com/spf13/cobra"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	hprinters "knative.dev/client/pkg/printers"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

var listExample = `
  # List all brokers
  kn broker list

  # List all brokers in JSON output format
  kn broker list -o json`

// NewBrokerListCommand represents command to list all brokers
func NewBrokerListCommand(p *commands.KnParams) *cobra.Command {
	brokerListFlags := flags.NewListPrintFlags(ListHandlers)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List brokers",
		Aliases: []string{"ls"},
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			eventingClient, err := p.NewEventingClient(namespace)
			if err != nil {
				return err
			}

			brokerList, err := eventingClient.ListBrokers(cmd.Context())
			if err != nil {
				return err
			}
			if !brokerListFlags.GenericPrintFlags.OutputFlagSpecified() && len(brokerList.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No brokers found.\n")
				return nil
			}

			// empty namespace indicates all-namespaces flag is specified
			if namespace == "" {
				brokerListFlags.EnsureWithNamespace()
			}

			err = brokerListFlags.Print(brokerList, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), true)
	brokerListFlags.AddFlags(cmd)
	return cmd
}

// ListHandlers handles printing human readable table for `kn broker list` command's output
func ListHandlers(h hprinters.PrintHandler) {
	brokerColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Broker instance", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Broker instance", Priority: 1},
		{Name: "URL", Type: "string", Description: "URL of the Broker instance", Priority: 1},
		{Name: "Age", Type: "string", Description: "Age of the Broker instance", Priority: 1},
		{Name: "Conditions", Type: "string", Description: "Ready state conditions", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the Broker instance", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason if state is not Ready", Priority: 1},
	}
	h.TableHandler(brokerColumnDefinitions, printBroker)
	h.TableHandler(brokerColumnDefinitions, printBrokerList)
}

// printBrokerList populates the broker list table rows
func printBrokerList(kServiceList *eventingv1.BrokerList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(kServiceList.Items))

	for i := range kServiceList.Items {
		ksvc := &kServiceList.Items[i]
		r, err := printBroker(ksvc, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

// printBroker populates the broker table rows
func printBroker(broker *eventingv1.Broker, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := broker.Name
	url := &apis.URL{}
	if broker.Status.AddressStatus.Address != nil {
		url = broker.Status.AddressStatus.Address.URL
	}
	age := commands.TranslateTimestampSince(broker.CreationTimestamp)
	conditions := commands.ConditionsValue(broker.Status.Conditions)
	ready := commands.ReadyCondition(broker.Status.Conditions)
	reason := commands.NonReadyConditionReason(broker.Status.Conditions)

	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: broker},
	}

	if options.AllNamespaces {
		row.Cells = append(row.Cells, broker.Namespace)
	}

	row.Cells = append(row.Cells,
		name,
		url,
		age,
		conditions,
		ready,
		reason)
	return []metav1beta1.TableRow{row}, nil
}
