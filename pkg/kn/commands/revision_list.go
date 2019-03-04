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
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var revisionListPrintFlags *genericclioptions.PrintFlags

// listCmd represents the list command
func NewRevisionListCommand(p *KnParams) *cobra.Command {
	var getAllNamespaces bool
	revisionListPrintFlags = genericclioptions.NewPrintFlags("").WithDefaultOutput(
		"jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}")
	revisionListCmd := &cobra.Command{
		Use:   "list",
		Short: "List available revisions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := p.ServingFactory()
			if err != nil {
				return err
			}
			namespace := cmd.Flag("namespace").Value.String()
			if getAllNamespaces {
				namespace = ""
			}
			revision, err := client.Revisions(namespace).List(v1.ListOptions{})
			if err != nil {
				return err
			}

			printer, err := revisionListPrintFlags.ToPrinter()
			if err != nil {
				return err
			}
			revision.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "knative.dev",
				Version: "v1alpha1",
				Kind:    "Revision"})
			err = printer.PrintObj(revision, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}
	revisionListPrintFlags.AddFlags(revisionListCmd)
	revisionListCmd.PersistentFlags().BoolVar(&getAllNamespaces, "all-namespaces", false,
		"If present, list the requested object(s) across all namespaces. Namespace in current "+
			"context is ignored even if specified with --namespace.")
	return revisionListCmd
}
