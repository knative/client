/*
Copyright 2020 The Knative Authors

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

package container

import (
	"bytes"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	kndynamic "knative.dev/client/pkg/dynamic"
	clientv1alpha2 "knative.dev/client/pkg/sources/v1alpha2"

	"knative.dev/client/pkg/kn/commands"
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

func executeContainerSourceCommand(containerSourceClient clientv1alpha2.KnContainerSourcesClient, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kndynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewContainerCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	containerSourceClientFactory = func(config clientcmd.ClientConfig, namespace string) (clientv1alpha2.KnContainerSourcesClient, error) {
		return containerSourceClient, nil
	}
	defer cleanupContainerServerMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupContainerServerMockClient() {
	containerSourceClientFactory = nil
}

func createContainerSource(name, image string, sink duckv1.Destination) *v1alpha2.ContainerSource {
	return clientv1alpha2.NewContainerSourceBuilder(name).
		PodSpec(corev1.PodSpec{
			Containers: []corev1.Container{{
				Image: image,
				Resources: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{},
					Requests: corev1.ResourceList{},
				},
			}}}).
		Sink(sink).
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
