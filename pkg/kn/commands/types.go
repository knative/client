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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client/pkg/util"
	eventing "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	eventing_sources "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1alpha1"
	serving_v1alpha1_client "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"

	dynamic_kn "knative.dev/client/pkg/dynamic"
	sources_kn_v1alpha1 "knative.dev/client/pkg/eventing/sources/v1alpha1"
	eventing_kn_v1alpha1 "knative.dev/client/pkg/eventing/v1alpha1"
	serving_kn_v1alpha1 "knative.dev/client/pkg/serving/v1alpha1"
)

// CfgFile is Kn's config file is the path for the Kubernetes config
var CfgFile string

// Cfg is Kn's configuration values
var Cfg Config

// Config contains the variables for the Kn config
type Config struct {
	PluginsDir    string
	LookupPlugins bool
}

// KnParams for creating commands. Useful for inserting mocks for testing.
type KnParams struct {
	Output            io.Writer
	KubeCfgPath       string
	ClientConfig      clientcmd.ClientConfig
	NewServingClient  func(namespace string) (serving_kn_v1alpha1.KnServingClient, error)
	NewSourcesClient  func(namespace string) (sources_kn_v1alpha1.KnSourcesClient, error)
	NewEventingClient func(namespace string) (eventing_kn_v1alpha1.KnEventingClient, error)
	NewDynamicClient  func(namespace string) (dynamic_kn.KnDynamicClient, error)

	// General global options
	LogHTTP bool

	// Set this if you want to nail down the namespace
	fixedCurrentNamespace string
}

func (params *KnParams) Initialize() {
	if params.NewServingClient == nil {
		params.NewServingClient = params.newServingClient
	}

	if params.NewSourcesClient == nil {
		params.NewSourcesClient = params.newSourcesClient
	}

	if params.NewEventingClient == nil {
		params.NewEventingClient = params.newEventingClient
	}

	if params.NewDynamicClient == nil {
		params.NewDynamicClient = params.newDynamicClient
	}
}

func (params *KnParams) newServingClient(namespace string) (serving_kn_v1alpha1.KnServingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := serving_v1alpha1_client.NewForConfig(restConfig)
	return serving_kn_v1alpha1.NewKnServingClient(client, namespace), nil
}

func (params *KnParams) newSourcesClient(namespace string) (sources_kn_v1alpha1.KnSourcesClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := eventing_sources.NewForConfig(restConfig)
	return sources_kn_v1alpha1.NewKnSourcesClient(client, namespace), nil
}

func (params *KnParams) newEventingClient(namespace string) (eventing_kn_v1alpha1.KnEventingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := eventing.NewForConfig(restConfig)
	return eventing_kn_v1alpha1.NewKnEventingClient(client, namespace), nil
}

func (params *KnParams) newDynamicClient(namespace string) (dynamic_kn.KnDynamicClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := dynamic.NewForConfig(restConfig)
	return dynamic_kn.NewKnDynamicClient(client, namespace), nil
}

// RestConfig returns REST config, which can be to use to create specific clientset
func (params *KnParams) RestConfig() (*rest.Config, error) {
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
	if params.LogHTTP {
		// TODO: When we update to the newer version of client-go, replace with
		// config.Wrap() for future compat.
		config.WrapTransport = util.NewLoggingTransport
	}

	return config, nil
}

// GetClientConfig gets ClientConfig from KubeCfgPath
func (params *KnParams) GetClientConfig() (clientcmd.ClientConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if len(params.KubeCfgPath) == 0 {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
	}

	_, err := os.Stat(params.KubeCfgPath)
	if err == nil {
		loadingRules.ExplicitPath = params.KubeCfgPath
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}), nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	paths := filepath.SplitList(params.KubeCfgPath)
	if len(paths) > 1 {
		return nil, fmt.Errorf("Can not find config file. '%s' looks like a path. "+
			"Please use the env var KUBECONFIG if you want to check for multiple configuration files", params.KubeCfgPath)
	}
	return nil, fmt.Errorf("Config file '%s' can not be found", params.KubeCfgPath)
}
