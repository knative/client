// Copyright © 2019 The Knative Authors
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

	"github.com/spf13/pflag"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type boolPairTestCase struct {
	name            string
	defaultVal      bool
	flags           []string
	expectedResult  bool
	expectedErrText string
}

func TestBooleanPair(t *testing.T) {
	cases := []*boolPairTestCase{
		{"foo", true, []string{}, true, ""},
		{"foo", true, []string{"--foo"}, true, ""},
		{"foo", true, []string{"--no-foo"}, false, ""},
		{"foo", false, []string{"--foo"}, true, ""},
		{"foo", false, []string{}, false, ""},
		{"foo", false, []string{"--no-foo"}, false, ""},
		{"foo", true, []string{"--foo", "--no-foo"}, false, "only one of"},
	}
	for _, c := range cases {
		f := &pflag.FlagSet{}
		var result bool
		AddBothBoolFlags(f, &result, c.name, "", c.defaultVal, "set "+c.name)
		f.Parse(c.flags)
		err := ReconcileBoolFlags(f)
		if c.expectedErrText != "" {
			assert.ErrorContains(t, err, c.expectedErrText)
		} else {
			assert.Assert(t, cmp.Equal(result, c.expectedResult))
		}
	}
}
