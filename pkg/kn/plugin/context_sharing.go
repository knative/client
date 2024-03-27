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

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"knative.dev/client/pkg/kn/config"
)

//--TYPES--

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

	// GetManifest
	GetManifest() *Manifest

	// GetContextData
	GetContextData() map[string]string

	// ExecuteWithContext
	ExecuteWithContext(ctx map[string]string, args []string) error
}

//--TYPES--

var CtxManager *ContextDataManager

type ContextDataManager struct {
	//ContextData map[string]ContextData `json:"-"`
	PluginManager *Manager            `json:"-"`
	Producers     map[string][]string `json:"producers"`
	Consumers     map[string][]string `json:"consumers"`
	Manifests     map[string]Manifest `json:"manifests"`
}

func NewContextManager(pluginManager *Manager) (*ContextDataManager, error) {
	if CtxManager == nil {
		CtxManager = &ContextDataManager{
			PluginManager: pluginManager,
			Producers:     map[string][]string{},
			Consumers:     map[string][]string{},
			Manifests:     map[string]Manifest{},
		}
	}
	return CtxManager, nil
}

// GetConsumesKeys returns array of keys consumed by plugin
func (c *ContextDataManager) GetConsumesKeys(pluginName string) []string {
	return c.Manifests[pluginName].ConsumesContextDataKeys
}

// GetProducesKeys returns array of keys produced by plugin
func (c *ContextDataManager) GetProducesKeys(pluginName string) []string {
	return c.Manifests[pluginName].ProducesContextDataKeys
}

func (c *ContextDataManager) FetchContextData() (map[string]string, error) {
	// Load cached data first
	if err := c.loadCache(); err != nil {
		return nil, err
	}

	// Fetch manifests
	if err := c.FetchManifests(); err != nil {
		return nil, err
	}

	// Get context data, limited to func only
	for _, p := range c.PluginManager.GetInternalPlugins() {
		if p.Name() == "kn func" {
			if pwm, ok := p.(PluginWithManifest); ok {
				return pwm.GetContextData(), nil
			}
		}
	}
	return map[string]string{}, nil
}

// FetchManifests it tries to retrieve manifest from both inlined and external plugins
func (c *ContextDataManager) FetchManifests() error {
	for _, plugin := range c.PluginManager.GetInternalPlugins() {
		manifest := &Manifest{}
		// For the integrity build the same name format as external plugins
		pluginName := "kn-" + strings.Join(plugin.CommandParts(), "-")
		if pwm, ok := plugin.(PluginWithManifest); ok {
			manifest = pwm.GetManifest()
		}

		// Add manifest to map
		c.Manifests[pluginName] = *manifest

		if manifest.HasManifest {
			c.populateDataKeys(manifest, pluginName)
		}
	}
	plugins, err := c.PluginManager.ListPlugins()
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		// Add new plugins only
		if _, exists := c.Manifests[plugin.Name()]; !exists {

			if plugin.Path() != "" {
				manifest := &Manifest{
					Path: plugin.Path(),
				}
				// Fetch from external plugin
				if m := fetchExternalManifest(plugin); m != nil {
					manifest = m
				}

				// Add manifest to map
				c.Manifests[plugin.Name()] = *manifest

				if manifest.HasManifest {
					c.populateDataKeys(manifest, plugin.Name())
				}
			}
		}
	}
	return nil
}

func (c *ContextDataManager) populateDataKeys(manifest *Manifest, pluginName string) {
	// Build producers mapping
	for _, key := range manifest.ProducesContextDataKeys {
		c.Producers[key] = append(c.Producers[key], pluginName)
	}
	// Build consumers mapping
	for _, key := range manifest.ConsumesContextDataKeys {
		c.Consumers[key] = append(c.Consumers[key], pluginName)
	}
}

// TODO: We should cautiously execute external binaries
// fetchExternalManifest returns Manifest from external plugin by exec `$plugin manifest get`
func fetchExternalManifest(p Plugin) *Manifest {
	cmd := exec.Command(p.Path(), "manifest") //nolint:gosec
	stdOut := new(bytes.Buffer)
	cmd.Stdout = stdOut
	manifest := &Manifest{
		Path: p.Path(),
	}
	if err := cmd.Run(); err != nil {
		return nil
	}
	d := json.NewDecoder(stdOut)
	if err := d.Decode(manifest); err != nil {
		return nil
	}
	manifest.HasManifest = true
	return manifest
}

func (c *ContextDataManager) loadCache() error {
	cacheFile := filepath.Join(filepath.Dir(config.GlobalConfig.ConfigFile()), "context.json")
	if _, err := os.Stat(cacheFile); err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file yet
		} else {
			return err
		}
	}

	file, err := os.Open(cacheFile)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	ctxManager := &ContextDataManager{}
	if err := decoder.Decode(ctxManager); err != nil {
		return err
	}
	c.Manifests = ctxManager.Manifests
	c.Producers = ctxManager.Producers
	c.Consumers = ctxManager.Consumers
	return nil
}

// WriteCache store data back to cache file
func (c *ContextDataManager) WriteCache() error {
	out := new(bytes.Buffer)
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(c); err != nil {
		return nil
	}
	return os.WriteFile(filepath.Join(filepath.Dir(config.GlobalConfig.ConfigFile()), "context.json"), out.Bytes(), fs.FileMode(0664))
}
