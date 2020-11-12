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
	"testing"

	"gotest.tools/assert"
)

// Test to keep code coverage quality gate happy.
//
func TestTestConfig(t *testing.T) {
	cfg := TestConfig{
		TestPluginsDir:          "pluginsDir",
		TestConfigFile:          "configFile",
		TestLookupPluginsInPath: true,
		TestSinkMappings:        nil,
		TestChannelTypeMappings: nil,
	}

	assert.Equal(t, cfg.PluginsDir(), "pluginsDir")
	assert.Equal(t, cfg.ConfigFile(), "configFile")
	assert.Assert(t, cfg.LookupPluginsInPath())
	assert.Assert(t, cfg.SinkMappings() == nil)
	assert.Assert(t, cfg.ChannelTypeMappings() == nil)
}
