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

package logging_test

import (
	"errors"
	"testing"

	"knative.dev/client/pkg/context"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/logging"
)

func TestEnsureLogger(t *testing.T) {
	ctx := context.TestContext(t)
	ctx = logging.EnsureLogger(ctx)
	got := logging.LoggerFrom(ctx)
	if got == nil {
		t.Errorf("want logger, got nil")
	}
	got.Debug("test")
}

func TestEnsureLoggerWithoutTestingT(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("FORCE_COLOR", "true")
	ctx := context.TODO()
	printer := output.NewTestPrinter()
	ctx = output.WithContext(ctx, printer)
	ctx = logging.EnsureLogger(ctx)

	l := logging.LoggerFrom(ctx)
	if l == nil {
		t.Errorf("want logger, got nil")
	}
	l.Debug("test")
	out := printer.Outputs()
	got := out.Err.String()
	want := "0 \x1b[35mDEBUG\x1b[0m test\n"
	if got != want {
		t.Errorf("\nwant %q,\n got %q", want, got)
	}
}

func TestLoggerFrom(t *testing.T) {
	ctx := context.TestContext(t)
	args := logging.WithFatalCaptured(func() {
		logging.LoggerFrom(ctx)
	})
	if len(args) != 1 {
		t.Errorf("want 1 arg, got %d", len(args))
	}
	err := args[0].(error)
	if !errors.Is(err, logging.ErrCallEnsureLoggerFirst) {
		t.Errorf("want %v, got %v", logging.ErrCallEnsureLoggerFirst, err)
	}
}
