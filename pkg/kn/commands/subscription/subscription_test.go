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

package subscription

import (
	"bytes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	kndynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"
	clientv1 "knative.dev/client/pkg/messaging/v1"
)

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

func executeSubscriptionCommand(subscriptionClient clientv1.KnSubscriptionsClient, dynamicClient kndynamic.KnDynamicClient, args ...string) (string, error) {
	knParams := &commands.KnParams{}
	knParams.ClientConfig = blankConfig

	output := new(bytes.Buffer)
	knParams.Output = output
	knParams.NewDynamicClient = func(namespace string) (kndynamic.KnDynamicClient, error) {
		return dynamicClient, nil
	}

	cmd := NewSubscriptionCommand(knParams)
	cmd.SetArgs(args)
	cmd.SetOutput(output)

	subscriptionClientFactory = func(config clientcmd.ClientConfig, namespace string) (clientv1.KnSubscriptionsClient, error) {
		return subscriptionClient, nil
	}
	defer cleanupSubscriptionMockClient()

	err := cmd.Execute()

	return output.String(), err
}

func cleanupSubscriptionMockClient() {
	subscriptionClientFactory = nil
}

func createSubscription(name, channel, subscriber, reply, dls string) *messagingv1.Subscription {
	return clientv1.
		NewSubscriptionBuilder(name).
		Channel(createIMCObjectReference(channel)).
		Subscriber(createServiceSink(subscriber)).
		Reply(createBrokerSink(reply)).
		DeadLetterSink(createBrokerSink(dls)).
		Build()
}

func createIMCObjectReference(channel string) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		APIVersion: "messaging.knative.dev/v1beta1",
		Kind:       "InMemoryChannel",
		Name:       channel,
	}
}

func createServiceSink(service string) *duckv1.Destination {
	if service == "" {
		return nil
	}
	return &duckv1.Destination{
		Ref: &duckv1.KReference{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
			Name:       service,
			Namespace:  "default",
		},
	}
}

func createBrokerSink(broker string) *duckv1.Destination {
	if broker == "" {
		return nil
	}
	return &duckv1.Destination{
		Ref: &duckv1.KReference{
			Kind:       "Broker",
			APIVersion: "eventing.knative.dev/v1",
			Name:       broker,
			Namespace:  "default",
		},
	}
}

func createService(name string) *servingv1.Service {
	return &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
	}
}

func createBroker(name string) *eventingv1.Broker {
	return &eventingv1.Broker{
		TypeMeta:   metav1.TypeMeta{Kind: "Broker", APIVersion: "eventing.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
	}
}
