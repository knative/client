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

package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestMapFromArray(t *testing.T) {
	testMapFromArray(t, []string{"good=value"}, "=", map[string]string{"good": "value"})
	testMapFromArray(t, []string{"multi=value", "other=value"}, "=", map[string]string{"multi": "value", "other": "value"})
	testMapFromArray(t, []string{"over|write", "over|written"}, "|", map[string]string{"over": "written"})
	testMapFromArray(t, []string{"only,split,once", "just,once,"}, ",", map[string]string{"only": "split,once", "just": "once,"})
	testMapFromArray(t, []string{"empty=", "="}, "=", map[string]string{"empty": "", "": ""})
}

func testMapFromArray(t *testing.T, input []string, delimiter string, expected map[string]string) {
	actual, err := MapFromArray(input, delimiter)
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, actual)
}

func TestMapFromArrayNoDelimiter(t *testing.T) {
	input := []string{"badvalue"}
	_, err := MapFromArray(input, "+")
	assert.ErrorContains(t, err, "Argument requires")
	assert.ErrorContains(t, err, "+")
}

func TestMapFromArrayNoDelimiterAllowingSingles(t *testing.T) {
	input := []string{"okvalue"}
	actual, err := MapFromArrayAllowingSingles(input, "+")
	expected := map[string]string{"okvalue": ""}
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, actual)
}

func TestMapFromArrayEmptyValueEmptyDelimiter(t *testing.T) {
	input := []string{""}
	_, err := MapFromArray(input, "")
	assert.ErrorContains(t, err, "Argument requires")
}

func TestMapFromArrayEmptyValueEmptyDelimiterAllowingSingles(t *testing.T) {
	input := []string{""}
	_, err := MapFromArrayAllowingSingles(input, "")
	assert.ErrorContains(t, err, "Argument requires")
}
