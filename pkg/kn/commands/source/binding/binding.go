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
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	v1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1"
	"knative.dev/pkg/tracker"

	"knative.dev/client/pkg/kn/commands"
	clientsourcesv1alpha1 "knative.dev/client/pkg/sources/v1"
)

// NewBindingCommand is the root command for all binding related commands
func NewBindingCommand(p *commands.KnParams) *cobra.Command {
	bindingCmd := &cobra.Command{
		Use:   "binding COMMAND",
		Short: "Manage sink bindings",
	}
	bindingCmd.AddCommand(NewBindingCreateCommand(p))
	bindingCmd.AddCommand(NewBindingUpdateCommand(p))
	bindingCmd.AddCommand(NewBindingDeleteCommand(p))
	bindingCmd.AddCommand(NewBindingListCommand(p))
	bindingCmd.AddCommand(NewBindingDescribeCommand(p))
	return bindingCmd
}

// This var can be used to inject a factory for fake clients when doing
// tests.
var sinkBindingClientFactory func(config clientcmd.ClientConfig, namespace string) (clientsourcesv1alpha1.KnSinkBindingClient, error)

func newSinkBindingClient(p *commands.KnParams, cmd *cobra.Command) (clientsourcesv1alpha1.KnSinkBindingClient, error) {
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

	client, err := v1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return clientsourcesv1alpha1.NewKnSourcesClient(client, namespace).SinkBindingClient(), nil
}

// subjectToString converts a reference to a string representation
func subjectToString(ref tracker.Reference) string {

	ret := ref.Kind + ":" + ref.APIVersion
	if ref.Name != "" {
		return ret + ":" + ref.Name
	}
	var keyValues []string
	selector := ref.Selector
	if selector != nil {
		for k, v := range selector.MatchLabels {
			keyValues = append(keyValues, k+"="+v)
		}
		return ret + ":" + strings.Join(keyValues, ",")
	}
	return ret
}
