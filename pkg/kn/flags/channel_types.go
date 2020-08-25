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
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/client/pkg/kn/config"
)

type ChannelTypeFlags struct {
	ctype string
}

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
