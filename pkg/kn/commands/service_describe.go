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

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewServiceDescribeCommand(p *KnParams) *cobra.Command {

	serviceDescribePrintFlags := genericclioptions.NewPrintFlags("").WithDefaultOutput("yaml")
	serviceDescribeCommand := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe available services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires the service name.")
			}
			cmd.SilenceUsage = true

			client, err := p.ServingFactory()
			if err != nil {
				return err
			}

			namespace := cmd.Flag("namespace").Value.String()
			describeService, err := client.Services(namespace).Get(args[0], v1.GetOptions{})
			if err != nil {
				return err
			}

			printer, err := serviceDescribePrintFlags.ToPrinter()
			if err != nil {
				return err
			}
			describeService.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"})
			err = printer.PrintObj(describeService, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	serviceDescribePrintFlags.AddFlags(serviceDescribeCommand)
	return serviceDescribeCommand
}
