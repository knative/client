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

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/kn/commands/flags"
	"knative.dev/client/pkg/kn/traffic"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

var updateExample = `
  # Updates a service 'svc' with new environment variables
  kn service update svc --env KEY1=VALUE1 --env KEY2=VALUE2

  # Update a service 'svc' with new port
  kn service update svc --port 80

  # Updates a service 'svc' with new request and limit parameters
  kn service update svc --request cpu=500m --limit memory=1024Mi --limit cpu=1000m

  # Assign tag 'latest' and 'stable' to revisions 'echo-v2' and 'echo-v1' respectively
  kn service update svc --tag echo-v2=latest --tag echo-v1=stable
  OR
  kn service update svc --tag echo-v2=latest,echo-v1=stable

  # Update tag from 'testing' to 'staging' for latest ready revision of service
  kn service update svc --untag testing --tag @latest=staging

  # Add tag 'test' to echo-v3 revision with 10% traffic and rest to latest ready revision of service
  kn service update svc --tag echo-v3=test --traffic test=10,@latest=90

  # Update the service in offline mode instead of kubernetes cluster
  kn service update gitopstest -n test-ns --env KEY1=VALUE1 --target=/user/knfiles
  kn service update gitopstest --env KEY1=VALUE1 --target=/user/knfiles/test.yaml
  kn service update gitopstest --env KEY1=VALUE1 --target=/user/knfiles/test.json`

func NewServiceUpdateCommand(p *commands.KnParams) *cobra.Command {
	var editFlags ConfigurationEditFlags
	var waitFlags commands.WaitFlags
	var trafficFlags flags.Traffic
	serviceUpdateCommand := &cobra.Command{
		Use:     "update NAME",
		Short:   "Update a service",
		Example: updateExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'service update' requires the service name given as single argument")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			targetFlag := cmd.Flag("target").Value.String()
			client, err := newServingClient(p, namespace, targetFlag)
			if err != nil {
				return err
			}

			// Use to store the latest revision name
			var latestRevisionBeforeUpdate string
			name := args[0]

			updateFunc := func(service *servingv1.Service) (*servingv1.Service, error) {
				latestRevisionBeforeUpdate = service.Status.LatestReadyRevisionName
				var baseRevision *servingv1.Revision
				if !cmd.Flags().Changed("image") && editFlags.LockToDigest {
					baseRevision, err = client.GetBaseRevision(cmd.Context(), service)
					var errNoBaseRevision clientservingv1.NoBaseRevisionError
					if errors.As(err, &errNoBaseRevision) {
						fmt.Fprintf(cmd.OutOrStdout(), "Warning: No revision found to update image digest")
					}
				}
				err = editFlags.Apply(service, baseRevision, cmd)
				if err != nil {
					return nil, err
				}

				if trafficFlags.Changed(cmd) {
					traffic, err := traffic.Compute(cmd, service.Spec.Traffic, &trafficFlags, service.Name)
					if err != nil {
						return nil, err
					}

					service.Spec.Traffic = traffic
				}
				return service, nil
			}

			// Do the actual update with retry in case of conflicts
			changed, err := client.UpdateServiceWithRetry(cmd.Context(), name, updateFunc, MaxUpdateRetries)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			// No need to wait if not changed
			if !changed {
				fmt.Fprintf(out, "Service '%s' updated in namespace '%s'.\n", args[0], namespace)
				fmt.Fprintln(out, "No new revision has been created.")
				return nil
			}

			if waitFlags.Wait && targetFlag == "" {
				fmt.Fprintf(out, "Updating Service '%s' in namespace '%s':\n", args[0], namespace)
				fmt.Fprintln(out, "")
				err := waitForService(cmd.Context(), client, name, out, waitFlags.TimeoutInSeconds)
				if err != nil {
					return err
				}
				fmt.Fprintln(out, "")
				return showUrl(cmd.Context(), client, name, latestRevisionBeforeUpdate, "updated", out)
			} else {
				fmt.Fprintf(out, "Service '%s' updated in namespace '%s'.\n", args[0], namespace)
			}

			return nil

		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preCheck(cmd)
		},
	}

	commands.AddNamespaceFlags(serviceUpdateCommand.Flags(), false)
	commands.AddGitOpsFlags(serviceUpdateCommand.Flags())
	editFlags.AddUpdateFlags(serviceUpdateCommand)
	waitFlags.AddConditionWaitFlags(serviceUpdateCommand, commands.WaitDefaultTimeout, "update", "service", "ready")
	trafficFlags.Add(serviceUpdateCommand)
	return serviceUpdateCommand
}

func preCheck(cmd *cobra.Command) error {
	if cmd.Flags().NFlag() == 0 {
		return fmt.Errorf("flag(s) not set\nUsage: %s", cmd.Use)
	}

	return nil
}
