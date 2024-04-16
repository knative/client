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

package apiserver

import (
	"bytes"

	"k8s.io/client-go/tools/clientcmd"
	v1 "knative.dev/eventing/pkg/apis/sources/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kndynamic "knative.dev/client-pkg/pkg/dynamic"
	clientv1 "knative.dev/client-pkg/pkg/sources/v1"

	"knative.dev/client/pkg/commands"
)

var blankConfig clientcmd.ClientConfig

// TODO: Remove that blankConfig hack for tests in favor of overwriting GetConfig()
func init() {
	var err error
	blankConfig, err = clientcmd.NewClientConfigFromBytes([]byte(`kind: Config
version: v1
users:
- name: u
clusters:
- name: c
  cluster:
    server: example.com
contexts:
- name: x
  context:
    user: u
    cluster: c
current-context: x
`))
	if err != nil {
		panic(err)
	}
}

func executeAPIServerSourceCommand(apiServerSourceClient clientv1.KnAPIServerSourcesClient, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kndynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewAPIServerCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	apiServerSourceClientFactory = func(config clientcmd.ClientConfig, namespace string) (clientv1.KnAPIServerSourcesClient, error) {
		return apiServerSourceClient, nil
	}
	defer cleanupAPIServerMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupAPIServerMockClient() {
	apiServerSourceClientFactory = nil
}

func createAPIServerSource(name, serviceAccount, mode string, resourceKind, resourceVersion []string, ceOverrides map[string]string, sink duckv1.Destination) *v1.ApiServerSource {
	resources := make([]v1.APIVersionKindSelector, len(resourceKind))

	for i, r := range resourceKind {
		resources[i] = v1.APIVersionKindSelector{
			APIVersion: resourceVersion[i],
			Kind:       r,
		}
	}

	return clientv1.NewAPIServerSourceBuilder(name).
		Resources(resources).
		ServiceAccount(serviceAccount).
		EventMode(mode).
		Sink(sink).
		CloudEventOverrides(ceOverrides, []string{}).
		Build()
}

func createSinkv1(serviceName, namespace string) duckv1.Destination {
	return duckv1.Destination{
		Ref: &duckv1.KReference{
			Kind:       "Service",
			Name:       serviceName,
			APIVersion: "serving.knative.dev/v1",
			Namespace:  namespace,
		},
	}
}
