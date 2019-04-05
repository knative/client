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

	serving_lib "github.com/knative/client/pkg/serving"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceUpdateCommand(p *KnParams) *cobra.Command {
	var editFlags ConfigurationEditFlags

	serviceUpdateCommand := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a service.",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the service name.")
			}

			namespace := cmd.Flag("namespace").Value.String()

			client, err := p.ServingFactory()
			if err != nil {
				return err
			}

			service, err := client.Services(namespace).Get(args[0], v1.GetOptions{})
			if err != nil {
				return err
			}

			config, err := serving_lib.GetConfiguration(service)
			if err != nil {
				return err
			}
			err = editFlags.Apply(config)
			if err != nil {
				return err
			}

			_, err = client.Services(namespace).Update(service)
			if err != nil {
				return err
			}

			return nil
		},
	}
	editFlags.AddFlags(serviceUpdateCommand)
	return serviceUpdateCommand
}
