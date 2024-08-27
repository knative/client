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
	"fmt"

	"github.com/erikgeiser/promptkit/selection"
	"knative.dev/client/pkg/output"
	"knative.dev/client/pkg/output/logging"
)

// NewChooser returns a new Chooser.
func NewChooser[T any](iw *InteractiveWidgets) Chooser[T] {
	return &bubbleChooser[T]{iw.ctx}
}

type Chooser[T any] interface {
	Choose(options []T, format string, a ...any) T
}

type bubbleChooser[T any] struct {
	ctx context.Context
}

func (c *bubbleChooser[T]) Choose(options []T, format string, a ...any) T {
	ctx := c.ctx
	prt := output.PrinterFrom(ctx)
	l := logging.LoggerFrom(ctx)
	sel := selection.New(fmt.Sprintf(format, a...), options)
	sel.PageSize = 3
	sel.Input = prt.InOrStdin()
	sel.Output = prt.OutOrStdout()
	chosen, err := sel.RunPrompt()
	if err != nil {
		l.Fatal(err)
	}
	return chosen
}
