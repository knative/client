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

package ping

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"

	"k8s.io/client-go/tools/clientcmd"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"knative.dev/client/pkg/commands"
	kndynamic "knative.dev/client/pkg/dynamic"
	clientv1 "knative.dev/client/pkg/sources/v1"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

// TODO: Remove that blankConfig hack for tests in favor of overwriting GetConfig()
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

func TestPingBuilder(t *testing.T) {
	name := "mockName"
	schedule := "* * * * *"
	data := "mockData"
	dataBase64 := "mockDataBase64"
	sink := "mockService"
	ceOverrideMap := map[string]string{}
	ps := createPingSource(name, schedule, data, dataBase64, sink, ceOverrideMap)
	assert.Equal(t, name, ps.Name)
	assert.Equal(t, schedule, ps.Spec.Schedule)
	assert.Equal(t, data, ps.Spec.Data)
	assert.Equal(t, dataBase64, ps.Spec.DataBase64)
	assert.Equal(t, sink, ps.Spec.Sink.Ref.Name)
	assert.DeepEqual(t, ceOverrideMap, ps.Spec.CloudEventOverrides.Extensions)
}

func executePingSourceCommand(pingSourceClient clientv1.KnPingSourcesClient, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kndynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewPingCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	pingSourceClientFactory = func(config clientcmd.ClientConfig, namespace string) (clientv1.KnPingSourcesClient, error) {
		return pingSourceClient, nil
	}
	defer cleanupPingMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupPingMockClient() {
	pingSourceClientFactory = nil
}

func createPingSource(name, schedule, data, dataBase64, service string, ceOverridesMap map[string]string) *sourcesv1.PingSource {
	sink := &duckv1.Destination{
		Ref: &duckv1.KReference{Name: service, Kind: "Service", APIVersion: "serving.knative.dev/v1", Namespace: "default"},
	}
	return clientv1.NewPingSourceBuilder(name).
		Schedule(schedule).
		Data(data).
		DataBase64(dataBase64).
		Sink(*sink).
		CloudEventOverrides(ceOverridesMap, []string{}).
		Build()
}
