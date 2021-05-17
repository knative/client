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
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type flusher interface {
	Flush() error
}

func NewBarePrefixWriter(out io.Writer) PrefixWriter {
	return &prefixWriter{out: out, nested: nil, colIndent: 0, spaceIndent: 0}
}

// NewPrefixWriter creates a new PrefixWriter.
func NewPrefixWriter(out io.Writer) PrefixWriter {
	tabWriter := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	return &prefixWriter{out: tabWriter, nested: nil, colIndent: 0, spaceIndent: 0}
}

// PrefixWriter can write text at various indentation levels.
type PrefixWriter interface {
	// Write writes text with the specified indentation level.
	Writef(format string, a ...interface{})
	// WriteLine writes an entire line with no indentation level.
	WriteLine(a ...interface{})
	// Write columns with an initial indentation
	WriteCols(cols ...string) PrefixWriter
	// Write columns with an initial indentation and a newline at the end
	WriteColsLn(cols ...string) PrefixWriter
	// Flush forces indentation to be reset.
	Flush() error
	// WriteAttribute writes the attr (as a label) with the given value and returns
	// a PrefixWriter for writing any subattributes.
	WriteAttribute(attr, value string) PrefixWriter
}

// prefixWriter implements PrefixWriter
type prefixWriter struct {
	out         io.Writer
	nested      PrefixWriter
	colIndent   int
	spaceIndent int
}

var _ PrefixWriter = &prefixWriter{}

func (pw *prefixWriter) Writef(format string, a ...interface{}) {
	prefix := ""
	levelSpace := "  "
	for i := 0; i < pw.spaceIndent; i++ {
		prefix += levelSpace
	}
	levelTab := "\t"
	for i := 0; i < pw.colIndent; i++ {
		prefix += levelTab
	}
	if pw.nested != nil {
		pw.nested.Writef(prefix+format, a...)
	} else {
		fmt.Fprintf(pw.out, prefix+format, a...)
	}
}

func (pw *prefixWriter) WriteCols(cols ...string) PrefixWriter {
	ss := make([]string, len(cols))
	for i := range cols {
		ss[i] = "%s"
	}
	format := strings.Join(ss, "\t")
	s := make([]interface{}, len(cols))
	for i, v := range cols {
		s[i] = v
	}

	pw.Writef(format, s...)
	return &prefixWriter{pw.out, pw, 1, 0}
}

// WriteCols writes the columns to the writer and returns a PrefixWriter for
// writing any further parts of the "record" in the last column.
func (pw *prefixWriter) WriteColsLn(cols ...string) PrefixWriter {
	ret := pw.WriteCols(cols...)
	pw.WriteLine()
	return ret
}

func (pw *prefixWriter) WriteLine(a ...interface{}) {
	fmt.Fprintln(pw.out, a...)
}

// WriteAttribute writes the attr (as a label) with the given value and returns
// a PrefixWriter for writing any subattributes.
func (pw *prefixWriter) WriteAttribute(attr, value string) PrefixWriter {
	pw.WriteColsLn(Label(attr), value)
	return &prefixWriter{pw.out, pw, 0, 1}
}

func (pw *prefixWriter) Flush() error {
	if f, ok := pw.out.(flusher); ok {
		return f.Flush()
	}
	return fmt.Errorf("output stream %v doesn't support Flush", pw.out)
}

// Format label (extracted so that color could be added more easily to all labels)
func Label(label string) string {
	return label + ":"
}
