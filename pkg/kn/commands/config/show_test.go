// Copyright © 2022 The Knative Authors
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

package configCmd

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"knative.dev/client/pkg/kn/config"
)

func TestConfigCommandShowSuccess(t *testing.T) {
	config := map[string]interface{}{
		"eventing.sink-mappings": []config.SinkMapping{
			{Prefix: "svc", Resource: "services", Group: "core", Version: "v2"},
		},
		"eventing.channel-type-mappings": []config.ChannelTypeMapping{
			{Alias: "kafka", Kind: "KafkaChannel", Group: "messaging.knative.dev", Version: "v1alpha1"},
		},
		"plugins.directory": "~/.config/kn/v2/plugins",
	}
	prevConfigs := map[string]interface{}{
		"eventing.sink-mappings":         nil,
		"eventing.channel-type-mappings": nil,
		"plugins.directory":              "~/.config/kn/v2/plugins",
	}
	for cfgKey, newValue := range config {
		prevConfigs[cfgKey] = viper.Get(cfgKey)
		viper.Set(cfgKey, newValue)
	}
	defer func() {
		for key, value := range prevConfigs {
			viper.Set(key, value)
		}
	}()
	expectedValue := `eventing:
  channel-type-mappings:
  - alias: kafka
    kind: KafkaChannel
    group: messaging.knative.dev
    version: v1alpha1
  sink-mappings:
  - prefix: svc
    resource: services
    group: core
    version: v2
plugins:
  directory: ~/.config/kn/v2/plugins`
	output, err := executeConfigCommand("show")
	if err != nil {
		t.Errorf("Test failed: expected no errors but got %s", err)
	}
	result := strings.TrimSuffix(output, "\n")
	if result != expectedValue {
		t.Errorf("Test failed: expected %q but got %q", expectedValue, result)
	}
}
