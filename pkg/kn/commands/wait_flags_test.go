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

package commands

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type waitTestCase struct {
	args                 []string
	timeoutExpected      int
	isNoWaitExpected     bool
	isParseErrorExpected bool
}

//TODO: deprecated test should be removed with --async flag
func TestAddWaitForReadyDeprecatedFlags(t *testing.T) {

	for i, tc := range []waitTestCase{
		{[]string{"--async"}, 60, true, false},
		{[]string{}, 60, false, false},
		{[]string{"--wait-timeout=120"}, 120, false, false},
		// Can't be easily prevented, the timeout is just ignored in this case:
		{[]string{"--async", "--wait-timeout=120"}, 120, true, false},
		{[]string{"--wait-timeout=bla"}, 0, true, true},
	} {

		flags := &WaitFlags{}
		cmd := cobra.Command{}
		flags.AddConditionWaitFlags(&cmd, 60, "Create", "service", "ready")

		err := cmd.ParseFlags(tc.args)
		if err != nil && !tc.isParseErrorExpected {
			t.Errorf("%d: parse flags: %v", i, err)
		}
		if err == nil && tc.isParseErrorExpected {
			t.Errorf("%d: parse error expected, but got none: %v", i, err)
		}
		if tc.isParseErrorExpected {
			continue
		}
		if flags.Async != tc.isNoWaitExpected {
			t.Errorf("%d: wrong async mode detected: %t (expected) != %t (actual)", i, tc.isNoWaitExpected, flags.Async)
		}
		if flags.TimeoutInSeconds != tc.timeoutExpected {
			t.Errorf("%d: Invalid timeout set. %d (expected) != %d (actual)", i, tc.timeoutExpected, flags.TimeoutInSeconds)
		}
	}
}

func TestAddWaitForReadyFlags(t *testing.T) {

	for i, tc := range []waitTestCase{
		{[]string{"--no-wait"}, 60, true, false},
		{[]string{}, 60, false, false},
		{[]string{"--wait-timeout=120"}, 120, false, false},
		// Can't be easily prevented, the timeout is just ignored in this case:
		{[]string{"--no-wait", "--wait-timeout=120"}, 120, true, false},
		{[]string{"--wait-timeout=bla"}, 0, true, true},
	} {

		flags := &WaitFlags{}
		cmd := cobra.Command{}
		flags.AddConditionWaitFlags(&cmd, 60, "Create", "service", "ready")

		err := cmd.ParseFlags(tc.args)
		if err != nil && !tc.isParseErrorExpected {
			t.Errorf("%d: parse flags: %v", i, err)
		}
		if err == nil && tc.isParseErrorExpected {
			t.Errorf("%d: parse error expected, but got none: %v", i, err)
		}
		if tc.isParseErrorExpected {
			continue
		}
		if flags.NoWait != tc.isNoWaitExpected {
			t.Errorf("%d: wrong wait mode detected: %t (expected) != %t (actual)", i, tc.isNoWaitExpected, flags.NoWait)
		}
		if flags.TimeoutInSeconds != tc.timeoutExpected {
			t.Errorf("%d: Invalid timeout set. %d (expected) != %d (actual)", i, tc.timeoutExpected, flags.TimeoutInSeconds)
		}
	}
}

func TestAddWaitUsageMessage(t *testing.T) {

	flags := &WaitFlags{}
	cmd := cobra.Command{}
	flags.AddConditionWaitFlags(&cmd, 60, "bla", "blub", "deleted")
	if !strings.Contains(cmd.UsageString(), "blub") {
		t.Error("no type returned in usage")
	}
	if !strings.Contains(cmd.UsageString(), "don't wait") {
		t.Error("wrong usage message")
	}
	if !strings.Contains(cmd.UsageString(), "60") {
		t.Error("default timeout not contained")
	}
	if !strings.Contains(cmd.UsageString(), "deleted") {
		t.Error("wrong until message")
	}
}
