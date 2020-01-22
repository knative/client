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

package cronjob

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	sources_v1alpha1 "knative.dev/eventing/pkg/legacyclient/clientset/versioned/typed/legacysources/v1alpha1"

	"knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
)

// NewCronJobCommand is the root command for all cronjob related commands
func NewCronJobCommand(p *commands.KnParams) *cobra.Command {
	cronImporterCmd := &cobra.Command{
		Use:   "cronjob",
		Short: "CronJob source command group",
	}
	cronImporterCmd.AddCommand(NewCronJobCreateCommand(p))
	cronImporterCmd.AddCommand(NewCronJobDeleteCommand(p))
	cronImporterCmd.AddCommand(NewCronJobDescribeCommand(p))
	cronImporterCmd.AddCommand(NewCronJobUpdateCommand(p))
	cronImporterCmd.AddCommand(NewCronJobListCommand(p))
	return cronImporterCmd
}

var cronJobSourceClientFactory func(config clientcmd.ClientConfig, namespace string) (v1alpha1.KnCronJobSourcesClient, error)

func newCronJobSourceClient(p *commands.KnParams, cmd *cobra.Command) (v1alpha1.KnCronJobSourcesClient, error) {
	namespace, err := p.GetNamespace(cmd)
	if err != nil {
		return nil, err
	}

	if cronJobSourceClientFactory != nil {
		config, err := p.GetClientConfig()
		if err != nil {
			return nil, err
		}
		return cronJobSourceClientFactory(config, namespace)
	}

	clientConfig, err := p.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := sources_v1alpha1.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return v1alpha1.NewKnSourcesClient(client, namespace).CronJobSourcesClient(), nil
}
