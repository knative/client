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

package ping

import (
	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"

	"knative.dev/client/pkg/commands"
	clientv1 "knative.dev/client/pkg/sources/v1"
	sourcesv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1"
)

// NewPingCommand is the root command for all Ping source related commands
func NewPingCommand(p *commands.KnParams) *cobra.Command {
	pingImporterCmd := &cobra.Command{
		Use:   "ping COMMAND",
		Short: "Manage ping sources",
	}
	pingImporterCmd.AddCommand(NewPingCreateCommand(p))
	pingImporterCmd.AddCommand(NewPingDeleteCommand(p))
	pingImporterCmd.AddCommand(NewPingDescribeCommand(p))
	pingImporterCmd.AddCommand(NewPingUpdateCommand(p))
	pingImporterCmd.AddCommand(NewPingListCommand(p))
	return pingImporterCmd
}

var pingSourceClientFactory func(config clientcmd.ClientConfig, namespace string) (clientv1.KnPingSourcesClient, error)

func newPingSourceClient(p *commands.KnParams, cmd *cobra.Command) (clientv1.KnPingSourcesClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if pingSourceClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return pingSourceClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := sourcesv1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return clientv1.NewKnSourcesClient(client, namespace).PingSourcesClient(), nil
}
