/*
 Copyright 2024 The Knative Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package tui_test

import (
	"strings"
	"testing"
	"time"

	"knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/tui"
)

// TestSpinner describes the functionality of the Spinner widget in TUI.
// This test verifies that the Spinner widget correctly updates its message
// and completes when all updates have been applied.
func TestSpinner(t *testing.T) {
	t.Parallel()
	ctx := context.TestContext(t)
	prt := output.NewTestPrinter()
	ctx = output.WithContext(ctx, prt)
	w := tui.NewWidgets(ctx)
	s := w.NewSpinner("message")

	if s == nil {
		t.Fatal("want spinner, got nil")
	}
	if err := s.With(func(sc tui.SpinnerControl) error {
		time.Sleep(5 * time.Millisecond)
		sc.UpdateMessage("msg-1")
		time.Sleep(5 * time.Millisecond)
		sc.UpdateMessage("msg-2")
		time.Sleep(5 * time.Millisecond)
		return nil
	}); err != nil {
		t.Errorf("want nil, got %v", err)
	}
	got := prt.Outputs().Out.String()
	expectedMsgs := []string{
		"message", "msg-1", "msg-2",
		"▰▱▱", "▰▰▱", "▰▰▰",
		"Done",
		"\u001B[?25l", "\u001B[0D",
		"\u001B[0D\u001B[2K",
	}
	for _, expected := range expectedMsgs {
		if !strings.Contains(got, expected) {
			t.Errorf("Expected to contain %#v within:\n%#v",
				expected, got)
		}
	}
}
