package commands

import (
	"github.com/spf13/cobra"
	"strings"
	"testing"
)

type waitTestCase struct {
	args                 []string
	waitByDefault        bool
	timeoutExpected      int
	isWaitExpected       bool
	isParseErrorExpected bool
}

func TestAddWaitForReadyFlags(t *testing.T) {

	for i, tc := range []waitTestCase{
		{[]string{"--no-wait"}, true, 60, false, false},
		{[]string{"--wait"}, true, 60, false, true},
		{[]string{"--no-wait"}, false, 60, false, true},
		{[]string{"--wait"}, false, 60, true, false},
		{[]string{}, true, 60, true, false},
		{[]string{}, false, 60, false, false},
		{[]string{"--wait-timeout=120"}, true, 120, true, false},
		{[]string{"--wait", "--wait-timeout=240"}, false, 240, true, false},
		// Can't be easily prevented, the timeout is just ignored in this case:
		{[]string{"--no-wait", "--wait-timeout=120"}, true, 120, false, false},
	} {

		flags := &WaitFlags{}
		cmd := cobra.Command{}
		flags.AddWaitFlags(&cmd, tc.waitByDefault, 60, "service")

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
		if flags.IsWait(tc.waitByDefault) != tc.isWaitExpected {
			t.Errorf("%d: wrong wait mode detected: %t (expected) != %t (actual)", i, tc.isWaitExpected, flags.IsWait(tc.waitByDefault))
		}
		if flags.Timeout != tc.timeoutExpected {
			t.Errorf("%d: Invalid timeout set. %d (expected) != %d (actual)", i, tc.timeoutExpected, flags.Timeout)
		}
	}
}

func TestAddWaitUsageMessage(t *testing.T) {

	flags := &WaitFlags{}
	cmd := cobra.Command{}
	flags.AddWaitFlags(&cmd, true, 60, "blub")
	if !strings.Contains(cmd.UsageString(), "blub") {
		t.Error("no type returned in usage")
	}
	if !strings.Contains(cmd.UsageString(), "Don't wait") {
		t.Error("wrong usage message")
	}
	if !strings.Contains(cmd.UsageString(), "60") {
		t.Error("default timeout not contained")
	}
	if strings.Contains(cmd.UsageString(), "--wait ") {
		t.Error("--wait shouldn't be in usage message")
		println(cmd.UsageString())
	}
	if !strings.Contains(cmd.UsageString(), "--no-wait ") {
		t.Error("--no-wait should be in usage message")
		println(cmd.UsageString())
	}

	cmd = cobra.Command{}
	flags.AddWaitFlags(&cmd, false, 60, "foobar")
	if strings.Contains(cmd.UsageString(), "--no-wait ") {
		t.Error("--no-wait shouldn't be in usage message")
	}
	if !strings.Contains(cmd.UsageString(), "--wait ") {
		t.Error("--wait should be in usage message")
		println(cmd.UsageString())
	}
}
