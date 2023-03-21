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
	"errors"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/kn/commands"
)

func NewSecretDeleteCommand(p *commands.KnParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete NAME",
		Short:   "Delete secret",
		Example: ``,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn secret delete' requires the Secret name given as single argument")
			}
			name := args[0]
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			client, err := p.NewKubeClient()
			if err != nil {
				return err
			}
			return client.CoreV1().Secrets(namespace).Delete(cmd.Context(), name, metav1.DeleteOptions{})
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	return cmd
}
