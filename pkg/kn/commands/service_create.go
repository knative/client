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

package commands

import (
	"errors"
	"fmt"

	serving_lib "github.com/knative/client/pkg/serving"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceCreateCommand(p *KnParams) *cobra.Command {
	var editFlags ConfigurationEditFlags

	serviceCreateCommand := &cobra.Command{
		Use:   "create NAME --image IMAGE",
		Short: "Create a service.",
		Example: `
  # Create a service 'mysvc' using image at dev.local/ns/image:latest
  kn service create mysvc --image dev.local/ns/image:latest

  # Create a service with multiple environment variables
  kn service create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest

  # Create or replace a service 's1' with image dev.local/ns/image:v2 using --force flag
  # if service 's1' doesn't exist, it's just a normal create operation
  kn service create --force s1 --image dev.local/ns/image:v2

  # Create or replace environment variables of service 's1' using --force flag
  kn service create --force s1 --env KEY1=NEW_VALUE1 --env NEW_KEY2=NEW_VALUE2 --image dev.local/ns/image:v1

  # Create or replace default resources of a service 's1' using --force flag
  # (earlier configured resource requests and limits will be replaced with default)
  # (earlier configured environment variables will be cleared too if any)
  kn service create --force s1 --image dev.local/ns/image:v1`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the service name.")
			}
			if editFlags.Image == "" {
				return errors.New("requires the image name to run.")
			}

			namespace, err := GetNamespace(cmd)
			if err != nil {
				return err
			}

			service := servingv1alpha1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      args[0],
					Namespace: namespace,
				},
			}
			service.Spec.DeprecatedRunLatest = &servingv1alpha1.RunLatestType{}

			config, err := serving_lib.GetConfiguration(&service)
			if err != nil {
				return err
			}
			err = editFlags.Apply(config, cmd)
			if err != nil {
				return err
			}
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			var serviceExists bool = false
			if editFlags.ForceCreate {
				existingService, err := client.Services(namespace).Get(args[0], v1.GetOptions{})
				if err == nil {
					serviceExists = true
					service.ResourceVersion = existingService.ResourceVersion
					_, err = client.Services(namespace).Update(&service)
					if err != nil {
						return err
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Service '%s' successfully replaced in namespace '%s'.\n", args[0], namespace)
				}
			}
			if !serviceExists {
				_, err = client.Services(namespace).Create(&service)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Service '%s' successfully created in namespace '%s'.\n", args[0], namespace)
			}
			return nil
		},
	}
	AddNamespaceFlags(serviceCreateCommand.Flags(), false)
	editFlags.AddCreateFlags(serviceCreateCommand)
	return serviceCreateCommand
}
