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
	"path/filepath"
	"strings"
	"testing"

	"knative.dev/client/lib/test"

	"gotest.tools/v3/assert"
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
	assert.NilError(t, err)
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
	kpEmptyConfig := &KnParams{}
	kpEmptyConfig.ClientConfig, err = clientcmd.NewClientConfigFromBytes([]byte(""))
	assert.NilError(t, err)
	_, err = kpEmptyConfig.RestConfig()
	assert.ErrorContains(t, err, "no kubeconfig")

	kpEmptyConfig = &KnParams{}
	kpEmptyConfig.KubeCfgPath = filepath.Join("non", "existing", "file")
	_, err = kpEmptyConfig.RestConfig()
	assert.ErrorContains(t, err, "can not be found")
}

type typeTestCase struct {
	kubeCfgPath   string
	kubeContext   string
	kubeAsUser    string
	kubeAsUID     string
	kubeAsGroup   []string
	kubeCluster   string
	explicitPath  string
	expectedError string
}

func TestGetClientConfig(t *testing.T) {
	multiConfigs := fmt.Sprintf("%s%s%s", "/testing/assets/kube-config-01.yml", string(os.PathListSeparator), "/testing/assets/kube-config-02.yml")

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "mock")
	err := os.WriteFile(tempFile, []byte(BASIC_KUBECONFIG), test.FileModeReadWrite)
	assert.NilError(t, err)

	for _, tc := range []typeTestCase{
		{
			"",
			"",
			"",
			"",
			[]string{},
			"",
			clientcmd.NewDefaultClientConfigLoadingRules().ExplicitPath,
			"",
		},
		{
			tempFile,
			"",
			"",
			"",
			[]string{},
			"",
			tempFile,
			"",
		},
		{
			"/testing/assets/kube-config-01.yml",
			"foo",
			"",
			"",
			[]string{},
			"bar",
			"",
			fmt.Sprintf("config file '%s' can not be found", "/testing/assets/kube-config-01.yml"),
		},
		{
			multiConfigs,
			"",
			"",
			"",
			[]string{},
			"",
			"",
			fmt.Sprintf("can not find config file. '%s' looks like a path. Please use the env var KUBECONFIG if you want to check for multiple configuration files", multiConfigs),
		},
		{
			tempFile,
			"",
			"admin",
			"",
			[]string{},
			"",
			tempFile,
			"",
		},
		{
			tempFile,
			"",
			"admin",
			"",
			[]string{"system:authenticated", "system:masters"},
			"",
			tempFile,
			"",
		},
		{
			tempFile,
			"",
			"admin",
			"abc123",
			[]string{},
			"",
			tempFile,
			"",
		},
	} {
		p := &KnParams{
			KubeCfgPath: tc.kubeCfgPath,
			KubeContext: tc.kubeContext,
			KubeAsUser:  tc.kubeAsUser,
			KubeAsUID:   tc.kubeAsUID,
			KubeAsGroup: tc.kubeAsGroup,
			KubeCluster: tc.kubeCluster,
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

			if tc.kubeContext != "" {
				config, err := clientConfig.RawConfig()
				assert.NilError(t, err)
				assert.Assert(t, config.CurrentContext == tc.kubeContext)
				assert.Assert(t, config.Contexts[tc.kubeContext].Cluster == tc.kubeCluster)
			}

			if tc.kubeAsUser != "" {
				config, err := clientConfig.ClientConfig()
				assert.NilError(t, err)
				assert.Assert(t, config.Impersonate.UserName == tc.kubeAsUser)
			}

			if tc.kubeAsUID != "" {
				config, err := clientConfig.ClientConfig()
				assert.NilError(t, err)
				assert.Assert(t, config.Impersonate.UID == tc.kubeAsUID)
			}

			if len(tc.kubeAsGroup) > 0 {
				config, err := clientConfig.ClientConfig()
				assert.NilError(t, err)
				assert.Assert(t, len(config.Impersonate.Groups) == len(tc.kubeAsGroup))
				for i := range tc.kubeAsGroup {
					assert.Assert(t, config.Impersonate.Groups[i] == tc.kubeAsGroup[i])
				}
			}
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
			assert.Assert(t, sourcesClient.ContainerSourcesClient().Namespace() == namespace)
			assert.Assert(t, sourcesClient.APIServerSourcesClient().Namespace() == namespace)
		}
	}
}

func TestNewSourcesV1beta2Client(t *testing.T) {
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

		sourcesClient, err := p.newSourcesClientV1beta2(namespace)

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
			assert.Assert(t, sourcesClient.PingSourcesClient().Namespace() == namespace)
		}
	}
}

func TestNewServingV1beta1Clients(t *testing.T) {
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

		servingV1beta1Client, err := p.newServingClientV1beta1(namespace)

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

		if servingV1beta1Client != nil {
			assert.Assert(t, servingV1beta1Client.Namespace() == namespace)
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

func TestNewMessagingClient(t *testing.T) {
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

		msgClient, err := p.newMessagingClient(namespace)

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

		if msgClient != nil {
			assert.Assert(t, msgClient.ChannelsClient().Namespace() == namespace)
		}
	}
}

func TestInitialize(t *testing.T) {
	params := &KnParams{}
	params.Initialize()
	assert.Assert(t, params.NewServingClient != nil)
	assert.Assert(t, params.NewServingV1beta1Client != nil)
	assert.Assert(t, params.NewGitopsServingClient != nil)
	assert.Assert(t, params.NewSourcesClient != nil)
	assert.Assert(t, params.NewEventingClient != nil)
	assert.Assert(t, params.NewMessagingClient != nil)
	assert.Assert(t, params.NewDynamicClient != nil)
	assert.Assert(t, params.NewEventingV1beta2Client != nil)

	basic, err := clientcmd.NewClientConfigFromBytes([]byte(BASIC_KUBECONFIG))
	if err != nil {
		t.Error(err)
	}

	// Test all clients are not nil
	params.ClientConfig = basic
	servingClient, err := params.NewServingClient("mockNamespace")
	assert.NilError(t, err)
	assert.Assert(t, servingClient != nil)

	eventingClient, err := params.NewEventingClient("mockNamespace")
	assert.NilError(t, err)
	assert.Assert(t, eventingClient != nil)

	gitOpsClient, err := params.NewGitopsServingClient("mockNamespace", "mockDir")
	assert.NilError(t, err)
	assert.Assert(t, gitOpsClient != nil)

	messagingClient, err := params.NewMessagingClient("mockNamespace")
	assert.NilError(t, err)
	assert.Assert(t, messagingClient != nil)

	sourcesClient, err := params.NewSourcesClient("mockNamespace")
	assert.NilError(t, err)
	assert.Assert(t, sourcesClient != nil)

	eventingBeta1Client, err := params.NewEventingV1beta2Client("mockNamespace")
	assert.NilError(t, err)
	assert.Assert(t, eventingBeta1Client != nil)
}
