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
	"k8s.io/client-go/tools/clientcmd"
	clientv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1"

	v1 "knative.dev/client-pkg/pkg/sources/v1"
	"knative.dev/client/pkg/commands"
)

// NewAPIServerCommand for managing ApiServer source
func NewAPIServerCommand(p *commands.KnParams) *cobra.Command {
	apiServerSourceCmd := &cobra.Command{
		Use:   "apiserver COMMAND",
		Short: "Manage Kubernetes api-server sources",
	}
	apiServerSourceCmd.AddCommand(NewAPIServerCreateCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerUpdateCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerDescribeCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerDeleteCommand(p))
	apiServerSourceCmd.AddCommand(NewAPIServerListCommand(p))
	return apiServerSourceCmd
}

var apiServerSourceClientFactory func(config clientcmd.ClientConfig, namespace string) (v1.KnAPIServerSourcesClient, error)

func newAPIServerSourceClient(p *commands.KnParams, cmd *cobra.Command) (v1.KnAPIServerSourcesClient, error) {
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

	client, err := clientv1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return v1.NewKnSourcesClient(client, namespace).APIServerSourcesClient(), nil
}
