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

package commands

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gotest.tools/assert"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"knative.dev/client/pkg/util"
)

type configTestCase struct {
	clientConfig      clientcmd.ClientConfig
	expectedErrString string
	logHttp           bool
}

var BASIC_KUBECONFIG = `apiVersion: v1
kind: Config
preferences: {}
users:
- name: a
  user:
    client-certificate-data: ""
    client-key-data: ""
clusters:
- name: a
  cluster:
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:8080
contexts:
- name: a
  context:
    cluster: a
    user: a
current-context: a
`

func TestPrepareConfig(t *testing.T) {
	basic, err := clientcmd.NewClientConfigFromBytes([]byte(BASIC_KUBECONFIG))
	if err != nil {
		t.Error(err)
	}
	for i, tc := range []configTestCase{
		{
			clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{}),
			"no kubeconfig has been provided, please use a valid configuration to connect to the cluster",
			false,
		},
		{
			basic,
			"",
			false,
		},
		{ // Test that the cast to wrap the http client in a logger works
			basic,
			"",
			true,
		},
	} {
		p := &KnParams{
			ClientConfig: tc.clientConfig,
			LogHTTP:      tc.logHttp,
		}

		_, err := p.RestConfig()

		switch len(tc.expectedErrString) {
		case 0:
			if err != nil {
				t.Errorf("%d: unexpected error: %s", i, err.Error())
			}
		default:
			if err == nil {
				t.Errorf("%d: wrong error detected: %s (expected) != %s (actual)", i, tc.expectedErrString, err)
			}
			if !strings.Contains(err.Error(), tc.expectedErrString) {
				t.Errorf("%d: wrong error detected: %s (expected) != %s (actual)", i, tc.expectedErrString, err.Error())
			}
		}
	}
}

type typeTestCase struct {
	kubeCfgPath   string
	explicitPath  string
	expectedError string
}

func TestGetClientConfig(t *testing.T) {
	multiConfigs := fmt.Sprintf("%s%s%s", "/testing/assets/kube-config-01.yml", string(os.PathListSeparator), "/testing/assets/kube-config-02.yml")
	for _, tc := range []typeTestCase{
		{
			"",
			clientcmd.NewDefaultClientConfigLoadingRules().ExplicitPath,
			"",
		},
		{
			"/testing/assets/kube-config-01.yml",
			"",
			fmt.Sprintf("Config file '%s' can not be found", "/testing/assets/kube-config-01.yml"),
		},
		{
			multiConfigs,
			"",
			fmt.Sprintf("Can not find config file. '%s' looks like a path. Please use the env var KUBECONFIG if you want to check for multiple configuration files", multiConfigs),
		},
	} {
		p := &KnParams{
			KubeCfgPath: tc.kubeCfgPath,
		}

		clientConfig, err := p.GetClientConfig()
		if tc.expectedError != "" {
			assert.Assert(t, util.ContainsAll(err.Error(), tc.expectedError))
		} else {
			assert.Assert(t, err == nil, err)
		}

		if clientConfig != nil {
			configAccess := clientConfig.ConfigAccess()

			assert.Assert(t, configAccess.GetExplicitFile() == tc.explicitPath)
		}
	}
}

func TestNewSourcesClient(t *testing.T) {
	basic, err := clientcmd.NewClientConfigFromBytes([]byte(BASIC_KUBECONFIG))
	namespace := "test"
	if err != nil {
		t.Error(err)
	}
	for i, tc := range []configTestCase{
		{
			clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{}),
			"no kubeconfig has been provided, please use a valid configuration to connect to the cluster",
			false,
		},
		{
			basic,
			"",
			false,
		},
		{ // Test that the cast to wrap the http client in a logger works
			basic,
			"",
			true,
		},
	} {
		p := &KnParams{
			ClientConfig: tc.clientConfig,
			LogHTTP:      tc.logHttp,
		}

		sourcesClient, err := p.newSourcesClient(namespace)

		switch len(tc.expectedErrString) {
		case 0:
			if err != nil {
				t.Errorf("%d: unexpected error: %s", i, err.Error())
			}
		default:
			if err == nil {
				t.Errorf("%d: wrong error detected: %s (expected) != %s (actual)", i, tc.expectedErrString, err)
			}
			if !strings.Contains(err.Error(), tc.expectedErrString) {
				t.Errorf("%d: wrong error detected: %s (expected) != %s (actual)", i, tc.expectedErrString, err.Error())
			}
		}

		if sourcesClient != nil {
			assert.Assert(t, sourcesClient.SinkBindingClient().Namespace() == namespace)
			assert.Assert(t, sourcesClient.PingSourcesClient().Namespace() == namespace)
		}
	}
}

func TestNewDynamicClient(t *testing.T) {
	basic, err := clientcmd.NewClientConfigFromBytes([]byte(BASIC_KUBECONFIG))
	namespace := "test"
	if err != nil {
		t.Error(err)
	}
	for i, tc := range []configTestCase{
		{
			clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{}),
			"no kubeconfig has been provided, please use a valid configuration to connect to the cluster",
			false,
		},
		{
			basic,
			"",
			false,
		},
		{ // Test that the cast to wrap the http client in a logger works
			basic,
			"",
			true,
		},
	} {
		p := &KnParams{
			ClientConfig: tc.clientConfig,
			LogHTTP:      tc.logHttp,
		}

		dynamicClient, err := p.newDynamicClient(namespace)

		switch len(tc.expectedErrString) {
		case 0:
			if err != nil {
				t.Errorf("%d: unexpected error: %s", i, err.Error())
			}
		default:
			if err == nil {
				t.Errorf("%d: wrong error detected: %s (expected) != %s (actual)", i, tc.expectedErrString, err)
			}
			if !strings.Contains(err.Error(), tc.expectedErrString) {
				t.Errorf("%d: wrong error detected: %s (expected) != %s (actual)", i, tc.expectedErrString, err.Error())
			}
		}

		if dynamicClient != nil {
			assert.Assert(t, dynamicClient.Namespace() == namespace)
		}
	}
}
