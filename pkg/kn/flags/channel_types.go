// Copyright © 2020 The Knative Authors
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
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/client/pkg/kn/config"
)

type ChannelTypeFlags struct {
	ctype string
}

type ChannelRef struct {
	Cref string
}

// ctypeMappings maps aliases used for channel types to their GroupVersionKind
var ctypeMappings = map[string]schema.GroupVersionKind{
	"imcv1beta1": {
		Group:   "messaging.knative.dev",
		Version: "v1beta1",
		Kind:    "InMemoryChannel",
	},
	"imc": {
		Group:   "messaging.knative.dev",
		Version: "v1",
		Kind:    "InMemoryChannel",
	},
}

// Add sets channel type flag definition to given flagset
func (i *ChannelTypeFlags) Add(f *pflag.FlagSet) {
	f.StringVar(&i.ctype,
		"type",
		"",
		"Override channel type to create, in the format '--type Group:Version:Kind'. "+
			"If flag is not specified, it uses default messaging layer settings for channel type, cluster wide or specific namespace. "+
			"You can configure aliases for channel types in kn config and refer the aliases with this flag. "+
			"You can also refer inbuilt channel type InMemoryChannel using an alias 'imc' like '--type imc'. "+
			"Examples: '--type messaging.knative.dev:v1alpha1:KafkaChannel' for specifying explicit Group:Version:Kind.")

	for _, p := range config.GlobalConfig.ChannelTypeMappings() {
		//user configration might override the default configuration
		ctypeMappings[p.Alias] = schema.GroupVersionKind{
			Kind:    p.Kind,
			Group:   p.Group,
			Version: p.Version,
		}
	}
}

// Parse parses the CLI value for channel type flag and populates GVK or returns error
func (i *ChannelTypeFlags) Parse() (*schema.GroupVersionKind, error) {
	parts := strings.Split(i.ctype, ":")
	switch len(parts) {
	case 1:
		if typ, ok := ctypeMappings[i.ctype]; ok {
			return &typ, nil
		}
		return nil, fmt.Errorf("Error: unknown channel type alias: '%s'", i.ctype)
	case 3:
		if parts[0] == "" || parts[1] == "" || parts[2] == "" {
			return nil, fmt.Errorf("Error: incorrect value '%s' for '--type', must be in the format 'Group:Version:Kind' or configure an alias in kn config", i.ctype)
		}
		return &schema.GroupVersionKind{Group: parts[0], Version: parts[1], Kind: parts[2]}, nil
	default:
		return nil, fmt.Errorf("Error: incorrect value '%s' for '--type', must be in the format 'Group:Version:Kind' or configure an alias in kn config", i.ctype)
	}
}

// Add sets channel reference flag definition to given flagset
func (i *ChannelRef) Add(f *pflag.FlagSet) {
	f.StringVar(&i.Cref,
		"channel",
		"",
		"Specify the channel to subscribe to, in the format '--channel Group:Version:Kind:Name'. "+
			"You can use channel type aliases from kn config with this flag. "+
			"You can also refer inbuilt channel type InMemoryChannel using alias 'imc' like '--type imc:CHANNEL_NAME'. "+
			"Examples: '--channel messaging.knative.dev:v1alpha1:KafkaChannel:k1' for specifying explicit Group:Version:Kind:Name.")
}

// Parse parses the CLI value for channel ref flag and populates object reference or return error
func (i *ChannelRef) Parse() (*corev1.ObjectReference, error) {
	parts := strings.Split(i.Cref, ":")
	switch len(parts) {
	// TODO: Check if we want an implicit assumed channel type to be of IMC
	// case 1:
	case 2:
		if typ, ok := ctypeMappings[parts[0]]; ok {
			return &corev1.ObjectReference{Kind: typ.Kind, APIVersion: typ.GroupVersion().String(), Name: parts[1]}, nil
		}
	case 4:
		if parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
			return nil, fmt.Errorf("Error: incorrect value '%s' for '--channel', must be in the format 'Group:Version:Kind:Name' or configure an alias in kn config and refer as: '--channel ALIAS:NAME'", i.Cref)
		}
		return &corev1.ObjectReference{Kind: parts[2], APIVersion: parts[0] + "/" + parts[1], Name: parts[3]}, nil
	default:
		return nil, fmt.Errorf("Error: incorrect value '%s' for '--channel', must be in the format 'Group:Version:Kind:Name' or configure an alias in kn config and refer as: '--channel ALIAS:NAME'", i.Cref)
	}
	return nil, nil
}
