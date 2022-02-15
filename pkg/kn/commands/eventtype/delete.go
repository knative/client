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
	"knative.dev/client/pkg/kn/commands"
)

var deleteExample = `
  # Delete eventtype 'myeventtype' in the current namespace
  kn eventtype delete myeventtype

  # Delete eventtype 'myeventtype' in the 'myproject' namespace
  kn eventtype delete myeventtype --namespace myproject
`

// NewEventtypeDeleteCommand represents command to describe the details of an eventtype instance
func NewEventtypeDeleteCommand(p *commands.KnParams) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete eventtype",
		Example: deleteExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'eventtype delete' requires the eventtype name given as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			eventingV1Beta1Client, err := p.NewEventingV1beta1Client(namespace)
			if err != nil {
				return err
			}
			err = eventingV1Beta1Client.DeleteEventtype(cmd.Context(), name)
			if err != nil {
				return fmt.Errorf(
					"cannot delete eventtype '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Eventtype '%s' successfully deleted in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	return cmd
}
