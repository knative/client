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

package flags_test

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/client/pkg/commands/flags"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	"knative.dev/eventing/pkg/apis/sources/v1beta2"
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
	cases := []sinkFlagAddTestCases{{
		"",
		"sink",
		"s",
	}, {
		"subscriber",
		"subscriber",
		"",
	}}
	for _, tc := range cases {
		t.Run(tc.flagName, func(t *testing.T) {
			c := &cobra.Command{Use: "sinktest"}
			sinkFlags := flags.SinkFlags{}
			if tc.flagName == "" {
				sinkFlags.Add(c)
				assert.Equal(t, tc.expectedFlagName, c.Flag("sink").Name)
				assert.Equal(t, tc.expectedShortName, c.Flag("sink").Shorthand)
			} else {
				sinkFlags.AddWithFlagName(c, tc.flagName, "")
				assert.Equal(t, tc.expectedFlagName, c.Flag(tc.flagName).Name)
				assert.Equal(t, tc.expectedShortName, c.Flag(tc.flagName).Shorthand)
			}
		})
	}
}

func TestResolve(t *testing.T) {
	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "default"},
	}
	defaultBroker := &eventingv1.Broker{
		TypeMeta:   metav1.TypeMeta{Kind: "Broker", APIVersion: "eventing.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "default"},
	}
	pipeChannel := &messagingv1.Channel{
		TypeMeta:   metav1.TypeMeta{Kind: "Channel", APIVersion: "messaging.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "pipe", Namespace: "default"},
	}
	pingSource := &v1beta2.PingSource{
		TypeMeta:   metav1.TypeMeta{Kind: "PingSource", APIVersion: "sources.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
	}
	k8sService := &corev1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
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
				APIVersion: "eventing.knative.dev/v1",
				Namespace:  "default",
				Name:       "default"}}, ""},
		{"channel:pipe",
			&duckv1.Destination{
				Ref: &duckv1.KReference{Kind: "Channel",
					APIVersion: "messaging.knative.dev/v1",
					Namespace:  "default",
					Name:       "pipe",
				},
			},
			""},

		{"sources.knative.dev/v1/pingsource:foo", &duckv1.Destination{Ref: &duckv1.KReference{
			APIVersion: "sources.knative.dev/v1",
			Kind:       "PingSource",
			Namespace:  "default",
			Name:       "foo",
		}}, ""},
		{"sources.knative.dev/v1/pingsources:foo", &duckv1.Destination{Ref: &duckv1.KReference{
			APIVersion: "sources.knative.dev/v1",
			Kind:       "PingSource",
			Namespace:  "default",
			Name:       "foo",
		}}, ""},
		{"sources.knative.dev/v1/Pingsource:foo", &duckv1.Destination{Ref: &duckv1.KReference{
			APIVersion: "sources.knative.dev/v1",
			Kind:       "PingSource",
			Namespace:  "default",
			Name:       "foo",
		}}, ""},
		{"sources.knative.dev/v1/PingSources:foo", &duckv1.Destination{Ref: &duckv1.KReference{
			APIVersion: "sources.knative.dev/v1",
			Kind:       "PingSource",
			Namespace:  "default",
			Name:       "foo",
		}}, ""},
		{"http://target.example.com", &duckv1.Destination{
			URI: url(t, "http://target.example.com"),
		}, ""},
		{"k8ssvc:foo", nil, "k8ssvcs \"foo\" not found"},
		{"svc:foo", &duckv1.Destination{Ref: &duckv1.KReference{
			APIVersion: "v1",
			Kind:       "Service",
			Namespace:  "default",
			Name:       "foo",
		}}, ""},
		{"service:foo", &duckv1.Destination{Ref: &duckv1.KReference{
			APIVersion: "v1",
			Kind:       "Service",
			Namespace:  "default",
			Name:       "foo",
		}}, ""},
		{"absent:foo", nil, "absents \"foo\" not found"},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient(
		"default",
		mysvc, defaultBroker, pipeChannel, pingSource, k8sService,
	)

	for _, c := range cases {
		t.Run(c.sink, func(t *testing.T) {
			sf := &flags.SinkFlags{Sink: c.sink}
			result, err := sf.ResolveSink(context.Background(), dynamicClient, "default")
			if c.destination != nil {
				assert.DeepEqual(t, result, c.destination)
				assert.NilError(t, err)
				assert.Equal(t, c.errContents, "")
			} else {
				assert.ErrorContains(t, err, c.errContents)
			}
		})
	}
}

