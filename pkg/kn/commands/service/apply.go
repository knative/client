// Copyright © 2020 The Knative Authors
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

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/commands"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
)

var applyExample = `
# Create an initial service with using 'kn service apply', if the service has not
# been already created
kn service apply s0 --image knativesamples/helloworld

# Apply the service again which is a no-operation if none of the options changed
kn service apply s0 --image knativesamples/helloworld

# Add an environment variable to your service. Note, that you have to always fully
# specify all parameters (in contrast to 'kn service update')
kn service apply s0 --image knativesamples/helloworld --env foo=bar

# Read the service declaration from a file
kn service apply s0 --filename my-svc.yml
`

func NewServiceApplyCommand(p *commands.KnParams) *cobra.Command {
	var applyFlags ConfigurationEditFlags
	var waitFlags commands.WaitFlags

	serviceApplyCommand := &cobra.Command{
		Use:     "apply NAME",
		Short:   "Apply a service declaration",
		Example: applyExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 && applyFlags.Filename == "" {
				return errors.New("'service apply' requires the service name given as single argument")
			}
			name := ""
			if len(args) == 1 {
				name = args[0]
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			var service *servingv1.Service
			applyFlags.RevisionName = ""
			if applyFlags.Filename == "" {
				service, err = constructService(cmd, applyFlags, name, namespace)
			} else {
				service, err = constructServiceFromFile(cmd, applyFlags, name, namespace)
			}
			if err != nil {
				return err
			}

			client, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			waitDoing, waitVerb, err := examineServiceForApply(cmd, client, service.Name)
			if err != nil {
				return err
			}

			hasChanged, err := client.ApplyService(context.TODO(), service)
			if err != nil {
				return err
			}
			if !hasChanged {
				fmt.Fprintf(cmd.OutOrStdout(), "No changes to apply to service '%s'.\n", service.Name)

				return showUrl(client, service.Name, "unchanged", "", cmd.OutOrStdout())
			}
			return waitIfRequested(client, waitFlags, service.Name, waitDoing, waitVerb, "", cmd.OutOrStdout())
		},
	}
	commands.AddNamespaceFlags(serviceApplyCommand.Flags(), false)
	applyFlags.AddCreateFlags(serviceApplyCommand)
	waitFlags.AddConditionWaitFlags(serviceApplyCommand, commands.WaitDefaultTimeout, "apply", "service", "ready")
	return serviceApplyCommand
}

func examineServiceForApply(cmd *cobra.Command, client clientservingv1.KnServingClient, serviceName string) (string, string, error) {
	currentService, err := client.GetService(context.TODO(), serviceName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "Creating", "created", nil
		}
		return "", "", err
	}

	annotationMap := currentService.Annotations
	if annotationMap != nil {
		if _, ok := annotationMap[corev1.LastAppliedConfigAnnotation]; !ok {
			fmt.Fprintf(cmd.OutOrStdout(), "Warning: 'kn service apply' should be used only for services created by 'kn service apply'\n")
		}
	}
	return "Applying", "applied", nil
}
