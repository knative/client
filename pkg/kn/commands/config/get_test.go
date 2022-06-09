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
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"knative.dev/client/pkg/kn/config"
)

func TestConfigCommandGetSuccess(t *testing.T) {
	testCases := []struct {
		name          string
		key           string
		value         interface{}
		expectedValue string
	}{
		{
			name:          "get plugins directory",
			key:           "plugins.directory",
			value:         "/kn/plugins",
			expectedValue: "/kn/plugins",
		},
		{
			name:  "get plugins",
			key:   "plugins",
			value: map[string]map[string]string{"plugins": {"directory": "/kn/plugins"}},
			expectedValue: `plugins:
  directory: /kn/plugins`,
		},
		{
			name: "get channel-type-mappings",
			key:  "eventing.channel-type-mappings",
			value: map[string]map[string][]config.ChannelTypeMapping{
				"eventing": {
					"channel-type-mappings": []config.ChannelTypeMapping{
						{Alias: "kinesis", Kind: "KafkaChannel", Group: "dev.1", Version: "v1"},
						{Alias: "kafka", Kind: "KafkaChannel", Group: "knative.messaging", Version: "v1alpha1"},
						{Alias: "0MQ", Kind: "ZeroMQChannel", Group: "dev.1", Version: "v1beta1"},
					}}},
			expectedValue: `eventing:
  channel-type-mappings:
  - alias: kinesis
    kind: KafkaChannel
    group: dev.1
    version: v1
  - alias: kafka
    kind: KafkaChannel
    group: knative.messaging
    version: v1alpha1
  - alias: 0MQ
    kind: ZeroMQChannel
    group: dev.1
    version: v1beta1`,
		},
	}

	for _, tc := range testCases {
		previousValue := viper.Get(tc.key)
		viper.Set(tc.key, tc.value)
		output, err := executeConfigCommand("get", tc.key)
		if err != nil {
			t.Errorf("Test %q failed: expected no errors but got %s", tc.name, err)
		}
		viper.Set(tc.key, previousValue)
		result := strings.TrimSuffix(output, "\n")
		if result != tc.expectedValue {
			t.Errorf("Test %q failed: expected %q but got %q", tc.name, tc.expectedValue, result)
		}
	}
}

func TestConfigCommandGetNotSetKey(t *testing.T) {
	notExistKey := "system-config"
	_, err := executeConfigCommand("get", notExistKey)
	if err == nil {
		t.Errorf("Expected test to fail but it did not")
	}
	expected := fmt.Sprintf(ConfigKeyNotFoundErrorMsg, notExistKey)
	if err != nil && err.Error() != expected {
		t.Errorf("Expected to get %s but %s", expected, err)
	}
}
