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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/client/pkg/config"
	clientdynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/flags/sink"
	"knative.dev/client/pkg/util/errors"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// SinkFlags holds information about given sink together with optional mappings
// to allow ease of referencing the common types.
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
// pass empty short flag name if you don't want to set one.
func (i *SinkFlags) AddWithFlagName(cmd *cobra.Command, fname, short string) {
	i.AddToFlagSet(cmd.Flags(), fname, short)
}

// AddToFlagSet configures Sink flag with given flag name and a short flag name
// pass empty short flag name if you don't want to set one
func (i *SinkFlags) AddToFlagSet(fs *pflag.FlagSet, fname, short string) {
	if short == "" {
		fs.StringVar(&i.Sink, fname, "", "")
	} else {
		fs.StringVarP(&i.Sink, fname, short, "", "")
	}
	fs.Lookup(fname).Usage = sink.Usage(fname)
}

// Add configures Sink flag with name 'Sink' amd short name 's'
func (i *SinkFlags) Add(cmd *cobra.Command) {
	i.AddWithFlagName(cmd, sink.DefaultFlagName, sink.DefaultFlagShorthand)
}

// WithDefaultMappings will return a copy of SinkFlags with provided mappings
// and the default ones.
func (i *SinkFlags) WithDefaultMappings() *SinkFlags {
	sf := &SinkFlags{
		Sink: i.Sink,
		SinkMappings: make(map[string]schema.GroupVersionResource,
			len(i.SinkMappings)+len(sink.DefaultMappings)),
	}
	for k, v := range sink.DefaultMappings {
		sf.SinkMappings[k] = v
	}
	for k, v := range i.SinkMappings {
		sf.SinkMappings[k] = v
	}
	for _, p := range config.GlobalConfig.SinkMappings() {
		// user configuration might override the default configuration
		sf.SinkMappings[p.Prefix] = schema.GroupVersionResource{
			Resource: p.Resource,
			Group:    p.Group,
			Version:  p.Version,
		}
	}
	return sf
}

// Parse returns the sink reference, which may refer to URL or to Kubernetes
// resource. The namespace given should be the current namespace within the
// context.
func (i *SinkFlags) Parse(namespace string) (*sink.Reference, error) {
	// Use default mapping if empty
	sf := i.WithDefaultMappings()
	return sink.Parse(sf.Sink, namespace, sf.SinkMappings)
}

// ResolveSink returns the Destination referred to by the flags in the acceptor.
// It validates that any object the user is referring to exists.
func (i *SinkFlags) ResolveSink(ctx context.Context, knclient clientdynamic.KnDynamicClient, namespace string) (*duckv1.Destination, error) {
	s, err := i.Parse(namespace)
	if err != nil {
		return nil, err
	}
	var dest *duckv1.Destination
	dest, err = s.Resolve(ctx, knclient)
	if err != nil {
		// Returning original error that caused sink.ErrSinkIsInvalid as it is
		// directly presented to the end-user.
		return nil, errors.CauseOf(err, sink.ErrSinkIsInvalid)
	}
	return dest, nil
}

// SinkToString prepares a Sink for list output
// Deprecated: use (*sink.Reference).AsText instead.
func SinkToString(dest duckv1.Destination) string {
	ref := sink.GuessFromDestination(dest)
	return ref.String()
}
