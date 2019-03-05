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
	"k8s.io/apimachinery/pkg/runtime/schema"

	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	"github.com/spf13/cobra"
)

func NewCreateCommand(p *KnParams) *cobra.Command {

	serviceCreateCommand := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
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
				Spec: servingv1alpha1.ServiceSpec{},
			}
			service.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"},
			)

			return nil
		},
	}
	return serviceCreateCommand
}
