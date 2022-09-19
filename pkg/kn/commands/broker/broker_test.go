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
package broker

import (
	"bytes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/client/pkg/dynamic"
	dynamicfake "knative.dev/client/pkg/dynamic/fake"
	v1 "knative.dev/eventing/pkg/apis/duck/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	clientv1beta1 "knative.dev/client/pkg/eventing/v1"
	"knative.dev/client/pkg/kn/commands"
	v1beta1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

// Helper methods
var blankConfig clientcmd.ClientConfig

const (
	testSvc     = "test-svc"
	testTimeout = "PT10S"
	testRetry   = int32(5)
)

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

func executeBrokerCommand(brokerClient clientv1beta1.KnEventingClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output

	knParams.NewEventingClient = func(namespace string) (clientv1beta1.KnEventingClient, error) {
		return brokerClient, nil
	}

	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: testSvc, Namespace: "default"},
	}
	knParams.NewDynamicClient = func(namespace string) (dynamic.KnDynamicClient, error) {
		return dynamicfake.CreateFakeKnDynamicClient("default", mysvc), nil
	}

	cmd := NewBrokerCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	err := cmd.Execute()

	return output.String(), err
}

func createBroker(brokerName string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").Build()
}

func createBrokerWithGvk(brokerName string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").WithGvk().Build()
}

func createBrokerWithNamespace(brokerName, namespace string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace(namespace).Build()
}

func createBrokerWithClass(brokerName, class string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").Class(class).Build()
}

func createBrokerWithDlSink(brokerName, service string) *v1beta1.Broker {
	sink := &duckv1.Destination{
		Ref: &duckv1.KReference{Name: service, Kind: "Service", APIVersion: "serving.knative.dev/v1", Namespace: "default"},
	}
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").DlSink(sink).Build()
}

func createBrokerWithTimeout(brokerName, timeout string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").Timeout(&timeout).Build()
}

func createBrokerWithRetry(brokerName string, retry int32) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").Retry(&retry).Build()
}

func createBrokerWithBackoffPolicy(brokerName, policy string) *v1beta1.Broker {
	boPolicy := v1.BackoffPolicyType(policy)
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").BackoffPolicy(&boPolicy).Build()
}

func createBrokerWithBackoffDelay(brokerName, delay string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").BackoffDelay(&delay).Build()
}

func createBrokerWithRetryAfterMax(brokerName, timeout string) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").RetryAfterMax(&timeout).Build()
}

func createBrokerWithConfig(brokerName string, config *duckv1.KReference) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").Class("Kafka").Config(config).Build()
}

func createBrokerWithConfigAndClass(brokerName, class string, config *duckv1.KReference) *v1beta1.Broker {
	return clientv1beta1.NewBrokerBuilder(brokerName).Namespace("default").Class(class).Config(config).Build()
}
