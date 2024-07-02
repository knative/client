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

package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

var waitExample = `
  # Waits on a service 'svc'
  kn service wait svc

  # Waits on a service 'svc' with a timeout
  kn service wait svc --wait-timeout 10

  # Waits on a service 'svc' with a timeout and wait window
  kn service wait svc --wait-timeout 10 --wait-window 1`

// NewServiceWaitCommand represents 'kn service wait' command
func NewServiceWaitCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags
	command := &cobra.Command{
		Use:     "wait NAME",
		Short:   "Wait for a service to be ready",
		Example: waitExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'service wait' requires the service name given as single argument")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := newServingClient(p, namespace, "")
			if err != nil {
				return err
			}

			name := args[0]
			out := cmd.OutOrStdout()

			fmt.Fprintf(out, "Waiting for Service '%s' in namespace '%s':\n", args[0], namespace)
			fmt.Fprintln(out, "")
			wconfig := clientservingv1.WaitConfig{
				Timeout:     time.Duration(waitFlags.TimeoutInSeconds) * time.Second,
				ErrorWindow: time.Duration(waitFlags.ErrorWindowInSeconds) * time.Second,
			}
			err = waitForService(cmd.Context(), client, name, out, wconfig)
			if err != nil {
				return err
			}
			fmt.Fprintln(out, "")
			fmt.Fprintf(out, "Service '%s' in namespace '%s' is ready.\n", name, namespace)

			return nil

		},
	}
	commands.AddNamespaceFlags(command.Flags(), false)
	waitFlags.AddConditionWaitFlags(command, commands.WaitDefaultTimeout, "wait", "service", "ready")
	return command
}
