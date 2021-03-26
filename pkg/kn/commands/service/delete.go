// Copyright © 2018 The Knative Authors
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
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

// NewServiceDeleteCommand represent 'service delete' command
func NewServiceDeleteCommand(p *commands.KnParams) *cobra.Command {
	var waitFlags commands.WaitFlags

	serviceDeleteCommand := &cobra.Command{
		Use:   "delete NAME [NAME ...]",
		Short: "Delete services",
		Example: `
  # Delete a service 'svc1' in default namespace
  kn service delete svc1

  # Delete a service 'svc2' in 'ns1' namespace
  kn service delete svc2 -n ns1

  # Delete all services in 'ns1' namespace
  kn service delete --all -n ns1

  # Delete the services in offline mode instead of kubernetes cluster
  kn service delete test -n test-ns --target=/user/knfiles
  kn service delete test --target=/user/knfiles/test.yaml
  kn service delete test --target=/user/knfiles/test.json`,

		RunE: func(cmd *cobra.Command, args []string) error {
			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}
			argsLen := len(args)

			if argsLen < 1 && !all {
				return errors.New("'service delete' requires the service name(s)")
			}

			if argsLen > 0 && all {
				return errors.New("'service delete' with --all flag requires no arguments")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := newServingClient(p, namespace, cmd.Flag("target").Value.String())
			if err != nil {
				return err
			}

			if all {
				args, err = getServiceNames(cmd.Context(), client)
				if err != nil {
					return err
				}
				if len(args) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "No services found.\n")
					return nil
				}
			}

			errs := []string{}
			for _, name := range args {
				timeout := time.Duration(0)
				if waitFlags.Wait {
					timeout = time.Duration(waitFlags.TimeoutInSeconds) * time.Second
				}
				err = client.DeleteService(cmd.Context(), name, timeout)
				if err != nil {
					errs = append(errs, err.Error())
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Service '%s' successfully deleted in namespace '%s'.\n", name, namespace)
				}
			}
			if len(errs) > 0 {
				return errors.New("Error: " + strings.Join(errs, "\nError: "))
			}
			return nil
		},
	}
	flags := serviceDeleteCommand.Flags()
	flags.Bool("all", false, "Delete all services in a namespace.")
	commands.AddNamespaceFlags(serviceDeleteCommand.Flags(), false)
	commands.AddGitOpsFlags(serviceDeleteCommand.Flags())
	waitFlags.AddConditionWaitFlags(serviceDeleteCommand, commands.WaitDefaultTimeout, "delete", "service", "deleted")
	return serviceDeleteCommand
}

func getServiceNames(ctx context.Context, client clientservingv1.KnServingClient) ([]string, error) {
	serviceList, err := client.ListServices(ctx)
	if err != nil {
		return []string{}, err
	}
	serviceNames := []string{}
	for _, service := range serviceList.Items {
		serviceNames = append(serviceNames, service.Name)
	}
	return serviceNames, nil
}
