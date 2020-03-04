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

package cronjob

import (
	"bytes"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
	"knative.dev/pkg/apis/duck/v1beta1"

	kn_dynamic "knative.dev/client/pkg/dynamic"
	source_client_v1alpha1 "knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
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

func executeCronJobSourceCommand(cronJobSourceClient source_client_v1alpha1.KnCronJobSourcesClient, dynamicClient kn_dynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kn_dynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewCronJobCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	cronJobSourceClientFactory = func(config clientcmd.ClientConfig, namespace string) (source_client_v1alpha1.KnCronJobSourcesClient, error) {
		return cronJobSourceClient, nil
	}
	defer cleanupCronJobMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupCronJobMockClient() {
	cronJobSourceClientFactory = nil
}

func createCronJobSource(name, schedule, data, service string, sa string, requestcpu string, requestmm string, limitcpu string, limitmm string) *v1alpha1.CronJobSource {
	sink := &v1beta1.Destination{
		Ref: &corev1.ObjectReference{Name: service, Kind: "Service", APIVersion: "serving.knative.dev/v1", Namespace: "default"},
	}
	return source_client_v1alpha1.NewCronJobSourceBuilder(name).
		Schedule(schedule).
		Data(data).
		Sink(sink).
		ResourceRequestsCPU(requestcpu).
		ResourceRequestsMemory(requestmm).
		ResourceLimitsCPU(limitcpu).
		ResourceLimitsMemory(limitmm).
		ServiceAccount(sa).
		Build()
}
