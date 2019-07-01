// Copyright © 2018 The Knative Authors
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

package commands

import (
	"io"

	serving "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"k8s.io/client-go/tools/clientcmd"
)

// CfgFile is Kn's config file is the path for the Kubernetes config
var CfgFile string

// Parameters for creating commands. Useful for inserting mocks for testing.
type KnParams struct {
	Output           io.Writer
	ServingFactory   func() (serving.ServingV1alpha1Interface, error)
	NamespaceFactory func() (string, error)

	KubeCfgPath  string
	ClientConfig clientcmd.ClientConfig
}

func (c *KnParams) Initialize() {
	if c.ServingFactory == nil {
		c.ServingFactory = c.GetConfig
	}
	if c.NamespaceFactory == nil {
		c.NamespaceFactory = c.CurrentNamespace
	}
}

func (c *KnParams) CurrentNamespace() (string, error) {
	if c.ClientConfig == nil {
		c.ClientConfig = c.GetClientConfig()
	}
	name, _, err := c.ClientConfig.Namespace()
	return name, err
}

func (c *KnParams) GetConfig() (serving.ServingV1alpha1Interface, error) {
	if c.ClientConfig == nil {
		c.ClientConfig = c.GetClientConfig()
	}
	var err error
	config, err := c.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	client, err := serving.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *KnParams) GetClientConfig() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if len(c.KubeCfgPath) > 0 {
		loadingRules.ExplicitPath = c.KubeCfgPath
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
}
