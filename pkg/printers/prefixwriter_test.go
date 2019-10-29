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

package printers

import (
	"bytes"
	"testing"

	"gotest.tools/assert"
)

func TestWriteColsLn(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewBarePrefixWriter(buf)
	sub := w.WriteColsLn("a", "bbbb", "c", "ddd")
	sub.WriteColsLn("B", "C", "D", "E")
	expected := "a\tbbbb\tc\tddd\n\tB\tC\tD\tE\n"
	actual := buf.String()
	assert.Equal(t, actual, expected)
}

func TestWriteAttribute(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewBarePrefixWriter(buf)
	sub := w.WriteAttribute("Thing", "Stuff")
	sub.WriteColsLn("B", "C", "D", "E")
	expected := "Thing:\tStuff\n  B\tC\tD\tE\n"
	actual := buf.String()
	assert.Equal(t, actual, expected)
}

func TestWriteNested(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewBarePrefixWriter(buf)
	sub := w.WriteColsLn("*", "Header")
	subsub := sub.WriteAttribute("Thing", "stuff")
	subsub.WriteAttribute("Subthing", "substuff")
	expected := "*\tHeader\n\tThing:\tstuff\n\t  Subthing:\tsubstuff\n"
	actual := buf.String()
	assert.Equal(t, actual, expected)
}
