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

package ping

import (
	"errors"
	"sort"

	"github.com/spf13/cobra"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"

	"knative.dev/client/lib/printing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

// NewPingDescribeCommand returns a new command for describe a Ping source object
func NewPingDescribeCommand(p *commands.KnParams) *cobra.Command {

	pingDescribe := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a ping source",
		Example: `
  # Describe a Ping source with name 'myping'
  kn source ping describe myping`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source ping describe' requires name of the source as single argument")
			}
			name := args[0]

			pingSourceClient, err := newPingSourceClient(p, cmd)
			if err != nil {
				return err
			}

			pingSource, err := pingSourceClient.GetPingSource(name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writePingSource(dw, pingSource, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Revisions summary info
			printing.DescribeSink(dw, "Sink", pingSource.Namespace, &pingSource.Spec.Sink)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			if pingSource.Spec.CloudEventOverrides != nil && pingSource.Spec.CloudEventOverrides.Extensions != nil {
				writeCeOverrides(dw, pingSource.Spec.CloudEventOverrides.Extensions)
				dw.WriteLine()
				if err := dw.Flush(); err != nil {
					return err
				}
			}

			// Condition info
			commands.WriteConditions(dw, pingSource.Status.Conditions, printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := pingDescribe.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")

	return pingDescribe
}

func writePingSource(dw printers.PrefixWriter, source *v1alpha2.PingSource, printDetails bool) {
	commands.WriteMetadata(dw, &source.ObjectMeta, printDetails)
	dw.WriteAttribute("Schedule", source.Spec.Schedule)
	dw.WriteAttribute("Data", source.Spec.JsonData)
}

func writeCeOverrides(dw printers.PrefixWriter, ceOverrides map[string]string) {
	subDw := dw.WriteAttribute("CloudEvent Overrides", "")
	keys := make([]string, 0, len(ceOverrides))
	for k := range ceOverrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		subDw.WriteAttribute(k, ceOverrides[k])
	}
}
