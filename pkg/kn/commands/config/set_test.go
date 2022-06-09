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
)

func TestConfigCommandSetSuccess(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "plugins.*")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}
	tempfile, err := ioutil.TempFile(tmpdir, "test.*.yaml")
	if err != nil {
		t.Errorf("Test failed: %s", err)
	}
	defer func() {
		os.RemoveAll(tmpdir)
		os.Remove(tempfile.Name())
	}()
	cfgKey := "plugins.directory"
	initValue := ""
	newValue := tmpdir

	previousValue := viper.Get(cfgKey)
	viper.Set(cfgKey, initValue)
	viper.SetConfigFile(tempfile.Name())
	if _, err := executeConfigCommand("set", fmt.Sprintf("%s=%s", cfgKey, newValue)); err != nil {
		t.Errorf("Test failed: %s", err)
	}
	result := viper.GetString(cfgKey)
	viper.Set(cfgKey, previousValue)
	if result != newValue {
		t.Errorf("Expected to get %s but got %s", newValue, result)
	}
}

func TestConfigCommandSetMissingValueInArgument(t *testing.T) {
	arg := "plugins.directory"
	_, err := executeConfigCommand("set", arg)
	if err == nil {
		t.Errorf("Expected test to fail but it did not")
	}
	expected := fmt.Sprintf(ArgumentMissingValueErrorMsg, arg)

	if err != nil && err.Error() != expected {
		t.Errorf("Expected to get %s but %s", expected, err)
	}
}

func TestConfigCommandSetTwoManyEqualSignsInArgument(t *testing.T) {
	arg := "plugins.directory=bla=bla"
	_, err := executeConfigCommand("set", arg)
	if err == nil {
		t.Errorf("Expected test to fail but it did not")
	}
	expected := fmt.Sprintf(ArgumentMissingValueErrorMsg, arg)

	if err != nil && err.Error() != expected {
		t.Errorf("Expected to get %s but %s", expected, err)
	}
}

func TestConfigCommandSetNonScalarConfig(t *testing.T) {
	cfgKey := "plugins"
	arg := fmt.Sprintf("%s={directory: ~/kn/plugins}", cfgKey)
	_, err := executeConfigCommand("set", arg)
	if err == nil {
		t.Errorf("Expected test to fail but it did not")
	}
	expected := fmt.Sprintf(NonScalarConfigPathErrorMsg, cfgKey)

	if err != nil && err.Error() != expected {
		t.Errorf("Expected to get %s but %s", expected, err)
	}
}
