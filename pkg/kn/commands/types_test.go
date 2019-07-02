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
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
)

type typeTestCase struct {
	kubeCfgPath   string
	explicitPath  string
	expectedError error
}

func TestGetClientConfig(t *testing.T) {
	multiConfigs := fmt.Sprintf("%s%s%s", "/testing/assets/kube-config-01.yml", string(os.PathListSeparator), "/testing/assets/kube-config-02.yml")

	for i, tc := range []typeTestCase{
		{
			"",
			clientcmd.NewDefaultClientConfigLoadingRules().ExplicitPath,
			nil,
		},
		{
			"/testing/assets/kube-config-01.yml",
			"/testing/assets/kube-config-01.yml",
			nil,
		},
		{
			multiConfigs,
			"",
			errors.New(fmt.Sprintf("Config file '%s' could not be found. For configuration lookup path please use the env variable KUBECONFIG", multiConfigs)),
		},
	} {
		p := &KnParams{
			KubeCfgPath: tc.kubeCfgPath,
		}

		clientConfig, err := p.GetClientConfig()

		if !reflect.DeepEqual(err, tc.expectedError) {
			t.Errorf("%d: wrong error detected: %s (expected) != %s (actual)", i, tc.expectedError.Error(), err.Error())
		}

		if clientConfig != nil {
			configAccess := clientConfig.ConfigAccess()

			if configAccess.GetExplicitFile() != tc.explicitPath {
				t.Errorf("%d: wrong explicit file detected: %s (expected) != %s (actual)", i, tc.explicitPath, configAccess.GetExplicitFile())
			}
		}
	}
}
