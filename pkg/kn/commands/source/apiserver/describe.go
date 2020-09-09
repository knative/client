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

package apiserver

import (
	"errors"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
)

// NewAPIServerDescribeCommand to describe an ApiServer source object
func NewAPIServerDescribeCommand(p *commands.KnParams) *cobra.Command {
	apiServerDescribe := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of an api-server source",
		Example: `
  # Describe an ApiServer source with name 'k8sevents'
  kn source apiserver describe k8sevents`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source apiserver describe' requires name of the source as single argument")
			}
			name := args[0]

			apiSourceClient, err := newAPIServerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			apiSource, err := apiSourceClient.GetAPIServerSource(name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeAPIServerSource(dw, apiSource, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			writeSink(dw, apiSource.Spec.Sink)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			if apiSource.Spec.CloudEventOverrides != nil && apiSource.Spec.CloudEventOverrides.Extensions != nil {
				writeCeOverrides(dw, apiSource.Spec.CloudEventOverrides.Extensions)
			}

			writeResources(dw, apiSource.Spec.Resources)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, apiSource.Status.Conditions, printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := apiServerDescribe.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")

	return apiServerDescribe
}

func writeResources(dw printers.PrefixWriter, apiVersionKindSelectors []v1alpha2.APIVersionKindSelector) {
	subWriter := dw.WriteAttribute("Resources", "")
	for _, resource := range apiVersionKindSelectors {
		subWriter.WriteAttribute("Kind", fmt.Sprintf("%s (%s)", resource.Kind, resource.APIVersion))
		if resource.LabelSelector != nil {
			subWriter.WriteAttribute("Selector", labelSelectorToString(resource.LabelSelector))
		}
	}
}

func writeSink(dw printers.PrefixWriter, sink duckv1.Destination) {
	subWriter := dw.WriteAttribute("Sink", "")
	ref := sink.Ref
	if ref != nil {
		subWriter.WriteAttribute("Name", sink.Ref.Name)
		subWriter.WriteAttribute("Namespace", sink.Ref.Namespace)
		subWriter.WriteAttribute("Kind", fmt.Sprintf("%s (%s)", sink.Ref.Kind, sink.Ref.APIVersion))
	}
	uri := sink.URI
	if uri != nil {
		subWriter.WriteAttribute("URI", uri.String())
	}
}

func writeAPIServerSource(dw printers.PrefixWriter, source *v1alpha2.ApiServerSource, printDetails bool) {
	commands.WriteMetadata(dw, &source.ObjectMeta, printDetails)
	dw.WriteAttribute("ServiceAccountName", source.Spec.ServiceAccountName)
	dw.WriteAttribute("EventMode", source.Spec.EventMode)
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
