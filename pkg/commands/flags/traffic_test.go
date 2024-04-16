// Copyright © 2021 The Knative Authors
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

func TestAdd(t *testing.T) {
	trafficFlags := Traffic{
		RevisionsPercentages: []string{"20", "80"},
		RevisionsTags:        []string{"tag1", "tag2"},
		UntagRevisions:       []string{"tag1"},
	}
	trafficCmd := &cobra.Command{
		Use:   "kn",
		Short: "Traffic test kn command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
	trafficFlags.Add(trafficCmd)
	flagList := trafficCmd.Flags()
	_, err := flagList.GetStringSlice("traffic")
	assert.NilError(t, err)
	_, err = flagList.GetStringSlice("tag")
	assert.NilError(t, err)
	_, err = flagList.GetStringSlice("untag")
	assert.NilError(t, err)
	_, err = flagList.GetStringSlice("undefined")
	assert.ErrorContains(t, err, "not defined")

	assert.Equal(t, false, trafficFlags.PercentagesChanged(trafficCmd))
	assert.Equal(t, false, trafficFlags.TagsChanged(trafficCmd))
	assert.Equal(t, false, trafficFlags.Changed(trafficCmd))
}
