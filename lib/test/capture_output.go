// Copyright 2020 The Knative Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// OutputCapture allows to capture any text written to standard out or standard error
// which is especially useful during testing.
//
// Call it like:
//
// capture := CaptureOutput(t)
// doSomeActionThatWritesToStdOutAndStdErr()
// stdOut, stdErr := capture.Close()
//
// CaptureOutput() and capture.Close() should always come in pairs as Close() also
// restores the old streams
package test

import (
	"io/ioutil"
	"os"
	"testing"

	"gotest.tools/assert"
)

type OutputCapture struct {
	outRead, outWrite     *os.File
	errorRead, errorWrite *os.File
	t                     *testing.T

	oldStdout *os.File
	oldStderr *os.File
}

// CaptureOutput sets up standard our and standard error to capture any
// output which
func CaptureOutput(t *testing.T) OutputCapture {
	ret := OutputCapture{
		oldStdout: os.Stdout,
		oldStderr: os.Stderr,
		t:         t,
	}
	var err error
	ret.outRead, ret.outWrite, err = os.Pipe()
	assert.NilError(t, err)
	os.Stdout = ret.outWrite
	ret.errorRead, ret.errorWrite, err = os.Pipe()
	assert.NilError(t, err)
	os.Stderr = ret.errorWrite
	return ret
}

// Close return the output collected and restores the original standard out and error streams
// (i.e. those that were present before the call to CaptureOutput).
func (c OutputCapture) Close() (string, string) {
	err := c.outWrite.Close()
	assert.NilError(c.t, err)
	err = c.errorWrite.Close()
	assert.NilError(c.t, err)
	outOutput, err := ioutil.ReadAll(c.outRead)
	assert.NilError(c.t, err)
	errOutput, err := ioutil.ReadAll(c.errorRead)
	assert.NilError(c.t, err)
	os.Stdout = c.oldStdout
	os.Stderr = c.oldStderr
	return string(outOutput), string(errOutput)
}
