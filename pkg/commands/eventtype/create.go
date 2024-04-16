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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	sinkfl "knative.dev/client-pkg/pkg/commands/flags/sink"
	clienteventingv1beta2 "knative.dev/client-pkg/pkg/eventing/v1beta2"
	"knative.dev/client/pkg/commands"
	knflags "knative.dev/client/pkg/commands/flags"
	"knative.dev/pkg/apis"
)

var createExample = `
  # Create eventtype 'myeventtype' of type example.type in the current namespace
  kn eventtype create myeventtype --type example.type

  # Create eventtype 'myeventtype' of type example.type in the 'myproject' namespace
  kn eventtype create myeventtype --namespace myproject -t example.type
`

// NewEventtypeCreateCommand represents command to describe the details of an eventtype instance
func NewEventtypeCreateCommand(p *commands.KnParams) *cobra.Command {

	var eventtypeFlags knflags.EventtypeFlags

	referenceFlag := sinkfl.NewFlag(referenceMappings)

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create eventtype",
		Example: createExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'eventtype create' requires the eventtype name given as single argument")
			}
			name := args[0]

			if eventtypeFlags.Broker != "" && referenceFlag.Sink != "" {
				return errors.New("use only one of '--broker' or '--reference' flags")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return eventtypeCreateError(name, namespace, err)
			}

			eventingV1Beta2Client, err := p.NewEventingV1beta2Client(namespace)
			if err != nil {
				return err
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}
			var source *apis.URL

			if eventtypeFlags.Source != "" {
				source, err = apis.ParseURL(eventtypeFlags.Source)
				if err != nil {
					return eventtypeCreateError(name, namespace, err)
				}
			}
			etBuilder := clienteventingv1beta2.NewEventtypeBuilder(name).
				Namespace(namespace).
				Type(eventtypeFlags.Type).
				Source(source)

			if eventtypeFlags.Broker != "" {
				etBuilder.Broker(eventtypeFlags.Broker)
			}

			if referenceFlag.Sink != "" {
				dest, err := referenceFlag.ResolveSink(cmd.Context(), dynamicClient, namespace)
				if err != nil {
					return eventtypeCreateError(name, namespace, err)
				}
				etBuilder.Reference(dest.Ref)
			}
			err = eventingV1Beta2Client.CreateEventtype(cmd.Context(), etBuilder.Build())
			if err != nil {
				return eventtypeCreateError(name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Eventtype '%s' successfully created in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)

	referenceFlag.AddWithFlagName(cmd, "reference", "r")
	flag := "reference"
	cmd.Flag(flag).Usage = "Addressable Reference producing events. " +
		"You can specify a broker, channel, or fully qualified GroupVersionResource (GVR). " +
		"Examples: '--" + flag + " broker:nest' for a broker 'nest', " +
		"'--" + flag + " channel:pipe' for a channel 'pipe', " +
		"'--" + flag + " special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'."
	eventtypeFlags.Add(cmd)
	return cmd
}

func eventtypeCreateError(name string, namespace string, err error) error {
	return fmt.Errorf(
		"cannot create eventtype '%s' in namespace '%s' "+
			"because: %s", name, namespace, err)
}
