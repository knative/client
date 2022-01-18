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
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"

	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

var describeExample = `
  # Describe a channel 'pipe'
  kn channel describe pipe

  # Print only channel URL
  kn channel describe pipe -o url`

// NewChannelDescribeCommand returns a new command for describe a channel object
func NewChannelDescribeCommand(p *commands.KnParams) *cobra.Command {

	// For machine readable output
	machineReadablePrintFlags := genericclioptions.NewPrintFlags("")

	cmd := &cobra.Command{
		Use:               "describe NAME",
		Short:             "Show details of a channel",
		Example:           describeExample,
		ValidArgsFunction: commands.ResourceNameCompletionFunc(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn channel describe' requires the channel name given as single argument")
			}
			name := args[0]

			client, err := newChannelClient(p, cmd)
			if err != nil {
				return err
			}

			channel, err := client.GetChannel(cmd.Context(), name)
			if err != nil {
				return knerrors.GetError(err)
			}

			out := cmd.OutOrStdout()

			if machineReadablePrintFlags.OutputFlagSpecified() {
				if strings.ToLower(*machineReadablePrintFlags.OutputFormat) == "url" {
					fmt.Fprintf(out, "%s\n", extractURL(channel))
					return nil
				}
				printer, err := machineReadablePrintFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(channel, out)
			}

			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeChannel(dw, channel, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, channel.Status.Conditions, printDetails)
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
	cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(append(machineReadablePrintFlags.AllowedFormats(), "url"), "|"))
	return cmd
}

func writeChannel(dw printers.PrefixWriter, channel *messagingv1.Channel, printDetails bool) {
	commands.WriteMetadata(dw, &channel.ObjectMeta, printDetails)
	ctype := fmt.Sprintf("%s (%s)", channel.Spec.ChannelTemplate.Kind, channel.Spec.ChannelTemplate.APIVersion)
	dw.WriteAttribute("Type", ctype)
	if channel.Status.Address != nil {
		dw.WriteAttribute("URL", extractURL(channel))
	}
}

func extractURL(channel *messagingv1.Channel) string {
	return channel.Status.Address.URL.String()
}
