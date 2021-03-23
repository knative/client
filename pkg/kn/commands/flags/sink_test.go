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

package flags

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	dynamicfake "knative.dev/client/pkg/dynamic/fake"
)

type resolveCase struct {
	sink        string
	destination *duckv1.Destination
	errContents string
}

type sinkFlagAddTestCases struct {
	flagName          string
	expectedFlagName  string
	expectedShortName string
}

func TestSinkFlagAdd(t *testing.T) {
	cases := []*sinkFlagAddTestCases{
		{
			"",
			"sink",
			"s",
		},
		{
			"subscriber",
			"subscriber",
			"",
		},
	}
	for _, tc := range cases {
		c := &cobra.Command{Use: "sinktest"}
		sinkFlags := new(SinkFlags)
		if tc.flagName == "" {
			sinkFlags.Add(c)
			assert.Equal(t, tc.expectedFlagName, c.Flag("sink").Name)
			assert.Equal(t, tc.expectedShortName, c.Flag("sink").Shorthand)
		} else {
			sinkFlags.AddWithFlagName(c, tc.flagName, "")
			assert.Equal(t, tc.expectedFlagName, c.Flag(tc.flagName).Name)
			assert.Equal(t, tc.expectedShortName, c.Flag(tc.flagName).Shorthand)
		}
	}
}

func TestResolve(t *testing.T) {
	targetExampleCom, err := apis.ParseURL("http://target.example.com")
	assert.NilError(t, err)

	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	defaultBroker := &eventingv1beta1.Broker{
		TypeMeta:   metav1.TypeMeta{Kind: "Broker", APIVersion: "eventing.knative.dev/v1beta1"},
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "default"},
	}
	pipeChannel := &messagingv1beta1.Channel{
		TypeMeta:   metav1.TypeMeta{Kind: "Channel", APIVersion: "messaging.knative.dev/v1beta1"},
		ObjectMeta: metav1.ObjectMeta{Name: "pipe", Namespace: "default"},
	}

	cases := []resolveCase{
		{"ksvc:mysvc", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Service",
				APIVersion: "serving.knative.dev/v1",
				Namespace:  "default",
				Name:       "mysvc"}}, ""},
		{"mysvc", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Service",
				APIVersion: "serving.knative.dev/v1",
				Namespace:  "default",
				Name:       "mysvc"}}, ""},
		{"ksvc:absent", nil, "\"absent\" not found"},
		{"broker:default", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Broker",
				APIVersion: "eventing.knative.dev/v1beta1",
				Namespace:  "default",
				Name:       "default"}}, ""},
		{"channel:pipe",
			&duckv1.Destination{
				Ref: &duckv1.KReference{Kind: "Channel",
					APIVersion: "messaging.knative.dev/v1beta1",
					Namespace:  "default",
					Name:       "pipe",
				},
			},
			""},

		{"http://target.example.com", &duckv1.Destination{
			URI: targetExampleCom,
		}, ""},
		{"k8ssvc:foo", nil, "unsupported sink prefix: 'k8ssvc'"},
		{"svc:foo", nil, "please use prefix 'ksvc' for knative service"},
		{"service:foo", nil, "please use prefix 'ksvc' for knative service"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("default", mysvc, defaultBroker, pipeChannel)
	for _, c := range cases {
		i := &SinkFlags{c.sink}
		result, err := i.ResolveSink(context.Background(), dynamicClient, "default")
		if c.destination != nil {
			assert.DeepEqual(t, result, c.destination)
			assert.NilError(t, err)
		} else {
			assert.ErrorContains(t, err, c.errContents)
		}
	}
}

func TestResolveWithNamespace(t *testing.T) {
	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "my-namespace"},
	}
	defaultBroker := &eventingv1beta1.Broker{
		TypeMeta:   metav1.TypeMeta{Kind: "Broker", APIVersion: "eventing.knative.dev/v1beta1"},
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "my-namespace"},
	}
	pipeChannel := &messagingv1beta1.Channel{
		TypeMeta:   metav1.TypeMeta{Kind: "Channel", APIVersion: "messaging.knative.dev/v1beta1"},
		ObjectMeta: metav1.ObjectMeta{Name: "pipe", Namespace: "my-namespace"},
	}

	cases := []resolveCase{
		{"ksvc:mysvc:my-namespace", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Service",
				APIVersion: "serving.knative.dev/v1",
				Namespace:  "my-namespace",
				Name:       "mysvc"}}, ""},
		{"broker:default:my-namespace", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Broker",
				APIVersion: "eventing.knative.dev/v1beta1",
				Namespace:  "my-namespace",
				Name:       "default"}}, ""},
		{"channel:pipe:my-namespace", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Channel",
				APIVersion: "messaging.knative.dev/v1beta1",
				Namespace:  "my-namespace",
				Name:       "pipe"}}, ""},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("my-namespace", mysvc, defaultBroker, pipeChannel)
	for _, c := range cases {
		i := &SinkFlags{c.sink}
		result, err := i.ResolveSink(context.Background(), dynamicClient, "default")
		if c.destination != nil {
			assert.DeepEqual(t, result, c.destination)
			assert.Equal(t, c.destination.Ref.Namespace, "my-namespace")
			assert.NilError(t, err)
		} else {
			assert.ErrorContains(t, err, c.errContents)
		}
	}
}
