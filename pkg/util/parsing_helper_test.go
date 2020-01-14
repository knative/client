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
	testMapFromArray(t, []string{"only,split,once", "just,once,"}, ",", map[string]string{"only": "split,once", "just": "once,"})
	testMapFromArray(t, []string{"empty="}, "=", map[string]string{"empty": ""})
}

func testMapFromArray(t *testing.T, input []string, delimiter string, expected map[string]string) {
	actual, err := MapFromArray(input, delimiter)
	assert.NilError(t, err)
	assert.DeepEqual(t, expected, actual)
}

func TestKeyValuePairListAndRemovalListFromArray(t *testing.T) {
	testKeyValuePairListAndRemovalListFromArray(t, []string{"add=value"}, "=", [][]string{{"add", "value"}}, []string{})
	testKeyValuePairListAndRemovalListFromArray(t, []string{"add=value", "remove-"}, "=", [][]string{{"add", "value"}}, []string{"remove"})
}

func testKeyValuePairListAndRemovalListFromArray(t *testing.T, input []string, delimiter string, expectedKVList [][]string, expectedList []string) {
	actualKVList, actualList, err := OrderedMapAndRemovalListFromArray(input, delimiter)
	assert.NilError(t, err)
	assert.DeepEqual(t, NewOrderedMapWithKVStrings(expectedKVList), actualKVList)
	assert.DeepEqual(t, expectedList, actualList)
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

func TestMapFromArrayMapRepeat(t *testing.T) {
	input := []string{"a1=b1", "a1=b2"}
	_, err := MapFromArrayAllowingSingles(input, "=")
	assert.ErrorContains(t, err, "duplicate")
}

func TestMapFromArrayMapKeyEmpty(t *testing.T) {
	input := []string{"=a1"}
	_, err := MapFromArrayAllowingSingles(input, "=")
	assert.ErrorContains(t, err, "empty")
}

func TestParseMinusSuffix(t *testing.T) {
	inputMap := map[string]string{"a1": "b1", "a2-": ""}
	expectedMap := map[string]string{"a1": "b1"}
	expectedStringToRemove := []string{"a2"}
	stringToRemove := ParseMinusSuffix(inputMap)
	assert.DeepEqual(t, expectedMap, inputMap)
	assert.DeepEqual(t, expectedStringToRemove, stringToRemove)
}

func TestStringMap(t *testing.T) {
	inputMap := StringMap{"a1": "b1", "a2": "b2"}
	mergedMap := map[string]string{"a1": "b1-new", "a3": "b3"}
	removedKeys := []string{"a2", "a4"}

	inputMap.Merge(mergedMap).Remove(removedKeys)
	expectedMap := StringMap{"a1": "b1-new", "a3": "b3"}
	assert.DeepEqual(t, expectedMap, inputMap)
}
