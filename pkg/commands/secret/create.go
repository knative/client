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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/util"

	"knative.dev/client/pkg/commands"
)

func NewSecretCreateCommand(p *commands.KnParams) *cobra.Command {
	var literals []string
	var cert, key, secretType string
	cmd := &cobra.Command{
		Use:     "create NAME",
		Short:   "Create secret",
		Example: ``,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn secret create' requires the secret name given as single argument")
			}
			name := args[0]
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("from-literal") && (cmd.Flags().Changed("tls-cert") || cmd.Flags().Changed("tls-key")) {
				return errors.New("TLS flags can't be combined with other options")
			}
			if cmd.Flags().Changed("tls-cert") != cmd.Flags().Changed("tls-key") {
				return errors.New("both --tls-cert and --tls-key flags are required")
			}

			toCreate := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name}}
			// --from-literal
			if len(literals) > 0 {
				literalsMap, err := util.MapFromArray(literals, "=")
				if err != nil {
					return err
				}
				toCreate.StringData = literalsMap
			}
			// --tls-cert && --tls-key
			if cert != "" && key != "" {
				cerFromFile, err := os.ReadFile(cert)
				if err != nil {
					return err
				}
				keyFromFile, err := os.ReadFile(key)
				if err != nil {
					return err
				}
				certData := map[string]string{
					"tls.cert": string(cerFromFile),
					"tls.key":  string(keyFromFile),
				}
				toCreate.Type = corev1.SecretTypeTLS
				toCreate.StringData = certData
			}
			// override Secret type with provided value, otherwise `Opaque`
			if secretType != "" {
				toCreate.Type = corev1.SecretType(secretType)
			}

			client, err := p.NewKubeClient()
			if err != nil {
				return err
			}
			_, err = client.CoreV1().Secrets(namespace).Create(cmd.Context(), &toCreate, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Secret '%s' created in namespace '%s'.\n", name, namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	cmd.Flags().StringSliceVarP(&literals, "from-literal", "l", []string{}, "Specify comma separated list of key=value pairs.")
	cmd.Flags().StringVar(&secretType, "type", "", "Specify Secret type.")
	cmd.Flags().StringVar(&cert, "tls-cert", "", "Path to TLS certificate file.")
	cmd.Flags().StringVar(&key, "tls-key", "", "Path to TLS key file.")
	return cmd
}