func TestResolveWithNamespace(t *testing.T) {
	mysvc := &servingv1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "serving.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "mysvc", Namespace: "my-namespace"},
	}
	defaultBroker := &eventingv1.Broker{
		TypeMeta:   metav1.TypeMeta{Kind: "Broker", APIVersion: "eventing.knative.dev/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "default", Namespace: "my-namespace"},
	}
	pipeChannel := &messagingv1.Channel{
		TypeMeta:   metav1.TypeMeta{Kind: "Channel", APIVersion: "messaging.knative.dev/v1"},
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
				APIVersion: "eventing.knative.dev/v1",
				Namespace:  "my-namespace",
				Name:       "default"}}, ""},
		{"channel:pipe:my-namespace", &duckv1.Destination{
			Ref: &duckv1.KReference{Kind: "Channel",
				APIVersion: "messaging.knative.dev/v1",
				Namespace:  "my-namespace",
				Name:       "pipe"}}, ""},
	}
	dynamicClient := dynamicfake.CreateFakeKnDynamicClient("my-namespace", mysvc, defaultBroker, pipeChannel)
	for _, c := range cases {
		t.Run(c.sink, func(t *testing.T) {
			i := &flags.SinkFlags{Sink: c.sink}
			result, err := i.ResolveSink(context.Background(), dynamicClient, "default")
			if c.destination != nil {
				assert.DeepEqual(t, result, c.destination)
				assert.Equal(t, c.destination.Ref.Namespace, "my-namespace")
				assert.NilError(t, err)
			} else {
				assert.ErrorContains(t, err, c.errContents)
			}
		})
	}
}

func TestSinkToString(t *testing.T) {
	tcs := []resolveCase{{
		sink: "ksvc:mysvc",
		destination: &duckv1.Destination{Ref: &duckv1.KReference{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
			Namespace:  "my-namespace",
			Name:       "mysvc",
		}},
	}, {
		sink: "broker:default",
		destination: &duckv1.Destination{Ref: &duckv1.KReference{
			Kind:       "Broker",
			APIVersion: "eventing.knative.dev/v1",
			Namespace:  "my-namespace",
			Name:       "default",
		}},
	}, {
		sink: "svc:mysvc",
		destination: &duckv1.Destination{Ref: &duckv1.KReference{
			Kind:       "Service",
			APIVersion: "v1",
			Namespace:  "my-namespace",
			Name:       "mysvc",
		}},
	}, {
		sink: "things.acme.dev/v1alpha1:abc",
		destination: &duckv1.Destination{Ref: &duckv1.KReference{
			Kind:       "Thing",
			APIVersion: "acme.dev/v1alpha1",
			Namespace:  "my-namespace",
			Name:       "abc",
		}},
	}, {
		sink: "http://target.example.com",
		destination: &duckv1.Destination{
			URI: url(t, "http://target.example.com"),
		},
	}, {
		sink:        "",
		destination: &duckv1.Destination{},
	}}
	for _, tc := range tcs {
		t.Run(tc.sink, func(t *testing.T) {
			got := flags.SinkToString(*tc.destination)
			assert.Equal(t, got, tc.sink)
		})
	}
}

func url(t testing.TB, uri string) *apis.URL {
	t.Helper()
	u, err := apis.ParseURL(uri)
	assert.NilError(t, err)
	return u
}
