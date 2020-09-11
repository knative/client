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
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type channelTypeFlagsTestCase struct {
	name            string
	arg             string
	expectedGVK     *schema.GroupVersionKind
	expectedErrText string
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
