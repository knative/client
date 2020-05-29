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
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kndynamic "knative.dev/client/pkg/dynamic"
	clientv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"

	"knative.dev/client/pkg/kn/commands"
)

const testNamespace = "default"

// Helper methods
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

func executeAPIServerSourceCommand(apiServerSourceClient clientv1alpha2.KnAPIServerSourcesClient, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {
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

	apiServerSourceClientFactory = func(config clientcmd.ClientConfig, namespace string) (clientv1alpha2.KnAPIServerSourcesClient, error) {
		return apiServerSourceClient, nil
	}
	defer cleanupAPIServerMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupAPIServerMockClient() {
	apiServerSourceClientFactory = nil
}

func createAPIServerSource(name, resourceKind, resourceVersion, serviceAccount, mode, service string, ceOverrides map[string]string) *v1alpha2.ApiServerSource {
	resources := []v1alpha2.APIVersionKindSelector{{
		APIVersion: resourceVersion,
		Kind:       resourceKind,
	}}

	sink := duckv1.Destination{
		Ref: &duckv1.KReference{
			Kind:       "Service",
			Name:       service,
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "default",
		}}

	return clientv1alpha2.NewAPIServerSourceBuilder(name).
		Resources(resources).
		ServiceAccount(serviceAccount).
		EventMode(mode).
		Sink(sink).
		CloudEventOverrides(ceOverrides, []string{}).
		Build()
}
