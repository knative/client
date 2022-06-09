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
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"knative.dev/client/pkg/kn/commands/config/channel"
	"knative.dev/client/pkg/kn/config"
)

const cfgKey = config.KeyChannelTypeMappings

var initValue = []config.ChannelTypeMapping{
	{Alias: "Kafka", Kind: "KafkaChannel", Group: "messaging.knative.dev", Version: "v1alpha1"},
}

func TestConfigCommandChannelTypeMappingList(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expectedResult := `- alias: Kafka
  kind: KafkaChannel
  group: messaging.knative.dev
  version: v1alpha1
`

	previousValue := viper.Get(cfgKey)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(cfgKey, previousValue)
	}()

	viper.Set(cfgKey, initValue)
	viper.SetConfigFile(tempfile.Name())
	output, err := executeConfigCommand("channel-type-mapping", "list")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}
	if output != expectedResult {
		t.Errorf("Expected to get %q but got %q", expectedResult, output)
	}
}

func TestConfigCommandChannelTypeMappingUpdate(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := []config.ChannelTypeMapping{
		{Alias: "Kafka", Kind: "KafkaChannel", Group: "messaging.knative.dev", Version: "v1alpha1"},
		{Alias: "RabbitMQ", Kind: "RabbitMQChannel", Group: "messaging.knative.dev", Version: "v2"},
	}

	previousValue := viper.Get(cfgKey)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(cfgKey, previousValue)
	}()

	viper.Set(cfgKey, initValue)
	viper.SetConfigFile(tempfile.Name())

	if _, err := executeConfigCommand("channel-type-mapping", "add", "RabbitMQ", "--kind", "RabbitMQChannel", "--version", "v2", "--group", "messaging.knative.dev"); err != nil {
		t.Errorf("Test failed: %s", err)
	}

	result := []config.ChannelTypeMapping{}
	viper.UnmarshalKey(cfgKey, &result)
	if len(result) != 2 {
		t.Errorf("Expected to get %q but got %q", expected, result)
	}
	for ix := range result {
		if result[ix].Alias != expected[ix].Alias ||
			result[ix].Kind != expected[ix].Kind ||
			result[ix].Group != expected[ix].Group ||
			result[ix].Version != expected[ix].Version {
			t.Errorf("Expected at index %d to get %q but got %q", ix, expected[ix], result[ix])
		}
	}
}

func TestConfigCommandChannelTypeMappingAdd(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := []config.ChannelTypeMapping{
		{Alias: "Kafka", Kind: "KafkaChannel", Group: "msg.knative.dev", Version: "v1"},
	}

	previousValue := viper.Get(cfgKey)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(cfgKey, previousValue)
	}()

	viper.Set(cfgKey, initValue)
	viper.SetConfigFile(tempfile.Name())

	if _, err := executeConfigCommand("channel-type-mapping", "update", "Kafka", "--version", "v1", "--group", "msg.knative.dev"); err != nil {
		t.Errorf("Test failed: %s", err)
	}

	result := []config.ChannelTypeMapping{}
	viper.UnmarshalKey(cfgKey, &result)
	if len(result) != 1 ||
		result[0].Alias != expected[0].Alias ||
		result[0].Kind != expected[0].Kind ||
		result[0].Group != expected[0].Group ||
		result[0].Version != expected[0].Version {
		t.Errorf("Expected to get %q but got %q", expected, result)
	}
}

func TestConfigCommandChannelTypeMappingDelete(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := []config.ChannelTypeMapping{}

	previousValue := viper.Get(cfgKey)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(cfgKey, previousValue)
	}()

	viper.Set(cfgKey, initValue)
	viper.SetConfigFile(tempfile.Name())

	if _, err := executeConfigCommand("channel-type-mapping", "delete", "Kafka"); err != nil {
		t.Errorf("Test failed: %s", err)
	}

	result := []config.ChannelTypeMapping{}
	viper.UnmarshalKey(cfgKey, &result)
	if len(result) != 0 {
		t.Errorf("Expected to get %q but got %q", expected, result)
	}
}

func TestConfigCommandChannelTypeMappingUpdateNotFoundAlias(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := fmt.Sprintf(channel.ChannelTypeAliasNotFoundErrorMsg, "Kafka2")

	previousValue := viper.Get(cfgKey)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(cfgKey, previousValue)
	}()

	viper.Set(cfgKey, initValue)
	viper.SetConfigFile(tempfile.Name())

	_, err = executeConfigCommand("channel-type-mapping", "update", "Kafka2")
	if err == nil {
		t.Errorf("Expected to get an error but got nothing")
	}
	if err.Error() != expected {
		t.Errorf("Expected to get error %q but got %q", expected, err)
	}
}

func TestConfigCommandChannelTypeMappingAddExistsAlias(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := fmt.Sprintf(channel.ChannelTypeAliasExistsErrorMsg, "Kafka")

	previousValue := viper.Get(cfgKey)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(cfgKey, previousValue)
	}()

	viper.Set(cfgKey, initValue)
	viper.SetConfigFile(tempfile.Name())

	_, err = executeConfigCommand("channel-type-mapping", "add", "Kafka")
	if err == nil {
		t.Errorf("Expected to get an error but got nothing")
	}
	if err.Error() != expected {
		t.Errorf("Expected to get error %q but got %q", expected, err)
	}
}
