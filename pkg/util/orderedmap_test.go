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

func TestOrderedMapCreate(t *testing.T) {
	initial := [][]string{{"1", "v1"}, {"2", "v2"}, {"3", "v3"}}
	o := NewOrderedMapWithKVStrings(initial)
	it := o.Iterator()

	assert.Equal(t, o.Len(), len(initial))
	i := 0

	for k, v, ok := it.NextString(); ok; k, v, ok = it.NextString() {
		assert.Equal(t, k, initial[i][0])
		assert.Equal(t, v, initial[i][1])
		i++
	}
}

func TestOrderedMapSet(t *testing.T) {
	initial := [][]string{{"1", "v1"}, {"2", "v2"}, {"3", "v3"}}
	o := NewOrderedMapWithKVStrings(initial)
	o.Set("4", "v4")
	o.Set("2", "v2-1")

	expected := [][]string{{"1", "v1"}, {"2", "v2-1"}, {"3", "v3"}, {"4", "v4"}}
	assert.Equal(t, o.Len(), len(expected))

	i := 0
	it := o.Iterator()

	for k, v, ok := it.NextString(); ok; k, v, ok = it.NextString() {
		assert.Equal(t, k, expected[i][0])
		assert.Equal(t, v, expected[i][1])
		i++
	}
}

func TestOrderedMapGet(t *testing.T) {
	initial := [][]string{{"1", "v1"}, {"2", "v2"}, {"3", "v3"}}
	o := NewOrderedMapWithKVStrings(initial)
	o.Set("4", "v4")
	o.Set("2", "v2-1")

	expected := [][]string{{"1", "v1"}, {"2", "v2-1"}, {"3", "v3"}, {"4", "v4"}}
	assert.Equal(t, o.Len(), len(expected))

	for i := 0; i < len(expected); i++ {
		assert.Equal(t, o.GetStringWithDefault(expected[i][0], ""), expected[i][1])
	}
}

func TestOrderedMapDelete(t *testing.T) {
	initial := [][]string{{"1", "v1"}, {"2", "v2"}, {"3", "v3"}}
	o := NewOrderedMapWithKVStrings(initial)
	o.Set("4", "v4")
	o.Set("2", "v2-1")
	o.Delete("3")
	o.Delete("1")

	expected := [][]string{{"2", "v2-1"}, {"4", "v4"}}
	assert.Equal(t, o.Len(), len(expected))

	i := 0
	it := o.Iterator()

	for k, v, ok := it.NextString(); ok; k, v, ok = it.NextString() {
		assert.Equal(t, k, expected[i][0])
		assert.Equal(t, v, expected[i][1])
		i++
	}
}
