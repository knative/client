// Copyright © 2021 The Knative Authors
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

package domain

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/validation"
	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"
	clientv1alpha1 "knative.dev/client/pkg/serving/v1alpha1"
)

// NewDomainMappingCreateCommand to create event channels
func NewDomainMappingCreateCommand(p *commands.KnParams) *cobra.Command {
	var refFlags RefFlags
	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a domain mapping",
		Example: `
  # Create a domain mappings 'hello.example.com' for Knative service 'hello' with TLS secret set (Beta)
  kn domain create hello.example.com --ref hello --tls my-cert-secret `,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn domain create' requires the domain name given as single argument")
			}
			name := args[0]
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}
			reference, err := refFlags.Resolve(cmd.Context(), dynamicClient, namespace)
			if err != nil {
				return err
			}

			errs := validation.IsDNS1123Subdomain(refFlags.tls)
			if refFlags.tls != "" && len(errs) != 0 {
				return fmt.Errorf("invalid secret name %q: %s", refFlags.tls, errs[0])
			}
			builder := clientv1alpha1.NewDomainMappingBuilder(name).
				Namespace(namespace).
				Reference(*reference).
				TLS(refFlags.tls)

			client, err := p.NewServingV1alpha1Client(namespace)
			if err != nil {
				return err
			}
			err = client.CreateDomainMapping(cmd.Context(), builder.Build())
			if err != nil {
				return knerrors.GetError(err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Domain mapping '%s' created in namespace '%s'.\n", name, namespace)
			return nil
		},
	}
	cmd.Flags().StringVar(&refFlags.tls, "tls", "", "Enable TLS and point to the secret that holds the server certificate.")
	commands.AddNamespaceFlags(cmd.Flags(), false)
	refFlags.Add(cmd)
	cmd.MarkFlagRequired("ref")
	return cmd
}
