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

package container

import (
	context2 "context"
	"errors"
	"sort"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/client/lib/printing"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/printers"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
)

// NewContainerDescribeCommand to describe an Container source object
func NewContainerDescribeCommand(p *commands.KnParams) *cobra.Command {
	apiServerDescribe := &cobra.Command{
		Use:   "describe NAME",
		Short: "Show details of a container source",
		Example: `
  # Describe a container source with name 'k8sevents'
  kn source container describe k8sevents`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'kn source container describe' requires name of the source as single argument")
			}
			name := args[0]

			sourceClient, err := newContainerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			source, err := sourceClient.GetContainerSource(context2.TODO(), name)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeContainerSource(dw, source, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			printing.DescribeSink(dw, "Sink", source.Namespace, &source.Spec.Sink)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			if source.Spec.CloudEventOverrides != nil && source.Spec.CloudEventOverrides.Extensions != nil {
				writeCeOverrides(dw, source.Spec.CloudEventOverrides.Extensions)
			}

			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, source.Status.Conditions, printDetails)
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

func writeContainerSource(dw printers.PrefixWriter, source *v1alpha2.ContainerSource, printDetails bool) {
	commands.WriteMetadata(dw, &source.ObjectMeta, printDetails)
	writeContainer(dw, &source.Spec.Template.Spec.Containers[0])
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

func writeContainer(dw printers.PrefixWriter, container *corev1.Container) {
	subDw := dw.WriteAttribute("Container", "")
	subDw.WriteAttribute("Image", container.Image)
	if len(container.Env) > 0 {
		envDw := subDw.WriteAttribute("Env", "")
		for _, env := range container.Env {
			value := env.Value
			if env.ValueFrom != nil {
				value = "[ref]"
			}
			envDw.WriteAttribute(env.Name, value)
		}
	}
	if len(container.Args) > 0 {
		envDw := subDw.WriteAttribute("Args", "")
		for _, k := range container.Args {
			envDw.WriteAttribute(k, "")
		}
	}
}
