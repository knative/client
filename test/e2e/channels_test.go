// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or im
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build e2e && !serving
// +build e2e,!serving

package e2e

import (
	"testing"

	"gotest.tools/v3/assert"
	"knative.dev/client/lib/test"
	"knative.dev/client/pkg/util"
)

const (
	knChannelTypesConfigContent string = `
eventing:
  channel-type-mappings:
  - alias: imcbeta
    kind: InMemoryChannel
    group: messaging.knative.dev
    version: v1beta1`
)

type channelTypeAliasTestConfig struct {
	knConfigDir  string
	knConfigPath string
}

func (tc *channelTypeAliasTestConfig) setup(t *testing.T) error {
	var err error
	tc.knConfigDir = t.TempDir()
	tc.knConfigPath, err = test.CreateFile("config.yaml", knChannelTypesConfigContent, tc.knConfigDir, test.FileModeReadWrite)
	if err != nil {
		return err
	}
	return nil
}

func TestChannels(t *testing.T) {
	t.Parallel()
	it, err := test.NewKnTest()
	assert.NilError(t, err)
	defer func() {
		assert.NilError(t, it.Teardown())
	}()

	r := test.NewKnRunResultCollector(t, it)
	defer r.DumpIfFailed()

	tc := channelTypeAliasTestConfig{}
	assert.NilError(t, tc.setup(t))

	t.Log("Create a channel with default messaging layer settings")
	test.ChannelCreate(r, "c0")

	t.Log("Create a channel with explicit GVK on the command line")
	test.ChannelCreate(r, "c1", "--type", "messaging.knative.dev:v1beta1:InMemoryChannel")

	t.Log("Create a channel with an alias from kn config: ", tc.knConfigPath)
	test.ChannelCreate(r, "c2", "--type", "imcbeta", "--config", tc.knConfigPath)

	t.Log("List channels")
	listout := test.ChannelList(r)
	assert.Check(t, util.ContainsAll(listout, "NAME", "TYPE", "URL", "AGE", "READY", "REASON"))
	assert.Check(t, util.ContainsAll(listout, "c0", "c1", "c2"))

	t.Log("List channels in YAML format")
	listout = test.ChannelList(r, "-o", "yaml")
	assert.Check(t, util.ContainsAllIgnoreCase(listout, "apiVersion", "items", "c0", "c2", "kind", "ChannelList"))

	t.Log("Describe a channel")
	descout := test.ChannelDescribe(r, "c0")
	assert.Check(t, util.ContainsAllIgnoreCase(descout, "Name", "c0", "Namespace", "Annotations", "Age", "Type", "URL", "Conditions"))

	t.Log("Delete channels")
	test.ChannelDelete(r, "c0")
	test.ChannelDelete(r, "c1")
	test.ChannelDelete(r, "c2")

	t.Log("List channel types")
	listout = test.ChannelListTypes(r)
	assert.Check(t, util.ContainsAll(listout, "TYPE", "NAME", "DESCRIPTION", "InMemoryChannel"))

	t.Log("List channel types no header")
	listout = test.ChannelListTypes(r, "--no-headers")
	assert.Check(t, util.ContainsNone(listout, "TYPE", "NAME", "DESCRIPTION"))
	assert.Check(t, util.ContainsAll(listout, "InMemoryChannel"))
}
