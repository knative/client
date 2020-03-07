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
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	v1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/apis/duck/v1beta1"

	clientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"

	"knative.dev/client/pkg/kn/commands"
	v1alpha1 "knative.dev/client/pkg/sources/v1alpha1"
)

// NewAPIServerCommand for managing ApiServer source
func NewAPIServerCommand(p *commands.KnParams) *cobra.Command {
	apiServerSourceCmd := &cobra.Command{
		Use:   "apiserver",
		Short: "Kubernetes API Server Event Source command group",
	}
	apiServerSourceCmd.AddCommand(NewAPIServerCreateCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerUpdateCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerDescribeCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerDeleteCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerListCommand(p))
	return apiServerSourceCmd
}

var apiServerSourceClientFactory func(config clientcmd.ClientConfig, namespace string) (v1alpha1.KnAPIServerSourcesClient, error)

func newAPIServerSourceClient(p *commands.KnParams, cmd *cobra.Command) (v1alpha1.KnAPIServerSourcesClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if apiServerSourceClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return apiServerSourceClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := clientv1alpha1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return v1alpha1.NewKnSourcesClient(client, namespace).APIServerSourcesClient(), nil
}

func toDuckV1Beta1(destination *v1.Destination) *v1beta1.Destination {
	r := destination.Ref
	return &v1beta1.Destination{
		Ref: &corev1.ObjectReference{
			Kind:       r.Kind,
			Namespace:  r.Namespace,
			Name:       r.Name,
			APIVersion: r.APIVersion,
		},
		URI: destination.URI,
	}
}
