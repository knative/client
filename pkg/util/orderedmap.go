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

type valueEntry struct {
	Index int
	Value interface{}
}

type orderedMapIterator struct {
	orderedMap *OrderedMap
	nextIndex  int
}

// OrderedMap is similar implementation of OrderedDict in Python.
type OrderedMap struct {
	Keys     []string
	ValueMap map[string]*valueEntry
}

// NewOrderedMap returns new empty ordered map
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		Keys:     []string{},
		ValueMap: map[string]*valueEntry{},
	}
}

// NewOrderedMapWithKVStrings returns new empty ordered map
func NewOrderedMapWithKVStrings(kvList [][]string) *OrderedMap {
	o := &OrderedMap{
		Keys:     []string{},
		ValueMap: map[string]*valueEntry{},
	}

	for _, pair := range kvList {
		if len(pair) != 2 {
			return nil
		}

		o.Set(pair[0], pair[1])
	}
	return o
}

// Get returns a value corresponding the key
func (o *OrderedMap) Get(key string) (interface{}, bool) {
	ve, ok := o.ValueMap[key]
	if ve != nil {
		return ve.Value, ok
	} else {
		return nil, false
	}
}

// GetString returns a string value corresponding the key
func (o *OrderedMap) GetString(key string) (string, bool) {
	ve, ok := o.ValueMap[key]

	if ve != nil {
		return ve.Value.(string), ok
	} else {
		return "", false
	}
}

// GetStringWithDefault returns a string value corresponding the key if the key is existing.
// Otherwise, the default value is returned.
func (o *OrderedMap) GetStringWithDefault(key string, defaultValue string) string {
	if ve, ok := o.ValueMap[key]; ok {
		return ve.Value.(string)
	} else {
		return defaultValue
	}
}

// Set append the key and value if the key is not existing on the map
// Otherwise, the value does just replace the old value corresponding to the key.
func (o *OrderedMap) Set(key string, value interface{}) {
	if ve, ok := o.ValueMap[key]; !ok {
		o.Keys = append(o.Keys, key)
		o.ValueMap[key] = &valueEntry{
			Index: len(o.Keys) - 1,
			Value: value,
		}
	} else {
		ve.Value = value
	}
}

// Delete deletes the key and value from the map
func (o *OrderedMap) Delete(key string) {
	if ve, ok := o.ValueMap[key]; ok {
		delete(o.ValueMap, key)
		o.Keys = append(o.Keys[:ve.Index], o.Keys[ve.Index+1:]...)
	}
}

// Len returns a size of the ordered map
func (o *OrderedMap) Len() int {
	return len(o.Keys)
}

// Iterator creates a iterator object
func (o *OrderedMap) Iterator() *orderedMapIterator {
	return &orderedMapIterator{
		orderedMap: o,
		nextIndex:  0,
	}
}

// Next returns key and values on current iterating cursor.
// If the cursor moved over last entry, then the third return value will be false, otherwise true.
func (it *orderedMapIterator) Next() (string, interface{}, bool) {
	if it.nextIndex >= it.orderedMap.Len() {
		return "", nil, false
	}

	key := it.orderedMap.Keys[it.nextIndex]
	ve, _ := it.orderedMap.ValueMap[key]

	it.nextIndex++

	return key, ve.Value, true
}

// NextString is the same with Next, but the value is returned as string
func (it *orderedMapIterator) NextString() (string, string, bool) {
	key, value, isValid := it.Next()
	if isValid {
		return key, value.(string), isValid
	} else {
		return "", "", isValid
	}
}
