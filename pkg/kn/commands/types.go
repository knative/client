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

// PluginDir is Kn's plugin directory or directory list to search and
// install plugins
var PluginDir string

// KubeCfgFile is the path for the Kubernetes config
var KubeCfgFile string

// Parameters for creating commands. Useful for inserting mocks for testing.
type KnParams struct {
	Output         io.Writer
	ServingFactory func() (serving.ServingV1alpha1Interface, error)
}

func (c *KnParams) Initialize() {
	if c.ServingFactory == nil {
		c.ServingFactory = GetConfig
	}
}

func GetConfig() (serving.ServingV1alpha1Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", KubeCfgFile)
	if err != nil {
		return nil, err
	}
	client, err := serving.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
