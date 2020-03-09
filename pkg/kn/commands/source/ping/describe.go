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
	"fmt"

	"github.com/spf13/cobra"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

// NewPingDescribeCommand returns a new command for describe a Ping source object
func NewPingDescribeCommand(p *commands.KnParams) *cobra.Command {

	pingDescribe := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a Ping source",
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

			cjSource, err := pingSourceClient.GetPingSource(name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writePingSource(dw, cjSource, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Revisions summary info
			writeSink(dw, &cjSource.Spec.Sink)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, cjSource.Status.Conditions, printDetails)
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

func writeSink(dw printers.PrefixWriter, sink *duckv1.Destination) {
	subWriter := dw.WriteAttribute("Sink", "")
	subWriter.WriteAttribute("Name", sink.Ref.Name)
	subWriter.WriteAttribute("Namespace", sink.Ref.Namespace)
	ref := sink.Ref
	if ref != nil {
		subWriter.WriteAttribute("Resource", fmt.Sprintf("%s (%s)", sink.Ref.Kind, sink.Ref.APIVersion))
	}
	uri := sink.URI
	if uri != nil {
		subWriter.WriteAttribute("URI", uri.String())
	}
}

func writePingSource(dw printers.PrefixWriter, source *v1alpha2.PingSource, printDetails bool) {
	commands.WriteMetadata(dw, &source.ObjectMeta, printDetails)
	dw.WriteAttribute("Schedule", source.Spec.Schedule)
	dw.WriteAttribute("Data", source.Spec.JsonData)
}
