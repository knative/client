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

package apiserver

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
)

func NewApiServerCreateCommand(p *commands.KnParams) *cobra.Command {
	var apiServerUpdateFlags ApiServerSourceUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --resource RESOURCE --service-account ACCOUNTNAME --sink SINK --mode MODE",
		Short: "Create an ApiServerSource, which watches for Kubernetes events and forwards them to a sink",
		Example: `
  # Create an ApiServerSource 'k8sevents' which consumes Kubernetes events and sends message to service 'mysvc' as a cloudevent
  kn source apiserver create k8sevents --resource Event --service-account myaccountname --sink svc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'source apiserver create' requires the name of the source as single argument")
			}
			name := args[0]

			// get namespace
			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			// get client
			sourcesClient, err := p.NewSourcesClient(namespace)
			if err != nil {
				return err
			}

			// resolve sink
			servingClient, err := p.NewServingClient(namespace)
			if err != nil {
				return err
			}

			objectRef, err := sinkFlags.ResolveSink(servingClient)
			if err != nil {
				return fmt.Errorf(
					"cannot create ApiServerSource '%s' in namespace '%s' "+
						"because %s", name, namespace, err)
			}

			// construct ApiServerSource
			apisrvsrc := constructApiServerSource(name, namespace, apiServerUpdateFlags)
			apisrvsrc.Spec.Sink = objectRef

			// create
			err = sourcesClient.ApiServerSourcesClient().CreateApiServerSource(apisrvsrc)
			if err != nil {
				return fmt.Errorf(
					"cannot create ApiServerSource '%s' in namespace '%s' "+
						"because %s", name, namespace, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ApiServerSource '%s' successfully created in namespace '%s'.\n", args[0], namespace)
			return nil
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	apiServerUpdateFlags.Add(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("schedule")

	return cmd
}

// constructApiServerSource is to create an instance of v1alpha1.ApiServerSource
func constructApiServerSource(name string, namespace string, apiServerFlags ApiServerSourceUpdateFlags) *v1alpha1.ApiServerSource {

	source := v1alpha1.ApiServerSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ApiServerSourceSpec{
			Resources:          apiServerFlags.GetApiServerResourceArray(),
			ServiceAccountName: apiServerFlags.ServiceAccountName,
			Mode:               apiServerFlags.Mode,
		},
	}

	return &source
}
