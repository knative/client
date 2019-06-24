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

package revision

import (
	"errors"

	"github.com/knative/client/pkg/kn/commands"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewRevisionDescribeCommand(p *commands.KnParams) *cobra.Command {
	revisionDescribePrintFlags := genericclioptions.NewPrintFlags("").WithDefaultOutput("yaml")
	revisionDescribeCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe revisions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires the revision name.")
			}

			client, err := p.ServingFactory()
			if err != nil {
				return err
			}

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			revision, err := client.Revisions(namespace).Get(args[0], v1.GetOptions{})
			if err != nil {
				return err
			}

			printer, err := revisionDescribePrintFlags.ToPrinter()
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
	commands.AddNamespaceFlags(revisionDescribeCmd.Flags(), false)
	revisionDescribePrintFlags.AddFlags(revisionDescribeCmd)
	return revisionDescribeCmd
}
