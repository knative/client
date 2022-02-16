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
	clienteventingv1beta1 "knative.dev/client/pkg/eventing/v1beta1"
	"knative.dev/client/pkg/kn/commands"
	knflags "knative.dev/client/pkg/kn/commands/flags"
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
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create eventtype",
		Example: createExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'eventtype create' requires the eventtype name given as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return eventtypeCreateError(name, namespace, err)
			}

			eventingV1Beta1Client, err := p.NewEventingV1beta1Client(namespace)
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
			eventtype := clienteventingv1beta1.NewEventtypeBuilder(name).
				Namespace(namespace).
				Type(eventtypeFlags.Type).
				Source(source).
				Broker(eventtypeFlags.Broker).
				Build()

			err = eventingV1Beta1Client.CreateEventtype(cmd.Context(), eventtype)
			if err != nil {
				return eventtypeCreateError(name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Eventtype '%s' successfully created in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)

	eventtypeFlags.Add(cmd)
	return cmd
}

func eventtypeCreateError(name string, namespace string, err error) error {
	return fmt.Errorf(
		"cannot create eventtype '%s' in namespace '%s' "+
			"because: %s", name, namespace, err)
}
