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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	kn_dynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	knsource_v1alpha1 "knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
)

const testNamespace = "default"

// Helper methods
var blankConfig clientcmd.ClientConfig

// TOOD: Remove that blankConfig hack for tests in favor of overwriting GetConfig()
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

func executeAPIServerSourceCommand(apiServerSourceClient knsource_v1alpha1.KnAPIServerSourcesClient, dynamicClient kn_dynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kn_dynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewAPIServerCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	apiServerSourceClientFactory = func(config clientcmd.ClientConfig, namespace string) (knsource_v1alpha1.KnAPIServerSourcesClient, error) {
		return apiServerSourceClient, nil
	}
	defer cleanupAPIServerMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupAPIServerMockClient() {
	apiServerSourceClientFactory = nil
}

func createAPIServerSource(name, resourceKind, resourceVersion, serviceAccount, mode, service string, isController bool) *v1alpha1.ApiServerSource {
	resources := []v1alpha1.ApiServerResource{{
		APIVersion: resourceVersion,
		Kind:       resourceKind,
		Controller: isController,
	}}

	sink := &duckv1beta1.Destination{
		Ref: &corev1.ObjectReference{
			Kind:       "Service",
			Name:       service,
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "default",
		}}

	return knsource_v1alpha1.NewAPIServerSourceBuilder(name).
		Resources(resources).
		ServiceAccount(serviceAccount).
		Mode(mode).
		Sink(sink).
		Build()
}
