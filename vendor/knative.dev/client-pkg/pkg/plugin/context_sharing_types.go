// Copyright © 2023 The Knative Authors
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

package plugin

// Manifest represents plugin metadata
type Manifest struct {
	// Path to external plugin binary. Always empty for inlined plugins.
	Path string `json:"path,omitempty"`

	// Plugin declares its own manifest to be included in Context Sharing feature
	HasManifest bool `json:"hasManifest"`

	// ProducesContextDataKeys is a list of keys for the ContextData that
	// a plugin can produce. Nil or an empty list declares that this
	// plugin is not ContextDataProducer
	ProducesContextDataKeys []string `json:"producesKeys,omitempty"`

	// ConsumesContextDataKeys is a list of keys from a ContextData that a
	// plugin is interested in to consume. Nil or an empty list declares
	// that this plugin is not a ContextDataConsumer
	ConsumesContextDataKeys []string `json:"consumesKeys,omitempty"`
}

// PluginWithManifest represents extended plugin support for Manifest and Context Sharing feature
type PluginWithManifest interface {
	// Plugin original interface wrapper
	Plugin

	// GetManifest returns a metadata descriptor of plugin
	GetManifest() *Manifest

	// GetContextData returns a map structure that is used to share common context data from plugin
	GetContextData() map[string]string

	// ExecuteWithContextData an extended version of plugin.Execute command to allow injection and use of ContextData.
	// This is a map[string]string structure to share common value service name between kn and plugins.
	ExecuteWithContextData(ctx map[string]string, args []string) error
	// Alternative impl: use of native context.Context as a data structure.
	// We decided to use plain map, because Context can't be extended to external plugins anyway.
	// Another encode to JSON is required to pass it. We might reconsider that in future development iterations.

}
