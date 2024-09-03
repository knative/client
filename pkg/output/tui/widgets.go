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

package tui

import (
	"context"

	"emperror.dev/errors"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/term"
)

// ErrNotInteractive is returned when the user is not in an interactive session.
var ErrNotInteractive = errors.New("not interactive session")

// Widgets is a set of widgets that can be used to display progress, spinners,
// and other interactive elements.
type Widgets interface {
	// Printf prints a formatted string to the output.
	Printf(format string, a ...any)
	// NewSpinner returns a new spinner with the given message. The spinner will
	// be started when the With method is called and stopped when the function
	// returns.
	NewSpinner(message string) Spinner
	// NewProgress returns a new progress bar with the given total size and
	// message. The progress bar will be started when the With method is called
	// and stopped when the function returns. The progress bar can be updated
	// with the Write method.
	NewProgress(totalSize int, message Message) Progress
}

func NewWidgets(ctx context.Context) Widgets {
	return &widgets{ctx: ctx}
}

type widgets struct {
	ctx context.Context
}

// NewInteractiveWidgets returns a set of interactive widgets if the user is
// in an interactive session. If the user is not in an interactive session,
// ErrNotInteractive error is returned.
func NewInteractiveWidgets(ctx context.Context) (*InteractiveWidgets, error) {
	prt := output.PrinterFrom(ctx)
	if !term.IsReaderTerminal(prt.InOrStdin()) {
		return nil, errors.WithStack(ErrNotInteractive)
	}

	return &InteractiveWidgets{ctx: ctx}, nil
}

type InteractiveWidgets struct {
	ctx context.Context
}
