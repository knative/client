// Copyright © 2023 The Knative Authors
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

package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/client/pkg/kn/commands/flags"
	hprinters "knative.dev/client/pkg/printers"

	"knative.dev/client/pkg/kn/commands"
)

func NewSecretListCommand(p *commands.KnParams) *cobra.Command {
	listFlags := flags.NewListPrintFlags(SecretListHandlers)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List secrets",
		Aliases: []string{"ls"},
		Example: ``,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			if namespace == "" {
				listFlags.EnsureWithNamespace()
			}

			client, err := p.NewKubeClient()
			if err != nil {
				return err
			}

			list, err := client.CoreV1().Secrets(namespace).List(cmd.Context(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			if !listFlags.GenericPrintFlags.OutputFlagSpecified() && len(list.Items) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No secret found.\n")
				return nil
			}

			return listFlags.Print(list, cmd.OutOrStdout())
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	listFlags.AddFlags(cmd)
	return cmd
}

func SecretListHandlers(h hprinters.PrintHandler) {
	dmColumnDefinitions := []metav1beta1.TableColumnDefinition{
		{Name: "Namespace", Type: "string", Description: "Namespace of the Secret", Priority: 0},
		{Name: "Name", Type: "string", Description: "Name of the Secret", Priority: 1},
		{Name: "Type", Type: "string", Description: "Type of the Secret", Priority: 1},
	}
	h.TableHandler(dmColumnDefinitions, printSecret)
	h.TableHandler(dmColumnDefinitions, printSecretList)
}

func printSecretList(secretList *corev1.SecretList, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	rows := make([]metav1beta1.TableRow, 0, len(secretList.Items))
	for i := range secretList.Items {
		secret := &secretList.Items[i]
		r, err := printSecret(secret, options)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	return rows, nil
}

func printSecret(secret *corev1.Secret, options hprinters.PrintOptions) ([]metav1beta1.TableRow, error) {
	name := secret.Name
	sType := secret.Type
	row := metav1beta1.TableRow{
		Object: runtime.RawExtension{Object: secret},
	}
	if options.AllNamespaces {
		row.Cells = append(row.Cells, secret.Namespace)
	}

	row.Cells = append(row.Cells,
		name,
		sType)
	return []metav1beta1.TableRow{row}, nil
}
