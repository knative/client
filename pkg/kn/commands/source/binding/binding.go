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

package binding

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"

	"knative.dev/client/pkg/kn/commands"
	spources_v1alpha1 "knative.dev/client/pkg/sources/v1alpha1"
)

// NewBindingCommand is the root command for all binding related commands
func NewBindingCommand(p *commands.KnParams) *cobra.Command {
	bindingCmd := &cobra.Command{
		Use:   "binding",
		Short: "Sink binding command group",
	}
	bindingCmd.AddCommand(NewBindingCreateCommand(p))
	return bindingCmd
}

var sinkBindingClientFactory func(config clientcmd.ClientConfig, namespace string) (spources_v1alpha1.KnSinkBindingClient, error)

func newSinkBindingClient(p *commands.KnParams, cmd *cobra.Command) (spources_v1alpha1.KnSinkBindingClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if sinkBindingClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return sinkBindingClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := v1alpha1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return spources_v1alpha1.NewKnSourcesClient(client, namespace).SinkBindingClient(), nil
}
