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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	serving_cli "github.com/knative/client/pkg/serving"
	"github.com/spf13/cobra"
)

func NewServiceCreateCommand(p *KnParams) *cobra.Command {

	var image string
	var env map[string]string

	serviceCreateCommand := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a service.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			changes := []serving_cli.ConfigChange{}
			if image != "" {
				changes = append(changes, serving_cli.ImageUpdate(image))
			}
			if len(env) > 0 {
				changes = append(changes, serving_cli.EnvVarUpdate(env))
			}

			if len(args) < 1 {
				return errors.New("requires the service name.")
			}

			namespace := cmd.Flag("namespace").Value.String()

			service := servingv1alpha1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      args[0],
					Namespace: namespace,
				},
				Spec: servingv1alpha1.ServiceSpec{
					RunLatest: &servingv1alpha1.RunLatestType{
						Configuration: servingv1alpha1.ConfigurationSpec{
							RevisionTemplate: servingv1alpha1.RevisionTemplateSpec{
								Spec: servingv1alpha1.RevisionSpec{
									Container: corev1.Container{},
								},
							},
						},
					},
				},
			}
			service.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"},
			)

			for _, change := range changes {
				err = change(&service.Spec.RunLatest.Configuration)
				if err != nil {
					return err
				}
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
	AddConfigurationEditFlags(serviceCreateCommand, &image, &env)
	return serviceCreateCommand
}
