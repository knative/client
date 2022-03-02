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

package flags

import (
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/v3/assert"
)

func TestEventtypeFlags_Add(t *testing.T) {
	eventtypeCmd := &cobra.Command{
		Use:   "kn",
		Short: "Eventtype test kn command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	eventtypeFlags := &EventtypeFlags{}
	eventtypeFlags.Add(eventtypeCmd)

	eventtypeCmd.SetArgs([]string{"--type", "cetype", "--broker", "example-broker", "--source", "example.source"})
	eventtypeCmd.Execute()

	flagList := eventtypeCmd.Flags()

	val, err := flagList.GetString("type")
	assert.NilError(t, err)
	assert.Equal(t, val, "cetype")

	val, err = flagList.GetString("broker")
	assert.NilError(t, err)
	assert.Equal(t, val, "example-broker")

	val, err = flagList.GetString("source")
	assert.NilError(t, err)
	assert.Equal(t, val, "example.source")
}
