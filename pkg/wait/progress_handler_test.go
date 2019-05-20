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
	"bytes"
	"errors"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestSimpleProgressHandler(t *testing.T) {
	buf := new(bytes.Buffer)
	ph := NewSimpleProgressHandler(buf, "START", ".", "ERROR", "OK")
	ph.Start()
	assert.Assert(t, strings.Contains(buf.String(), "START"))
	ph.Tic(-1)
	assert.Assert(t, strings.Contains(buf.String(), "."))
	ph.Tic(-1)
	assert.Assert(t, strings.Contains(buf.String(), ".."))
	ph.Tic(8)
	assert.Assert(t, strings.Contains(buf.String(), " 8%"))
	ph.Tic(42)
	assert.Assert(t, strings.Contains(buf.String(), " 8% 42%"), buf.String())
	ph.Fail(errors.New("foobar"))
	assert.Assert(t, strings.Contains(buf.String(), "ERROR\n"))
	assert.Assert(t, !strings.Contains(buf.String(), "foobar"))
	ph.Success()
	assert.Assert(t, strings.Contains(buf.String(), "OK\n"), buf.String())
}
