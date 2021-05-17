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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

// AddNamespaceFlags adds the namespace-related flags:
// * --namespace
// * --all-namespaces
func AddNamespaceFlags(flags *pflag.FlagSet, allowAll bool) {
	flags.StringP(
		"namespace",
		"n",
		"",
		"Specify the namespace to operate in.",
	)

	if allowAll {
		flags.BoolP(
			"all-namespaces",
			"A",
			false,
			"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.",
		)
	}
}

// GetNamespace returns namespace from command specified by flag
func (params *KnParams) GetNamespace(cmd *cobra.Command) (string, error) {
	namespace := cmd.Flag("namespace").Value.String()
	// check value of all-namespaces only if its defined
	if cmd.Flags().Lookup("all-namespaces") != nil {
		all, err := cmd.Flags().GetBool("all-namespaces")
		if err != nil {
			return "", err
		} else if all { // if all-namespaces=True
			// namespace = "" <-- all-namespaces representation
			return "", nil
		}
	}
	// if all-namespaces=False or namespace not given, use default namespace
	if namespace == "" {
		var err error
		namespace, err = params.CurrentNamespace()
		if err != nil {
			if !clientcmd.IsEmptyConfig(err) {
				return "", err
			}
			// If no current namespace is set use "default"
			namespace = "default"
		}
	}
	return namespace, nil
}

// CurrentNamespace returns the current namespace which is either provided as option or picked up from kubeconfig
func (params *KnParams) CurrentNamespace() (string, error) {
	var err error
	if params.fixedCurrentNamespace != "" {
		return params.fixedCurrentNamespace, nil
	}
	if params.ClientConfig == nil {
		params.ClientConfig, err = params.GetClientConfig()
		if err != nil {
			return "", err
		}
	}
	name, _, err := params.ClientConfig.Namespace()
	return name, err
}
