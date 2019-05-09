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
)

// TODO: default namespace should be same scope for the request
const defaultNamespace = "default"

// AddNamespaceFlags adds the namespace-related flags:
// * --namespace
// * --all-namespaces
func AddNamespaceFlags(flags *pflag.FlagSet, allowAll bool) {
	flags.StringP(
		"namespace",
		"n",
		"",
		"List the requested object(s) in given namespace.",
	)

	if allowAll {
		flags.Bool(
			"all-namespaces",
			false,
			"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.",
		)
	}
}

// GetNamespace returns namespace from command specified by flag
func GetNamespace(cmd *cobra.Command) (string, error) {
	namespace := cmd.Flag("namespace").Value.String()
	// check value of all-namepace only if its defined
	if cmd.Flags().Lookup("all-namespaces") != nil {
		all, err := cmd.Flags().GetBool("all-namespaces")
		if err != nil {
			return "", err
		} else if all { // if all-namespaces=True
			// namespace = "" <-- all-namespaces representation
			return "", nil
		}
	}
	// if all-namepaces=False or namespace not given, use default namespace
	if namespace == "" {
		namespace = defaultNamespace
	}
	return namespace, nil
}
