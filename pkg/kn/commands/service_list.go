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

package commands

import (
	"os"

	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var serviceListPrintFlags *genericclioptions.PrintFlags

// listCmd represents the list command
func NewServiceListCommand() *cobra.Command {

	serviceListPrintFlags := genericclioptions.NewPrintFlags("").WithDefaultOutput(
		"jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}")
	serviceListCommand := &cobra.Command{
		Use:   "list",
		Short: "List available services.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// use the current context in kubeconfig
			config, err := clientcmd.BuildConfigFromFlags("", kubeCfgFile)
			if err != nil {
				return err
			}
			client, err := serving.NewForConfig(config)
			if err != nil {
				return err
			}
			namespace := cmd.Flag("namespace").Value.String()
			service, err := client.Services(namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}

			printer, err := serviceListPrintFlags.ToPrinter()
			if err != nil {
				return err
			}
			service.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Service"})
			err = printer.PrintObj(service, os.Stdout)
			if err != nil {
				return err
			}
			return nil
		},
	}
	serviceListPrintFlags.AddFlags(serviceListCommand)
	return serviceListCommand
}
