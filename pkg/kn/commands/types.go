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
	eventingv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1"
	messagingv1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1"
	sourcesv1client "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1"
	sourcesv1beta2client "knative.dev/eventing/pkg/client/clientset/versioned/typed/sources/v1beta2"
	servingv1client "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	servingv1alpha1client "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"

	"knative.dev/client/pkg/util"

	clientdynamic "knative.dev/client/pkg/dynamic"
	knerrors "knative.dev/client/pkg/errors"
	clienteventingv1 "knative.dev/client/pkg/eventing/v1"
	clientmessagingv1 "knative.dev/client/pkg/messaging/v1"
	clientservingv1 "knative.dev/client/pkg/serving/v1"
	clientservingv1alpha1 "knative.dev/client/pkg/serving/v1alpha1"
	clientsourcesv1 "knative.dev/client/pkg/sources/v1"
	clientsourcesv1beta2 "knative.dev/client/pkg/sources/v1beta2"
)

// KnParams for creating commands. Useful for inserting mocks for testing.
type KnParams struct {
	Output                   io.Writer
	KubeCfgPath              string
	KubeContext              string
	KubeCluster              string
	ClientConfig             clientcmd.ClientConfig
	NewServingClient         func(namespace string) (clientservingv1.KnServingClient, error)
	NewServingV1alpha1Client func(namespace string) (clientservingv1alpha1.KnServingClient, error)
	NewGitopsServingClient   func(namespace string, dir string) (clientservingv1.KnServingClient, error)
	NewSourcesClient         func(namespace string) (clientsourcesv1.KnSourcesClient, error)
	NewSourcesV1beta2Client  func(namespace string) (clientsourcesv1beta2.KnSourcesClient, error)
	NewEventingClient        func(namespace string) (clienteventingv1.KnEventingClient, error)
	NewMessagingClient       func(namespace string) (clientmessagingv1.KnMessagingClient, error)
	NewDynamicClient         func(namespace string) (clientdynamic.KnDynamicClient, error)

	// General global options
	LogHTTP bool

	// Set this if you want to nail down the namespace
	fixedCurrentNamespace string
}

func (params *KnParams) Initialize() {
	if params.NewServingClient == nil {
		params.NewServingClient = params.newServingClient
	}

	if params.NewServingV1alpha1Client == nil {
		params.NewServingV1alpha1Client = params.newServingClientV1alpha1
	}

	if params.NewGitopsServingClient == nil {
		params.NewGitopsServingClient = params.newGitopsServingClient
	}

	if params.NewSourcesClient == nil {
		params.NewSourcesClient = params.newSourcesClient
	}

	if params.NewEventingClient == nil {
		params.NewEventingClient = params.newEventingClient
	}

	if params.NewMessagingClient == nil {
		params.NewMessagingClient = params.newMessagingClient
	}

	if params.NewDynamicClient == nil {
		params.NewDynamicClient = params.newDynamicClient
	}
}

func (params *KnParams) newServingClient(namespace string) (clientservingv1.KnServingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := servingv1client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientservingv1.NewKnServingClient(client, namespace), nil
}

func (params *KnParams) newServingClientV1alpha1(namespace string) (clientservingv1alpha1.KnServingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, err := servingv1alpha1client.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientservingv1alpha1.NewKnServingClient(client, namespace), nil
}

func (params *KnParams) newGitopsServingClient(namespace string, dir string) (clientservingv1.KnServingClient, error) {
	return clientservingv1.NewKnServingGitOpsClient(namespace, dir), nil
}

func (params *KnParams) newSourcesClient(namespace string) (clientsourcesv1.KnSourcesClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := sourcesv1client.NewForConfig(restConfig)
	return clientsourcesv1.NewKnSourcesClient(client, namespace), nil
}

func (params *KnParams) newSourcesClientV1beta2(namespace string) (clientsourcesv1beta2.KnSourcesClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := sourcesv1beta2client.NewForConfig(restConfig)
	return clientsourcesv1beta2.NewKnSourcesClient(client, namespace), nil
}

func (params *KnParams) newEventingClient(namespace string) (clienteventingv1.KnEventingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := eventingv1.NewForConfig(restConfig)
	return clienteventingv1.NewKnEventingClient(client, namespace), nil
}

func (params *KnParams) newMessagingClient(namespace string) (clientmessagingv1.KnMessagingClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := messagingv1.NewForConfig(restConfig)
	return clientmessagingv1.NewKnMessagingClient(client, namespace), nil
}

func (params *KnParams) newDynamicClient(namespace string) (clientdynamic.KnDynamicClient, error) {
	restConfig, err := params.RestConfig()
	if err != nil {
		return nil, err
	}

	client, _ := dynamic.NewForConfig(restConfig)
	return clientdynamic.NewKnDynamicClient(client, namespace), nil
}

// RestConfig returns REST config, which can be to use to create specific clientset
func (params *KnParams) RestConfig() (*rest.Config, error) {
	var err error

	if params.ClientConfig == nil {
		params.ClientConfig, err = params.GetClientConfig()
		if err != nil {
			return nil, knerrors.GetError(err)
		}
	}

	config, err := params.ClientConfig.ClientConfig()
	if err != nil {
		return nil, knerrors.GetError(err)
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
	configOverrides := &clientcmd.ConfigOverrides{}
	if params.KubeContext != "" {
		configOverrides.CurrentContext = params.KubeContext
	}
	if params.KubeCluster != "" {
		configOverrides.Context.Cluster = params.KubeCluster
	}
	if len(params.KubeCfgPath) == 0 {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
	}

	_, err := os.Stat(params.KubeCfgPath)
	if err == nil {
		loadingRules.ExplicitPath = params.KubeCfgPath
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	paths := filepath.SplitList(params.KubeCfgPath)
	if len(paths) > 1 {
		return nil, fmt.Errorf("can not find config file. '%s' looks like a path. "+
			"Please use the env var KUBECONFIG if you want to check for multiple configuration files", params.KubeCfgPath)
	}
	return nil, fmt.Errorf("config file '%s' can not be found", params.KubeCfgPath)
}
