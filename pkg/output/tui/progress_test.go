//go:build !race

// TODO: there is a race condition in the progress code that needs to be fixed
//       somehow

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
	"crypto/rand"
	"strings"
	"testing"
	"time"

	"knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/tui"
)

func TestProgress(t *testing.T) {
	t.Parallel()
	ctx := context.TestContext(t)
	prt := output.NewTestPrinter()
	ctx = output.WithContext(ctx, prt)
	w := tui.NewWidgets(ctx)
	p := w.NewProgress(42_000, tui.Message{Text: "message"})
	// This is a hack to make the test run faster
	p.(*tui.BubbleProgress).FinalPause = 50 * time.Millisecond
	if p == nil {
		t.Errorf("want progress, got nil")
	}
	if err := p.With(func(pc tui.ProgressControl) error {
		time.Sleep(20 * time.Millisecond)
		for i := 0; i < 42; i++ {
			buf := make([]byte, 1000)
			if _, err := rand.Read(buf); err != nil {
				return err
			}
			if _, err := pc.Write(buf); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		t.Errorf("want nil, got %v", err)
	}

	got := prt.Outputs().Out.String()
	want := "\x1b[?25lmessage  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░   0% ⋮   0.00 B/s   ⋮  41.02 KiB\r\n" +
		"Press Ctrl+C to cancel"
	if !strings.HasPrefix(got, want) {
		t.Errorf("prefix missmatch\nwant %q,\n got %q", want, got)
	}
}
