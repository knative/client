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
	"testing"
	"time"

	"knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/tui"
)

func TestSpinner(t *testing.T) {
	t.Parallel()
	ctx := context.TestContext(t)
	prt := output.NewTestPrinter()
	ctx = output.WithContext(ctx, prt)
	w := tui.NewWidgets(ctx)
	s := w.NewSpinner("message")

	if s == nil {
		t.Errorf("want spinner, got nil")
	}
	if err := s.With(func(sc tui.SpinnerControl) error {
		time.Sleep(3 * time.Millisecond)
		sc.UpdateMessage("msg-1")
		time.Sleep(3 * time.Millisecond)
		sc.UpdateMessage("msg-2")
		time.Sleep(3 * time.Millisecond)
		return nil
	}); err != nil {
		t.Errorf("want nil, got %v", err)
	}
	got := prt.Outputs().Out.String()
	want := "\x1b[?25lmessage ▰▱▱\x1b[0D" +
		"\x1b[0D\x1b[2Kmsg-1 ▰▰▱\x1b[0D" +
		"\x1b[0D\x1b[2Kmsg-2 ▰▰▰\x1b[0D" +
		"\x1b[2K\x1b[?25h\x1b[?1002l\x1b[?1003l\x1b[?1006lmsg-2 Done\n"
	if got != want {
		t.Errorf("text missmatch\nwant %q,\n got %q", want, got)
	}
}
