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

package binding

import (
	"errors"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/tracker"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

// NewBindingDescribeCommand returns a new command for describe a sink binding object
func NewBindingDescribeCommand(p *commands.KnParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a sink binding",
		Example: `
  # Describe a sink binding with name 'mysinkbinding'
  kn source binding describe mysinkbinding`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source binding describe' requires name of the sink binding as single argument")
			}
			name := args[0]

			bindingClient, err := newSinkBindingClient(p, cmd)
			if err != nil {
				return err
			}

			binding, err := bindingClient.GetSinkBinding(name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeSinkBinding(dw, binding, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, binding.Status.Conditions, printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := cmd.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")

	return cmd
}

func writeSinkBinding(dw printers.PrefixWriter, binding *v1alpha2.SinkBinding, printDetails bool) {
	commands.WriteMetadata(dw, &binding.ObjectMeta, printDetails)
	writeSubject(dw, binding.Namespace, &binding.Spec.Subject)
	writeSink(dw, binding.Namespace, &binding.Spec.Sink)
	if binding.Spec.CloudEventOverrides != nil && binding.Spec.CloudEventOverrides.Extensions != nil {
		writeCeOverrides(dw, binding.Spec.CloudEventOverrides.Extensions)
	}
}

func writeSink(dw printers.PrefixWriter, namespace string, sink *duckv1.Destination) {
	subWriter := dw.WriteAttribute("Sink", "")
	if sink.Ref.Namespace != "" && sink.Ref.Namespace != namespace {
		subWriter.WriteAttribute("Namespace", sink.Ref.Namespace)
	}
	subWriter.WriteAttribute("Name", sink.Ref.Name)
	ref := sink.Ref
	if ref != nil {
		subWriter.WriteAttribute("Resource", fmt.Sprintf("%s (%s)", sink.Ref.Kind, sink.Ref.APIVersion))
	}
	uri := sink.URI
	if uri != nil {
		subWriter.WriteAttribute("URI", uri.String())
	}
}

func writeCeOverrides(dw printers.PrefixWriter, ceOverrides map[string]string) {
	subDw := dw.WriteAttribute("CloudEvent Overrides", "")
	var keys []string
	for k := range ceOverrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		subDw.WriteAttribute(k, ceOverrides[k])
	}
}

func writeSubject(dw printers.PrefixWriter, namespace string, subject *tracker.Reference) {
	subjectDw := dw.WriteAttribute("Subject", "")
	if subject.Namespace != "" && subject.Namespace != namespace {
		subjectDw.WriteAttribute("Namespace", subject.Namespace)
	}
	if subject.Name != "" {
		subjectDw.WriteAttribute("Name", subject.Name)
	}
	subjectDw.WriteAttribute("Resource", fmt.Sprintf("%s (%s)", subject.Kind, subject.APIVersion))
	if subject.Selector != nil {
		matchDw := subjectDw.WriteAttribute("Selector", "")
		selector := subject.Selector
		if len(selector.MatchLabels) > 0 {
			var lKeys []string
			for k := range selector.MatchLabels {
				lKeys = append(lKeys, k)
			}
			sort.Strings(lKeys)
			for _, k := range lKeys {
				matchDw.WriteAttribute(k, selector.MatchLabels[k])
			}
		}
		// TODO: Print out selector.MatchExpressions
	}
}
