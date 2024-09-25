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

package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

// ErrCantFindConfigFile is returned when given config file can't be found.
var ErrCantFindConfigFile = errors.New("can not find config file")

// Params contain Kubernetes specific params, that CLI should comply to.
type Params struct {
	KubeCfgPath string
	KubeContext string
	KubeCluster string
	KubeAsUser  string
	KubeAsUID   string
	KubeAsGroup []string
}

// SetFlags is used set flags to the given flagset.
func (kp *Params) SetFlags(flags *pflag.FlagSet) {
	flags.StringVar(&kp.KubeCfgPath, "kubeconfig", "",
		"kubectl configuration file (default: ~/.kube/config)")
	flags.StringVar(&kp.KubeContext, "context", "",
		"name of the kubeconfig context to use")
	flags.StringVar(&kp.KubeCluster, "cluster", "",
		"name of the kubeconfig cluster to use")
	flags.StringVar(&kp.KubeAsUser, "as", "",
		"username to impersonate for the operation")
	flags.StringVar(&kp.KubeAsUID, "as-uid", "",
		"uid to impersonate for the operation")
	flags.StringArrayVar(&kp.KubeAsGroup, "as-group",
		[]string{}, "group to impersonate for the operation, this flag can "+
			"be repeated to specify multiple groups")
}

// GetClientConfig gets ClientConfig from Kube' configuration params.
func (kp *Params) GetClientConfig() (clientcmd.ClientConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if kp.KubeContext != "" {
		configOverrides.CurrentContext = kp.KubeContext
	}
	if kp.KubeCluster != "" {
		configOverrides.Context.Cluster = kp.KubeCluster
	}
	if kp.KubeAsUser != "" {
		configOverrides.AuthInfo.Impersonate = kp.KubeAsUser
	}
	if kp.KubeAsUID != "" {
		configOverrides.AuthInfo.ImpersonateUID = kp.KubeAsUID
	}
	if len(kp.KubeAsGroup) > 0 {
		configOverrides.AuthInfo.ImpersonateGroups = kp.KubeAsGroup
	}
	if len(kp.KubeCfgPath) == 0 {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
	}

	_, err := os.Stat(kp.KubeCfgPath)
	if err == nil {
		loadingRules.ExplicitPath = kp.KubeCfgPath
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides), nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	paths := filepath.SplitList(kp.KubeCfgPath)
	if len(paths) > 1 {
		return nil, fmt.Errorf("%w. '%s' looks "+
			"like a path. Please use the env var KUBECONFIG if you want to "+
			"check for multiple configuration files", ErrCantFindConfigFile, kp.KubeCfgPath)
	}
	return nil, fmt.Errorf("%w: '%s'", ErrCantFindConfigFile, kp.KubeCfgPath)
}
