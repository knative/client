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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	clientdynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"
)

type SinkFlags struct {
	sink string
}

func (i *SinkFlags) Add(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&i.sink, "sink", "s", "", "Addressable sink for events")
}

// SinkPrefixes maps prefixes used for sinks to their GroupVersionResources.
var SinkPrefixes = map[string]schema.GroupVersionResource{
	"broker": {
		Resource: "brokers",
		Group:    "eventing.knative.dev",
		Version:  "v1alpha1",
	},
	"service": {
		Resource: "services",
		Group:    "serving.knative.dev",
		Version:  "v1",
	},
	// Shorthand alias for service
	"svc": {
		Resource: "services",
		Group:    "serving.knative.dev",
		Version:  "v1",
	},
}

func ConfigSinkPrefixes(prefixes []commands.SinkPrefixConfig) {
	for _, p := range prefixes {
		//user configration might override the default configuration
		SinkPrefixes[p.Prefix] = schema.GroupVersionResource{
			Resource: p.Resource,
			Group:    p.Group,
			Version:  p.Version,
		}
	}
}

// ResolveSink returns the Destination referred to by the flags in the acceptor.
// It validates that any object the user is referring to exists.
func (i *SinkFlags) ResolveSink(knclient clientdynamic.KnDynamicClient, namespace string) (*duckv1.Destination, error) {
	client := knclient.RawClient()
	if i.sink == "" {
		return nil, nil
	}

	prefix, name := parseSink(i.sink)
	if prefix == "" {
		// URI target
		uri, err := apis.ParseURL(name)
		if err != nil {
			return nil, err
		}
		return &duckv1.Destination{URI: uri}, nil
	}
	typ, ok := SinkPrefixes[prefix]
	if !ok {
		return nil, fmt.Errorf("unsupported sink type: %s", i.sink)
	}
	obj, err := client.Resource(typ).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	destination := &duckv1.Destination{
		Ref: &duckv1.KReference{
			Kind:       obj.GetKind(),
			APIVersion: obj.GetAPIVersion(),
			Name:       obj.GetName(),
			Namespace:  namespace,
		},
	}
	return destination, nil
}

// parseSink takes the string given by the user into the prefix and the name of
// the object. If the user put a URI instead, the prefix is empty and the name
// is the whole URI.
func parseSink(sink string) (string, string) {
	parts := strings.SplitN(sink, ":", 2)
	if len(parts) == 1 {
		return "svc", parts[0]
	} else if parts[0] == "http" || parts[0] == "https" {
		return "", sink
	} else {
		return parts[0], parts[1]
	}
}

// SinkToString prepares a sink for list output
func SinkToString(sink duckv1.Destination) string {
	if sink.Ref != nil {
		if sink.Ref.Kind == "Service" {
			return fmt.Sprintf("svc:%s", sink.Ref.Name)
		} else {
			return fmt.Sprintf("%s:%s", strings.ToLower(sink.Ref.Kind), sink.Ref.Name)
		}
	}
	if sink.URI != nil {
		return sink.URI.String()
	}
	return ""
}
