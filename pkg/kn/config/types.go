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

package config

// Package for holding configuration types used in bootstrapping
// and for types in configuration files

type Config interface {

	// ConfigFile returns the location of the configuration file
	ConfigFile() string

	// PluginsDir returns the path to the directory containing plugins
	PluginsDir() string

	// LookupPluginsInPath returns true if plugins should be also looked up
	// in the execution path
	LookupPluginsInPath() bool

	// SinkMappings returns additional mappings for sink prefixes to resources
	SinkMappings() []SinkMapping

	// ChannelTypeMappings returns additional mappings for channel type aliases
	ChannelTypeMappings() []ChannelTypeMapping
}

// SinkMappings is the struct of sink prefix config in kn config
type SinkMapping struct {

	// Prefix is the mapping prefix (like "svc")
	Prefix string

	// Resource is the name for the mapped resource (like "services", mind the plural)
	Resource string

	// Group is the api group for the mapped resource (like "core")
	Group string

	// Version is the api version for the mapped resource (like "v1")
	Version string
}

// ChannelTypeMapping is the struct of ChannelType alias config in kn config
type ChannelTypeMapping struct {

	// Alias is the mapping alias (like "kafka")
	Alias string

	// Kind is the name for the mapped resource kind (like "KafkaChannel")
	Kind string

	// Group is the API group for the mapped resource kind (like "messaging.knative.dev")
	Group string

	// Version is the API version for the mapped resource kind (like "v1alpha1")
	Version string
}

// config Keys for looking up in viper
const (
	keyPluginsDirectory    = "plugins.directory"
	keyPluginsLookupInPath = "plugins.path-lookup"
	keySinkMappings        = "eventing.sink-mappings"
	keyChannelTypeMappings = "eventing.channel-type-mappings"
)

// legacy config keys, deprecated
const (
	legacyKeyPluginsDirectory    = "plugins-dir"
	legacyKeyPluginsLookupInPath = "lookup-plugins"
	legacyKeySinkMappings        = "sink"
)

// Global (hidden) flags
// TODO: Remove me if decided that they are not needed
const (
	flagPluginsDir          = "plugins-dir"
	flagPluginsLookupInPath = "lookup-plugins"
)
