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
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	"knative.dev/pkg/apis/duck/v1beta1"

	source_client_v1alpha1 "knative.dev/client/pkg/eventing/sources/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	serving_client_v1alpha1 "knative.dev/client/pkg/serving/v1alpha1"
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

func executeCronJobSourceCommand(cronJobSourceClient source_client_v1alpha1.KnCronJobSourcesClient, servingClient serving_client_v1alpha1.KnServingClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewServingClient = func(namespace string) (serving_client_v1alpha1.KnServingClient, error) {
		return servingClient, nil
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

func createCronJobSource(name, schedule, data, service string) *v1alpha1.CronJobSource {
	sink := &v1beta1.Destination{
		Ref: &corev1.ObjectReference{Name: service, Kind: "Service"},
	}
	return source_client_v1alpha1.NewCronJobSourceBuilder(name).Schedule(schedule).Data(data).Sink(sink).Build()
}
