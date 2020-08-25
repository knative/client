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

// Implementation of Config useful for testing purposes
// Set an instance of this for config.GlobalConfig to mock
// your own configuration setup
type TestConfig struct {
	TestPluginsDir          string
	TestConfigFile          string
	TestLookupPluginsInPath bool
	TestSinkMappings        []SinkMapping
	TestChannelTypeMappings []ChannelTypeMapping
}

// Ensure that TestConfig implements the configuration interface
var _ Config = &TestConfig{}

func (t TestConfig) PluginsDir() string                        { return t.TestPluginsDir }
func (t TestConfig) ConfigFile() string                        { return t.TestConfigFile }
func (t TestConfig) LookupPluginsInPath() bool                 { return t.TestLookupPluginsInPath }
func (t TestConfig) SinkMappings() []SinkMapping               { return t.TestSinkMappings }
func (t TestConfig) ChannelTypeMappings() []ChannelTypeMapping { return t.TestChannelTypeMappings }
