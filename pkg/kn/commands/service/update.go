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

  # Updates a service 'svc' with new requests and limits parameters
  kn service update svc --requests-cpu 500m --limits-memory 1024Mi

  # Assign tag 'latest' and 'stable' to revisions 'echo-v2' and 'echo-v1' respectively
  kn service update svc --tag echo-v2=latest --tag echo-v1=stable
  OR
  kn service update svc --tag echo-v2=latest,echo-v1=stable

  # Update tag from 'testing' to 'staging' for latest ready revision of service
  kn service update svc --untag testing --tag @latest=staging

  # Add tag 'test' to echo-v3 revision with 10% traffic and rest to latest ready revision of service
  kn service update svc --tag echo-v3=test --traffic test=10,@latest=90`

func NewServiceUpdateCommand(p *commands.KnParams) *cobra.Command {
	var editFlags ConfigurationEditFlags
	var waitFlags commands.WaitFlags
	var trafficFlags flags.Traffic
	serviceUpdateCommand := &cobra.Command{
		Use:     "update NAME [flags]",
		Short:   "Update a service.",
		Example: updateExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the service name")
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewServingClient(namespace)
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
					baseRevision, err = client.GetBaseRevision(service)
					if _, ok := err.(*clientservingv1.NoBaseRevisionError); ok {
						fmt.Fprintf(cmd.OutOrStdout(), "Warning: No revision found to update image digest")
					}
				}
				err = editFlags.Apply(service, baseRevision, cmd)
				if err != nil {
					return nil, err
				}

				if trafficFlags.Changed(cmd) {
					traffic, err := traffic.Compute(cmd, service.Spec.Traffic, &trafficFlags)
					if err != nil {
						return nil, err
					}

					service.Spec.Traffic = traffic
				}
				return service, nil
			}

			// Do the actual update with retry in case of conflicts
			err = client.UpdateServiceWithRetry(name, updateFunc, MaxUpdateRetries)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			//TODO: deprecated condition should be once --async is gone
			if !waitFlags.Async && !waitFlags.NoWait {
				fmt.Fprintf(out, "Updating Service '%s' in namespace '%s':\n", args[0], namespace)
				fmt.Fprintln(out, "")
				err := waitForService(client, name, out, waitFlags.TimeoutInSeconds)
				if err != nil {
					return err
				}
				fmt.Fprintln(out, "")
				return showUrl(client, name, latestRevisionBeforeUpdate, "updated", out)
			} else {
				if waitFlags.Async {
					fmt.Fprintf(out, "\nWARNING: flag --async is deprecated and going to be removed in future release, please use --no-wait instead.\n\n")
				}
				fmt.Fprintf(out, "Service '%s' updated in namespace '%s'.\n", args[0], namespace)
			}

			return nil

		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preCheck(cmd, args)
		},
	}

	commands.AddNamespaceFlags(serviceUpdateCommand.Flags(), false)
	editFlags.AddUpdateFlags(serviceUpdateCommand)
	waitFlags.AddConditionWaitFlags(serviceUpdateCommand, commands.WaitDefaultTimeout, "Update", "service", "ready")
	trafficFlags.Add(serviceUpdateCommand)
	return serviceUpdateCommand
}

func preCheck(cmd *cobra.Command, args []string) error {
	if cmd.Flags().NFlag() == 0 {
		return fmt.Errorf("flag(s) not set\nUsage: %s", cmd.Use)
	}

	return nil
}
