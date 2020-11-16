/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package container

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/sources/v1alpha2"
	clientv1alpha2 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha2"
)

// NewContainerCommand for managing Container source
func NewContainerCommand(p *commands.KnParams) *cobra.Command {
	containerSourceCmd := &cobra.Command{
		Use:   "container create|delete|update|list|describe",
		Short: "Manage container sources",
	}
	containerSourceCmd.AddCommand(NewContainerCreateCommand(p))
	containerSourceCmd.AddCommand(NewContainerDeleteCommand(p))
	containerSourceCmd.AddCommand(NewContainerUpdateCommand(p))
	containerSourceCmd.AddCommand(NewContainerListCommand(p))
	containerSourceCmd.AddCommand(NewContainerDescribeCommand(p))
	return containerSourceCmd
}

var containerSourceClientFactory func(config clientcmd.ClientConfig, namespace string) (v1alpha2.KnContainerSourcesClient, error)

func newContainerSourceClient(p *commands.KnParams, cmd *cobra.Command) (v1alpha2.KnContainerSourcesClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if containerSourceClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return containerSourceClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := clientv1alpha2.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return v1alpha2.NewKnSourcesClient(client, namespace).ContainerSourcesClient(), nil
}
