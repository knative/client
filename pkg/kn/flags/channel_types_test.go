// Copyright 2020 The Knative Authors
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
	"testing"

	"github.com/spf13/pflag"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type channelTypeFlagsTestCase struct {
	name            string
	arg             string
	expectedGVK     *schema.GroupVersionKind
	expectedErrText string
}

type channelRefFlagsTestCase struct {
	name              string
	arg               string
	expectedObjectRef *corev1.ObjectReference
	expectedErrText   string
}

func TestChannelTypesFlags(t *testing.T) {
	cases := []*channelTypeFlagsTestCase{
		{
			"inbuilt alias 'imcv1beta1' case",
			"imcv1beta1",
			&schema.GroupVersionKind{Group: "messaging.knative.dev", Kind: "InMemoryChannel", Version: "v1beta1"},
			"",
		},
		{
			"inbuilt alias 'imc' case",
			"imc",
			&schema.GroupVersionKind{Group: "messaging.knative.dev", Kind: "InMemoryChannel", Version: "v1"},
			"",
		},
		{
			"explicit GVK case",
			"messaging.knative.dev:v1alpha1:KafkaChannel",
			&schema.GroupVersionKind{Group: "messaging.knative.dev", Kind: "KafkaChannel", Version: "v1alpha1"},
			"",
		},
		{
			"error case unknown alias",
			"natss",
			nil,
			"Error: unknown channel type alias: 'natss'",
		},
		{
			"error case incorrect gvk format, missing version",
			"foo::bar",
			nil,
			"Error: incorrect value 'foo::bar' for '--type', must be in the format 'Group:Version:Kind' or configure an alias in kn config",
		},
		{
			"error case incorrect gvk format, additional field",
			"foo:bar:baz:bat",
			nil,
			"Error: incorrect value 'foo:bar:baz:bat' for '--type', must be in the format 'Group:Version:Kind' or configure an alias in kn config",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &ChannelTypeFlags{}
			flagset := &pflag.FlagSet{}
			f.Add(flagset)
			flagset.Set("type", c.arg)
			gvk, err := f.Parse()
			if c.expectedErrText != "" {
				assert.Equal(t, err.Error(), c.expectedErrText)
			} else {
				assert.Equal(t, *gvk, *c.expectedGVK)
			}
		})
	}
}

func TestChannelRefFlags(t *testing.T) {
	cases := []*channelRefFlagsTestCase{
		{
			"inbuilt alias imcv1beta1 case",
			"imcv1beta1:i1",
			&corev1.ObjectReference{APIVersion: "messaging.knative.dev/v1beta1", Kind: "InMemoryChannel", Name: "i1"},
			"",
		},
		{
			"inbuilt alias 'imc' case",
			"imc:i2",
			&corev1.ObjectReference{APIVersion: "messaging.knative.dev/v1", Kind: "InMemoryChannel", Name: "i2"},
			"",
		},
		{
			"explicit GVK case",
			"messaging.knative.dev:v1alpha1:KafkaChannel:k1",
			&corev1.ObjectReference{APIVersion: "messaging.knative.dev/v1alpha1", Kind: "KafkaChannel", Name: "k1"},
			"",
		},
		{
			"default channel type prefix case",
			"c1",
			&corev1.ObjectReference{APIVersion: "messaging.knative.dev/v1beta1", Kind: "Channel", Name: "c1"},
			"",
		},
		{
			"error case unknown alias",
			"natss:n1",
			nil,
			"Error: unknown alias 'natss' for '--channel', please configure the alias in kn config or specify in the format '--channel Group:Version:Kind:Name'",
		},
		{
			"error case incorrect gvk format, missing version",
			"foo::bar",
			nil,
			"Error: incorrect value 'foo::bar' for '--channel', must be in the format 'Group:Version:Kind:Name' or configure an alias in kn config and refer as: '--channel ALIAS:NAME'",
		},
		{
			"error case incorrect gvk format, additional field",
			"foo:bar::bat",
			nil,
			"Error: incorrect value 'foo:bar::bat' for '--channel', must be in the format 'Group:Version:Kind:Name' or configure an alias in kn config and refer as: '--channel ALIAS:NAME'",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &ChannelRef{}
			flagset := &pflag.FlagSet{}
			f.Add(flagset)
			flagset.Set("channel", c.arg)
			obj, err := f.Parse()
			if c.expectedErrText != "" {
				assert.Equal(t, err.Error(), c.expectedErrText)
			} else {
				assert.Equal(t, *obj, *c.expectedObjectRef)
			}
		})
	}
}
