// Copyright © 2018 The Knative Authors
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
	"testing"

	"gotest.tools/v3/assert"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

func TestNewPluginCommand(t *testing.T) {
	pluginCmd := NewPluginCommand(&commands.KnParams{})
	if pluginCmd == nil {
		t.Fatal("pluginCmd = nil, want not nil")
	}

	assert.Equal(t, pluginCmd.Use, "plugin")
	assert.Assert(t, util.ContainsAllIgnoreCase(pluginCmd.Short, "plugin"))
	assert.Assert(t, util.ContainsAllIgnoreCase(pluginCmd.Long, "plugins"))
}
