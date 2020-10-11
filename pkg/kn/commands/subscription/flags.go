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

package subscription

import (
	"sort"

	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	hprinters "knative.dev/client/pkg/printers"

	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
)

// ListHandlers handles printing human readable table for `kn subscription list` command's output
func ListHandlers(h hprinters.PrintHandler) {
	subscriptionColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the subscription", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the subscription", Priority: 1},
		{Name: "Channel", Type: "string", Description: "Channel of the subscription", Priority: 1},
		{Name: "Subscriber", Type: "string", Description: "Subscriber sink of the subscription", Priority: 1},
		{Name: "Reply", Type: "string", Description: "Reply sink of the subscription", Priority: 1},
		{Name: "Dead Letter Sink", Type: "string", Description: "DeadLetterSink of the subscription", Priority: 1},
		{Name: "Ready", Type: "string", Description: "Ready state of the subscription", Priority: 1},
		{Name: "Reason", Type: "string", Description: "Reason for non ready subscription", Priority: 1},
	}
	h.TableHandler(subscriptionColumnDefinitions, printSubscription)
	h.TableHandler(subscriptionColumnDefinitions, printSubscriptionList)
}

// printSubscription populates a single row of Subscription list
func printSubscription(subscription *messagingv1beta1.Subscription, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: subscription},
	}

	name := subscription.Name
	ctype := subscription.Spec.Channel.Kind
	channel := subscription.Spec.Channel.Name

	var subscriber, reply, dls string
	if subscription.Spec.Subscriber != nil {
		subscriber = flags.SinkToString(*subscription.Spec.Subscriber)
	} else {
		subscriber = ""
	}
	if subscription.Spec.Reply != nil {
		reply = flags.SinkToString(*subscription.Spec.Reply)
	} else {
		reply = ""
	}
	if subscription.Spec.Delivery != nil && subscription.Spec.Delivery.DeadLetterSink != nil {
		dls = flags.SinkToString(*subscription.Spec.Delivery.DeadLetterSink)
	} else {
		dls = ""
	}
	ready := commands.ReadyCondition(subscription.Status.Conditions)
	reason := commands.NonReadyConditionReason(subscription.Status.Conditions)

	if options.AllNamespaces {
		row.Cells = append(row.Cells, subscription.Namespace)
	}

	row.Cells = append(row.Cells, name, ctype+":"+channel, subscriber, reply, dls, ready, reason)
	return []metav1beta1.TableRow{row}, nil
}

// printSubscriptionList populates the Subscription list table rows
func printSubscriptionList(subscriptionList *messagingv1beta1.SubscriptionList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	if options.AllNamespaces {
		return printSubscriptionListWithNamespace(subscriptionList, options)
	}

	rows := make([]metav1beta1.TableRow, 0, len(subscriptionList.Items))

	sort.SliceStable(subscriptionList.Items, func(i, j int) bool {
		return subscriptionList.Items[i].GetName() < subscriptionList.Items[j].GetName()
	})

	for _, item := range subscriptionList.Items {
		row, err := printSubscription(&item, options)
		if err != nil {
			return nil, err
		}

		rows = append(rows, row...)
	}
	return rows, nil
}

// printSubscriptionListWithNamespace populates the knative service table rows with namespace column
func printSubscriptionListWithNamespace(subscriptionList *messagingv1beta1.SubscriptionList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(subscriptionList.Items))

	// temporary slice for sorting services in non-default namespace
	var others []metav1beta1.TableRow

	for _, subscription := range subscriptionList.Items {
		// Fill in with services in `default` namespace at first
		if subscription.Namespace == "default" {
			r, err := printSubscription(&subscription, options)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
			continue
		}
		// put other services in temporary slice
		r, err := printSubscription(&subscription, options)
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
