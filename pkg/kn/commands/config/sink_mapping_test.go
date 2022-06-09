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
	"knative.dev/client/pkg/kn/commands/config/sink"
	"knative.dev/client/pkg/kn/config"
)

var initSinkMapping = []config.SinkMapping{
	{Prefix: "svc", Resource: "services", Group: "core", Version: "v1alpha1"},
}

func TestConfigCommandSinkMappingList(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expectedResult := `- prefix: svc
  resource: services
  group: core
  version: v1alpha1
`

	previousValue := viper.Get(config.KeySinkMappings)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(config.KeySinkMappings, previousValue)
	}()

	viper.Set(config.KeySinkMappings, initSinkMapping)
	viper.SetConfigFile(tempfile.Name())
	output, err := executeConfigCommand("sink-mapping", "list")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}
	if output != expectedResult {
		t.Errorf("Expected to get %q but got %q", expectedResult, output)
	}
}

func TestConfigCommandSinkMappingUpdate(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := []config.SinkMapping{
		{Prefix: "svc", Resource: "services", Group: "core", Version: "v1alpha1"},
		{Prefix: "ksvc", Resource: "services", Group: "api", Version: "v3beta1"},
	}

	previousValue := viper.Get(config.KeySinkMappings)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(config.KeySinkMappings, previousValue)
	}()

	viper.Set(config.KeySinkMappings, initSinkMapping)
	viper.SetConfigFile(tempfile.Name())

	if _, err := executeConfigCommand("sink-mapping", "add", "ksvc", "--resource", "services", "--version", "v3beta1", "--group", "api"); err != nil {
		t.Errorf("Test failed: %s", err)
	}

	result := []config.SinkMapping{}
	viper.UnmarshalKey(config.KeySinkMappings, &result)
	if len(result) != 2 {
		t.Errorf("Expected to get %q but got %q", expected, result)
	}
	for ix := range result {
		if result[ix].Prefix != expected[ix].Prefix ||
			result[ix].Resource != expected[ix].Resource ||
			result[ix].Group != expected[ix].Group ||
			result[ix].Version != expected[ix].Version {
			t.Errorf("Expected at index %d to get %q but got %q", ix, expected[ix], result[ix])
		}
	}
}

func TestConfigCommandSinkMappingAdd(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := []config.SinkMapping{
		{Prefix: "svc", Resource: "services", Group: "api", Version: "v2"},
	}

	previousValue := viper.Get(config.KeySinkMappings)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(config.KeySinkMappings, previousValue)
	}()

	viper.Set(config.KeySinkMappings, initSinkMapping)
	viper.SetConfigFile(tempfile.Name())

	if _, err := executeConfigCommand("sink-mapping", "update", "svc", "--version", "v2", "--group", "api"); err != nil {
		t.Errorf("Test failed: %s", err)
	}

	result := []config.SinkMapping{}
	viper.UnmarshalKey(config.KeySinkMappings, &result)
	if len(result) != 1 ||
		result[0].Prefix != expected[0].Prefix ||
		result[0].Resource != expected[0].Resource ||
		result[0].Group != expected[0].Group ||
		result[0].Version != expected[0].Version {
		t.Errorf("Expected to get %q but got %q", expected, result)
	}
}

func TestConfigCommandSinkMappingDelete(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := []config.ChannelTypeMapping{}

	previousValue := viper.Get(config.KeySinkMappings)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(config.KeySinkMappings, previousValue)
	}()

	viper.Set(config.KeySinkMappings, initSinkMapping)
	viper.SetConfigFile(tempfile.Name())

	if _, err := executeConfigCommand("sink-mapping", "delete", "svc"); err != nil {
		t.Errorf("Test failed: %s", err)
	}

	result := []config.SinkMapping{}
	viper.UnmarshalKey(config.KeySinkMappings, &result)
	if len(result) != 0 {
		t.Errorf("Expected to get %q but got %q", expected, result)
	}
}

func TestConfigCommandSinkMappingUpdateNotFoundAlias(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := fmt.Sprintf(sink.SinkPrefixNotFoundErrorMsg, "ksvc")

	previousValue := viper.Get(config.KeySinkMappings)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(config.KeySinkMappings, previousValue)
	}()

	viper.Set(config.KeySinkMappings, initSinkMapping)
	viper.SetConfigFile(tempfile.Name())

	_, err = executeConfigCommand("sink-mapping", "update", "ksvc")
	if err == nil {
		t.Errorf("Expected to get an error but got nothing")
	}
	if err.Error() != expected {
		t.Errorf("Expected to get error %q but got %q", expected, err)
	}
}

func TestConfigCommandSinkMappingAddExistsAlias(t *testing.T) {
	tempfile, err := ioutil.TempFile("", "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}

	expected := fmt.Sprintf(sink.SinkPrefixExistsErrorMsg, "svc")

	previousValue := viper.Get(config.KeySinkMappings)
	defer func() {
		os.Remove(tempfile.Name())
		viper.Set(config.KeySinkMappings, previousValue)
	}()

	viper.Set(config.KeySinkMappings, initSinkMapping)
	viper.SetConfigFile(tempfile.Name())

	_, err = executeConfigCommand("sink-mapping", "add", "svc")
	if err == nil {
		t.Errorf("Expected to get an error but got nothing")
	}
	if err.Error() != expected {
		t.Errorf("Expected to get error %q but got %q", expected, err)
	}
}
