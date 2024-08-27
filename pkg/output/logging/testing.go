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

package logging

import (
	"log"
	"testing"
)

var fatal = log.Fatal

// WithFatalCaptured runs the given function and captures the arguments passed
// to Fatal function. It is useful in tests to verify that the expected error
// was reported.
func WithFatalCaptured(fn func()) []any {
	if !testing.Testing() {
		panic("WithFatalCaptured should be used only in tests")
	}
	var captured []any
	fatal = func(args ...any) {
		captured = args
	}
	defer func() {
		fatal = log.Fatal
	}()
	fn()
	return captured
}
