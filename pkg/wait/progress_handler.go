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

package wait

import (
	"fmt"
	"io"
)

// A callback which is informed about the progress of an operation
type ProgressHandler interface {
	// Called when the operation starts
	Start()

	// Called with a percentage indicating completion
	// The argument will be in the range 0..100 if the progress percentage can be calculated
	// or -1 if only an ongoing operation should be indicated
	Tic(complete int)

	// Called when operation finished sucessfully
	Success()

	// Called when the operation failed with the erro
	Fail(err error)
}

// No operation progress handler which just stays silent
type NoopProgressHandler struct{}

func (w NoopProgressHandler) Start()     {}
func (w NoopProgressHandler) Tic(int)    {}
func (w NoopProgressHandler) Fail(error) {}
func (w NoopProgressHandler) Success()   {}

// Standard progress handler for printing out progress to a given writer
type simpleProgressHandler struct {
	out          io.Writer
	startLabel   string
	ticMark      string
	errorLabel   string
	successLabel string
}

// Create a new progress handler which writes out to the given stream.
// The label will be printed right after start, error & success are terminal message depending on the outcome
// In case of an error, the error itself is not printed as it is supposed that the calling function
// will deal with the error and eventually print it out.
func NewSimpleProgressHandler(out io.Writer, label string, tic string, error string, success string) ProgressHandler {
	return &simpleProgressHandler{out, label, tic, error, success}
}

// Print our intial label
func (w *simpleProgressHandler) Start() {
	fmt.Fprint(w.out, w.startLabel)
}

// Tic progress
func (w *simpleProgressHandler) Tic(complete int) {
	if complete < 0 {
		fmt.Fprint(w.out, w.ticMark)
	} else {
		fmt.Fprintf(w.out, " %d%%", complete)
	}
}

// Printout ERROR label, but not the error.
// The error will be printed out later anyway
func (w *simpleProgressHandler) Fail(err error) {
	fmt.Fprintln(w.out, w.errorLabel)
}

// Printout success label
func (w *simpleProgressHandler) Success() {
	fmt.Fprintln(w.out, w.successLabel)
}
