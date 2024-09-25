/*
 Copyright 2024 The Knative Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package k8s_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client/pkg/k8s"
	"knative.dev/client/pkg/util"
	"knative.dev/client/pkg/util/test"
)

func TestGetClientConfig(t *testing.T) {
	multiConfigs := fmt.Sprintf("%s%s%s",
		"/testing/assets/kube-config-01.yml",
		string(os.PathListSeparator),
		"/testing/assets/kube-config-02.yml")

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "mock")
	assert.NilError(t, os.WriteFile(tempFile, []byte(basicKubeconfig), test.FileModeReadWrite))

	for _, tc := range []typeTestCase{{
		"",
		"",
		"",
		"",
		[]string{},
		"",
		clientcmd.NewDefaultClientConfigLoadingRules().ExplicitPath,
		"",
	}, {
		tempFile,
		"",
		"",
		"",
		[]string{},
		"",
		tempFile,
		"",
	}, {
		"/testing/assets/kube-config-01.yml",
		"foo",
		"",
		"",
		[]string{},
		"bar",
		"",
		fmt.Sprintf("can not find config file: '%s'", "/testing/assets/kube-config-01.yml"),
	}, {
		multiConfigs,
		"",
		"",
		"",
		[]string{},
		"",
		"",
		fmt.Sprintf("can not find config file. '%s' looks like a path. Please use the env var KUBECONFIG if you want to check for multiple configuration files", multiConfigs),
	}, {
		tempFile,
		"",
		"admin",
		"",
		[]string{},
		"",
		tempFile,
		"",
	}, {
		tempFile,
		"",
		"admin",
		"",
		[]string{"system:authenticated", "system:masters"},
		"",
		tempFile,
		"",
	}, {
		tempFile,
		"",
		"admin",
		"abc123",
		[]string{},
		"",
		tempFile,
		"",
	}} {
		p := &k8s.Params{
			KubeCfgPath: tc.kubeCfgPath,
			KubeContext: tc.kubeContext,
			KubeAsUser:  tc.kubeAsUser,
			KubeAsUID:   tc.kubeAsUID,
			KubeAsGroup: tc.kubeAsGroup,
			KubeCluster: tc.kubeCluster,
		}
		var clientConfig clientcmd.ClientConfig
		{
			cc, err := p.GetClientConfig()
			if tc.expectedError != "" {
				assert.Assert(t, util.ContainsAll(err.Error(), tc.expectedError))
			} else {
				assert.Assert(t, err == nil, err)
			}
			clientConfig = cc
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

var basicKubeconfig = `apiVersion: v1
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
