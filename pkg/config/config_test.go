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

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gotest.tools/v3/assert"
)

// TestDefaultProfile tests that the default profile is correctly loaded
func TestDefaultProfile(t *testing.T) {
	profile := builtInProfiles()
	assert.Equal(t, len(profile[istio].Labels), 0)
	assert.Equal(t, len(profile[istio].Annotations), 3)
}

func TestBootstrapConfig(t *testing.T) {
	configYaml := `
plugins:
  directory: /tmp
profiles:
  knative:
    labels:
      - name: environment
        value: "knative"
    annotations:
      - name: sidecar.knative.io/inject
        value: "true"
      - name: sidecar.knative.io/rewriteAppHTTPProbers
        value: "true"
eventing:
  sink-mappings:
  - prefix: service
    resource: services
    group: core
    version: v1
  channel-type-mappings:
  - alias: kafka
    kind: KafkaChannel
    group: messaging.knative.dev
    version: v1alpha1
`

	configFile, cleanup := setupConfig(t, configYaml)
	defer cleanup()

	err := BootstrapConfig()
	assert.NilError(t, err)

	assert.Equal(t, GlobalConfig.ConfigFile(), configFile)
	assert.Equal(t, GlobalConfig.PluginsDir(), "/tmp")
	assert.Equal(t, GlobalConfig.LookupPluginsInPath(), true)
	assert.Equal(t, len(GlobalConfig.SinkMappings()), 1)
	assert.Equal(t, len(GlobalConfig.Profile("istio").Annotations), 3)
	assert.DeepEqual(t, GlobalConfig.Profile("istio").Annotations, []NamedValue{
		{
			Name:  "sidecar.istio.io/inject",
			Value: "true",
		},
		{
			Name:  "sidecar.istio.io/rewriteAppHTTPProbers",
			Value: "true",
		},
		{
			Name:  "serving.knative.openshift.io/enablePassthrough",
			Value: "true",
		},
	})
	assert.DeepEqual(t, GlobalConfig.Profile("knative").Annotations, []NamedValue{
		{
			Name:  "sidecar.knative.io/inject",
			Value: "true",
		},
		{
			Name:  "sidecar.knative.io/rewriteAppHTTPProbers",
			Value: "true",
		},
	})
	assert.Equal(t, len(GlobalConfig.Profile("knative").Labels), 1)
	assert.DeepEqual(t, GlobalConfig.Profile("knative").Labels, []NamedValue{
		{
			Name:  "environment",
			Value: "knative",
		},
	})
	assert.DeepEqual(t, (GlobalConfig.SinkMappings())[0], SinkMapping{
		Prefix:   "service",
		Resource: "services",
		Group:    "core",
		Version:  "v1",
	})
	assert.Equal(t, len(GlobalConfig.ChannelTypeMappings()), 1)
	assert.DeepEqual(t, (GlobalConfig.ChannelTypeMappings())[0], ChannelTypeMapping{
		Alias:   "kafka",
		Kind:    "KafkaChannel",
		Group:   "messaging.knative.dev",
		Version: "v1alpha1",
	})
}

func TestBootstrapConfigWithoutConfigFile(t *testing.T) {
	_, cleanup := setupConfig(t, "")
	defer cleanup()

	err := BootstrapConfig()
	assert.NilError(t, err)

	assert.Equal(t, GlobalConfig.ConfigFile(), bootstrapDefaults.configFile)
	assert.Equal(t, GlobalConfig.PluginsDir(), bootstrapDefaults.pluginsDir)
	assert.Equal(t, GlobalConfig.LookupPluginsInPath(), bootstrapDefaults.lookupPluginsInPath)
}

func TestBootstrapLegacyConfigFields(t *testing.T) {
	configYaml := `
plugins-dir: /legacy-plugin
lookup-plugins: true
sink:
- prefix: service
  resource: services
  group: core
  version: v1
`

	configFile, cleanup := setupConfig(t, configYaml)
	defer cleanup()

	err := BootstrapConfig()
	assert.NilError(t, err)

	assert.Equal(t, GlobalConfig.ConfigFile(), configFile)
	assert.Equal(t, GlobalConfig.PluginsDir(), "/legacy-plugin")
	assert.Equal(t, GlobalConfig.LookupPluginsInPath(), true)
	assert.Equal(t, len(GlobalConfig.SinkMappings()), 1)
	assert.DeepEqual(t, (GlobalConfig.SinkMappings())[0], SinkMapping{
		Prefix:   "service",
		Resource: "services",
		Group:    "core",
		Version:  "v1",
	})
}

func setupConfig(t *testing.T, configContent string) (string, func()) {
	tmpDir := t.TempDir()

	// Avoid to be fooled by the things in the the real homedir
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Save old args
	backupArgs := os.Args

	// Write out a temporary configContent file
	var cfgFile string
	if configContent != "" {
		cfgFile = filepath.Join(tmpDir, "config.yaml")
		os.Args = []string{"kn", "--config", cfgFile}
		err := os.WriteFile(cfgFile, []byte(configContent), 0644)
		assert.NilError(t, err)
	}

	// Reset various global state
	oldHomeDirDisableCache := homedir.DisableCache
	homedir.DisableCache = true
	viper.Reset()
	globalConfig = config{}
	GlobalConfig = &globalConfig
	bootstrapDefaults = initDefaults()

	return cfgFile, func() {
		// Cleanup everything
		os.Setenv("HOME", oldHome)
		os.Args = backupArgs
		bootstrapDefaults = initDefaults()
		viper.Reset()
		homedir.DisableCache = oldHomeDirDisableCache
		globalConfig = config{}
		GlobalConfig = &globalConfig
	}
}
