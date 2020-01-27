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

package binding

import (
	"bytes"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	v1 "knative.dev/pkg/apis/duck/v1"

	kn_dynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"
	cl_sources_v1alpha1 "knative.dev/client/pkg/sources/v1alpha1"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

// TOOD: Remove that blankConfig hack for tests in favor of overwriting GetConfig()
// Remove also in service_test.go
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

func executeSinkBindingCommand(sinkBindingClient cl_sources_v1alpha1.KnSinkBindingClient, dynamicClient kn_dynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kn_dynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewBindingCreateCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	sinkBindingClientFactory = func(config clientcmd.ClientConfig, namespace string) (cl_sources_v1alpha1.KnSinkBindingClient, error) {
		return sinkBindingClient, nil
	}
	defer cleanupSinkBindingClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupSinkBindingClient() {
	sinkBindingClientFactory = nil
}

func createSinkBinding(name, service string, subjectGvk schema.GroupVersionKind, subjectName string) *v1alpha1.SinkBinding {
	sink := v1.Destination{
		Ref: &corev1.ObjectReference{Name: service, Kind: "Service", Namespace: "default", APIVersion: "serving.knative.dev/v1alpha1"},
	}
	binding, _ := cl_sources_v1alpha1.NewSinkBindingBuilder(name).
		Sink(&sink).
		SubjectGVK(&subjectGvk).
		SubjectName(subjectName).
		Build()
	return binding
}
