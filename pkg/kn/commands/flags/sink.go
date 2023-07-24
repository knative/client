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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	clientdynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/config"
)

type SinkFlags struct {
	Sink         string
	SinkMappings map[string]schema.GroupVersionResource
}

// NewSinkFlag is a constructor function to create SinkFlags from provided map
func NewSinkFlag(mapping map[string]schema.GroupVersionResource) *SinkFlags {
	return &SinkFlags{
		SinkMappings: mapping,
	}
}

// AddWithFlagName configures Sink flag with given flag name and a short flag name
// pass empty short flag name if you don't want to set one
func (i *SinkFlags) AddWithFlagName(cmd *cobra.Command, fname, short string) {
	flag := "--" + fname
	if short == "" {
		cmd.Flags().StringVar(&i.Sink, fname, "", "")
	} else {
		cmd.Flags().StringVarP(&i.Sink, fname, short, "", "")
	}
	cmd.Flag(fname).Usage = "Addressable sink for events. " +
		"You can specify a broker, channel, Knative service or URI. " +
		"Examples: '" + flag + " broker:nest' for a broker 'nest', " +
		"'" + flag + " channel:pipe' for a channel 'pipe', " +
		"'" + flag + " ksvc:mysvc:mynamespace' for a Knative service 'mysvc' in another namespace 'mynamespace', " +
		"'" + flag + " https://event.receiver.uri' for an URI with an 'http://' or 'https://' schema, " +
		"'" + flag + " ksvc:receiver' or simply '" + flag + " receiver' for a Knative service 'receiver' in the current namespace. " +
		"'" + flag + " special.eventing.dev/v1alpha1/channels:pipe' for GroupVersionResource of v1alpha1 'pipe'. " +
		"If a prefix is not provided, it is considered as a Knative service in the current namespace. " +
		"If referring to a Knative service in another namespace, 'ksvc:name:namespace' combination must be provided explicitly."
	// Use default mapping if empty
	if i.SinkMappings == nil {
		i.SinkMappings = defaultSinkMappings
	}
	for _, p := range config.GlobalConfig.SinkMappings() {
		//user configuration might override the default configuration
		i.SinkMappings[p.Prefix] = schema.GroupVersionResource{
			Resource: p.Resource,
			Group:    p.Group,
			Version:  p.Version,
		}
	}
}

// Add configures Sink flag with name 'Sink' amd short name 's'
func (i *SinkFlags) Add(cmd *cobra.Command) {
	i.AddWithFlagName(cmd, "sink", "s")
}

// SinkPrefixes maps prefixes used for sinks to their GroupVersionResources.
var defaultSinkMappings = map[string]schema.GroupVersionResource{
	"broker": {
		Resource: "brokers",
		Group:    "eventing.knative.dev",
		Version:  "v1",
	},
	// Shorthand alias for service
	"ksvc": {
		Resource: "services",
		Group:    "serving.knative.dev",
		Version:  "v1",
	},
	"channel": {
		Resource: "channels",
		Group:    "messaging.knative.dev",
		Version:  "v1",
	},
}

// ResolveSink returns the Destination referred to by the flags in the acceptor.
// It validates that any object the user is referring to exists.
func (i *SinkFlags) ResolveSink(ctx context.Context, knclient clientdynamic.KnDynamicClient, namespace string) (*duckv1.Destination, error) {
	client := knclient.RawClient()
	if i.Sink == "" {
		return nil, nil
	}
	// Use default mapping if empty
	if i.SinkMappings == nil {
		i.SinkMappings = defaultSinkMappings
	}
	prefix, name, ns := parseSink(i.Sink)
	if prefix == "" {
		// URI target
		uri, err := apis.ParseURL(name)
		if err != nil {
			return nil, err
		}
		return &duckv1.Destination{URI: uri}, nil
	}
	gvr, ok := i.SinkMappings[prefix]
	if !ok {
		if prefix == "svc" || prefix == "service" {
			return nil, fmt.Errorf("unsupported Sink prefix: '%s', please use prefix 'ksvc' for knative service", prefix)
		}
		idx := strings.LastIndex(prefix, "/")
		var groupVersion string
		var kind string
		if idx != -1 && idx < len(prefix)-1 {
			groupVersion, kind = prefix[:idx], prefix[idx+1:]
		} else {
			kind = prefix
		}
		parsedVersion, err := schema.ParseGroupVersion(groupVersion)
		if err != nil {
			return nil, err
		}

		// For the RAWclient the resource name must be in lower case plural form.
		// This is the best effort to sanitize the inputs, but the safest way is to provide
		// the appropriate form in user's input.
		if !strings.HasSuffix(kind, "s") {
			kind = kind + "s"
		}
		kind = strings.ToLower(kind)
		gvr = parsedVersion.WithResource(kind)
	}
	if ns != "" {
		namespace = ns
	}
	obj, err := client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
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

// parseSink takes the string given by the user into the prefix, name and namespace of
// the object. If the user put a URI instead, the prefix is empty and the name
// is the whole URI.
func parseSink(sink string) (string, string, string) {
	parts := strings.SplitN(sink, ":", 3)
	switch {
	case len(parts) == 1:
		return "ksvc", parts[0], ""
	case parts[0] == "http" || parts[0] == "https":
		return "", sink, ""
	case len(parts) == 3:
		return parts[0], parts[1], parts[2]
	default:
		return parts[0], parts[1], ""
	}
}

// SinkToString prepares a Sink for list output
func SinkToString(sink duckv1.Destination) string {
	if sink.Ref != nil {
		if sink.Ref.Kind == "Service" && strings.HasPrefix(sink.Ref.APIVersion, defaultSinkMappings["ksvc"].Group) {
			return fmt.Sprintf("ksvc:%s", sink.Ref.Name)
		} else {
			return fmt.Sprintf("%s:%s", strings.ToLower(sink.Ref.Kind), sink.Ref.Name)
		}
	}
	if sink.URI != nil {
		return sink.URI.String()
	}
	return ""
}
