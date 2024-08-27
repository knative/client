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

package output

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func NewTestPrinter() TestPrinter {
	if !testing.Testing() {
		panic("NewTestPrinter() can only be used in tests")
	}
	buf := bytes.NewBufferString("")
	return NewTestPrinterWithInput(buf)
}

func NewTestPrinterWithInput(input io.Reader) TestPrinter {
	if !testing.Testing() {
		panic("NewTestPrinterWithInput() can only be used in tests")
	}
	return TestPrinter{
		stdPrinter{
			testInOut{
				in: input,
				TestOutputs: TestOutputs{
					Out: bytes.NewBufferString(""),
					Err: bytes.NewBufferString(""),
				},
			},
		},
	}
}

func NewTestPrinterWithAnswers(answers []string) TestPrinter {
	return NewTestPrinterWithInput(bytes.NewBufferString(strings.Join(answers, "\n")))
}

type TestPrinter struct {
	stdPrinter
}

func (p TestPrinter) Outputs() TestOutputs {
	return p.InputOutput.(testInOut).TestOutputs //nolint:forcetypeassert
}

type TestOutputs struct {
	Out, Err *bytes.Buffer
}

func (t TestOutputs) OutOrStdout() io.Writer {
	return t.Out
}

func (t TestOutputs) ErrOrStderr() io.Writer {
	return t.Err
}

type testInOut struct {
	in io.Reader
	TestOutputs
}

func (t testInOut) InOrStdin() io.Reader {
	return t.in
}

var _ InputOutput = testInOut{}
