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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	serving_lib "github.com/knative/client/pkg/serving"
	"github.com/spf13/cobra"
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
  kn service create mysvc --env KEY1=VALUE1 --env KEY2=VALUE2 --image dev.local/ns/image:latest`,

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
			service.Spec.RunLatest = &servingv1alpha1.RunLatestType{}

			config, err := serving_lib.GetConfiguration(&service)
			if err != nil {
				return err
			}
			err = editFlags.Apply(config)
			if err != nil {
				return err
			}
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			_, err = client.Services(namespace).Create(&service)
			if err != nil {
				return err
			}

			return nil
		},
	}
	AddNamespaceFlags(serviceCreateCommand.Flags(), false)
	editFlags.AddFlags(serviceCreateCommand)
	return serviceCreateCommand
}
