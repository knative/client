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
	"errors"
	"fmt"
	"io"
	"path/filepath"

	serving_kn_v1alpha1 "github.com/knative/client/pkg/serving/v1alpha1"

	serving_v1alpha1_client "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"k8s.io/client-go/tools/clientcmd"
)

// CfgFile is Kn's config file is the path for the Kubernetes config
var CfgFile string

// Parameters for creating commands. Useful for inserting mocks for testing.
type KnParams struct {
	Output       io.Writer
	KubeCfgPath  string
	ClientConfig clientcmd.ClientConfig
	NewClient    func(namespace string) (serving_kn_v1alpha1.KnClient, error)

	// Set this if you want to nail down the namespace
	fixedCurrentNamespace string
}

func (params *KnParams) Initialize() {
	if params.NewClient == nil {
		params.NewClient = params.newClient
	}
}

func (params *KnParams) newClient(namespace string) (serving_kn_v1alpha1.KnClient, error) {
	client, err := params.GetConfig()
	if err != nil {
		return nil, err
	}
	return serving_kn_v1alpha1.NewKnServingClient(client, namespace), nil
}

// GetConfig returns Serving Client
func (params *KnParams) GetConfig() (serving_v1alpha1_client.ServingV1alpha1Interface, error) {
	var err error

	if params.ClientConfig == nil {
		params.ClientConfig, err = params.GetClientConfig()
		if err != nil {
			return nil, err
		}
	}

	config, err := params.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return serving_v1alpha1_client.NewForConfig(config)
}

// GetClientConfig gets ClientConfig from KubeCfgPath
func (params *KnParams) GetClientConfig() (clientcmd.ClientConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if len(params.KubeCfgPath) > 0 {
		paths := filepath.SplitList(params.KubeCfgPath)
		if len(paths) == 1 {
			loadingRules.ExplicitPath = paths[0]
		} else {
			return nil, errors.New(fmt.Sprintf("Config file '%s' could not be found. For configuration lookup path please use the env variable KUBECONFIG", params.KubeCfgPath))
		}
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
}
