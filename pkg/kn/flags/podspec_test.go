// Copyright 2020 The Knative Authors
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
	"gotest.tools/assert"
)

func TestPodSpecFlags(t *testing.T) {
	args := []string{"--image", "repo/user/imageID:tag", "--env", "b=c"}
	wantedPod := &PodSpecFlags{
		Image:   "repo/user/imageID:tag",
		Env:     []string{"b=c"},
		EnvFrom: []string{},
		Mount:   []string{},
		Volume:  []string{},
		Arg:     []string{},
	}
	flags := &PodSpecFlags{}
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			assert.DeepEqual(t, wantedPod, flags)
		},
	}
	testCmd.SetArgs(args)
	flags.AddFlags(testCmd.Flags())
	testCmd.Execute()
}

func TestUniqueStringArg(t *testing.T) {
	var a uniqueStringArg
	a.Set("test")
	assert.Equal(t, "test", a.String())
	assert.Equal(t, "string", a.Type())
}
