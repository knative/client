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
	"io"
	"os"
	"testing"

	"gotest.tools/v3/assert"
)

func TestSafeguardBubbletea964(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	assert.NilError(t, os.WriteFile(tmp+"/file", []byte("test"), 0o600))
	td := openFile(t, tmp)
	tf := openFile(t, tmp+"/file")
	tcs := []safeguardBubbletea964TestCase{{
		name: "nil input",
		in:   nil,
		want: nil,
	}, {
		name: "non-regular file",
		in:   os.NewFile(td.Fd(), "/"),
		want: nil,
	}, {
		name: "regular file",
		in:   tf,
		want: tf,
	}, {
		name: "dev null",
		in:   openFile(t, os.DevNull),
		want: nil,
	}, {
		name: "stdin",
		in:   os.Stdin,
		want: bubbletea964Input{Reader: os.Stdin},
	}}
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := safeguardBubbletea964(tc.in)
			assert.Equal(t, got, tc.want)
		})

	}
}

type safeguardBubbletea964TestCase struct {
	name string
	in   io.Reader
	want io.Reader
}

func openFile(tb testing.TB, name string) *os.File {
	tb.Helper()
	f, err := os.Open(name)
	assert.NilError(tb, err)
	tb.Cleanup(func() {
		assert.NilError(tb, f.Close())
	})
	return f
}
