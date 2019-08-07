// Copyright © 2018 The Knative Authors
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

package commands

import (
	"reflect"
	"testing"

	"gotest.tools/assert"
)

func TestMapFromArray(t *testing.T) {
	testMapFromArray(t, []string{"good=value"}, "=", map[string]string{"good": "value"})
	testMapFromArray(t, []string{"multi=value", "other=value"}, "=", map[string]string{"multi": "value", "other": "value"})
	testMapFromArray(t, []string{"over|write", "over|written"}, "|", map[string]string{"over": "written"})
	testMapFromArray(t, []string{"only,split,once"}, ",", map[string]string{"only": "split,once"})
}

func testMapFromArray(t *testing.T, input []string, delimiter string, expected map[string]string) {
	actual, err := MapFromArray(input, delimiter, "--flag")
	assert.NilError(t, err)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("Map did not match expected: %s\nFound: %s", expected, actual)
	}
}

func TestMapFromArrayNoDelimiter(t *testing.T) {
	input := []string{"good=value", "badvalue"}
	_, err := MapFromArray(input, "=", "--flag")
	assert.ErrorContains(t, err, "argument requires")
}

func TestMapFromArrayEmptyValue(t *testing.T) {
	input := []string{""}
	_, err := MapFromArray(input, "=", "--flag")
	assert.ErrorContains(t, err, "argument requires")
}
