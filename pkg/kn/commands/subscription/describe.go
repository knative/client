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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"

	"knative.dev/client/lib/printing"
	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

// NewSubscriptionDescribeCommand returns a new command for describe a subscription object
func NewSubscriptionDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	cmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a subscription",
		Example: `
  # Describe a subscription 'pipe'
  kn subscription describe pipe`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn subscription describe' requires the subscription name given as single argument")
			}
			name := args[0]

			client, err := newSubscriptionClient(p, cmd)
			if err != nil {
				return err
			}

			subscription, err := client.GetSubscription(name)
			if err != nil {
				return knerrors.GetError(err)
			}

			out := cmd.OutOrStdout()

			if machineReadablePrintFlags.OutputFlagSpecified() {
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(subscription, out)
			}

			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeSubscription(dw, subscription, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, subscription.Status.Conditions, printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := cmd.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")
	machineReadablePrintFlags.AddFlags(cmd)
	return cmd
}

func writeSubscription(dw printers.PrefixWriter, subscription *messagingv1beta1.Subscription, printDetails bool) {
	commands.WriteMetadata(dw, &subscription.ObjectMeta, printDetails)
	ctype := fmt.Sprintf("%s:%s (%s)", subscription.Spec.Channel.Kind, subscription.Spec.Channel.Name, subscription.Spec.Channel.APIVersion)
	dw.WriteAttribute("Channel", ctype)
	printing.DescribeSink(dw, "Subscriber", subscription.Namespace, subscription.Spec.Subscriber)
	printing.DescribeSink(dw, "Reply", subscription.Namespace, subscription.Spec.Reply)
	if subscription.Spec.DeepCopy().Delivery != nil {
		printing.DescribeSink(dw, "DeadLetterSink", subscription.Namespace, subscription.Spec.Delivery.DeadLetterSink)
	}
}
