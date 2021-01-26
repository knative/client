// Copyright © 2019-2021 The Knative Authors
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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/tracker"
)

// OrderedMapAndRemovalListFromArray creates a list of key-value pair using MapFromArrayAllowingSingles, and a list of removal entries
func OrderedMapAndRemovalListFromArray(arr []string, delimiter string) (*OrderedMap, []string, error) {
	orderedMap := NewOrderedMap()
	removalList := []string{}

	for _, pairStr := range arr {
		pairSlice := strings.SplitN(pairStr, delimiter, 2)
		if len(pairSlice) == 0 || (len(pairSlice) == 1 && !strings.HasSuffix(pairSlice[0], "-")) {
			return nil, nil, fmt.Errorf("argument requires a value that contains the %q character; got %q", delimiter, pairStr)
		}
		key := pairSlice[0]
		if len(pairSlice) == 2 {
			value := pairSlice[1]
			orderedMap.Set(key, value)
		} else {
			// error cases are already filtered out from above part
			removalList = append(removalList, key[:len(key)-1])
		}
	}

	return orderedMap, removalList, nil
}

func MapFromArrayAllowingSingles(arr []string, delimiter string) (map[string]string, error) {
	return mapFromArray(arr, delimiter, true)
}

func MapFromArray(arr []string, delimiter string) (map[string]string, error) {
	return mapFromArray(arr, delimiter, false)
}

func Add(original *map[string]string, toAdd map[string]string, toRemove []string) map[string]string {
	for k, v := range toAdd {
		(*original)[k] = v
	}
	for _, k := range toRemove {
		delete(*original, k)
	}
	return *original
}

func ParseMinusSuffix(m map[string]string) []string {
	stringToRemove := []string{}
	for key := range m {
		if strings.HasSuffix(key, "-") {
			stringToRemove = append(stringToRemove, key[:len(key)-1])
			delete(m, key)
		}
	}
	return stringToRemove
}

// StringMap is a map which key and value are strings
type StringMap map[string]string

// Merge to merge a map to a StringMap
func (m StringMap) Merge(toMerge map[string]string) StringMap {
	for k, v := range toMerge {
		m[k] = v
	}
	return m
}

// Remove to remove from StringMap
func (m StringMap) Remove(toRemove []string) StringMap {
	for _, k := range toRemove {
		delete(m, k)
	}
	return m
}

// AddedAndRemovalListsFromArray returns a list of added entries and a list of removal entries
func AddedAndRemovalListsFromArray(m []string) ([]string, []string) {
	stringToRemove := []string{}
	stringToAdd := []string{}
	for _, key := range m {
		if strings.HasSuffix(key, "-") {
			stringToRemove = append(stringToRemove, key[:len(key)-1])
		} else {
			stringToAdd = append(stringToAdd, key)
		}
	}
	return stringToAdd, stringToRemove
}

// ToTrackerReference will parse a subject in form of kind:apiVersion:name for
// named resources or kind:apiVersion:labelKey1=value1,labelKey2=value2 for
// matching via a label selector.
func ToTrackerReference(subject, namespace string) (*tracker.Reference, error) {
	parts := strings.SplitN(subject, ":", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid subject argument '%s': not in format kind:api/version:nameOrSelector", subject)
	}
	kind := parts[0]
	gv, err := schema.ParseGroupVersion(parts[1])
	if err != nil {
		return nil, err
	}
	reference := &tracker.Reference{
		APIVersion: gv.String(),
		Kind:       kind,
		Namespace:  namespace,
	}
	if !strings.Contains(parts[2], "=") {
		reference.Name = parts[2]
	} else {
		selector, err := ParseSelector(parts[2])
		if err != nil {
			return nil, err
		}
		reference.Selector = &metav1.LabelSelector{MatchLabels: selector}
	}
	return reference, nil
}

// ParseSelector will parse a label selector in form of labelKey1=value1,labelKey2=value2.
func ParseSelector(labelSelector string) (map[string]string, error) {
	selector := map[string]string{}
	for _, p := range strings.Split(labelSelector, ",") {
		keyValue := strings.SplitN(p, "=", 2)
		if len(keyValue) != 2 {
			return nil, fmt.Errorf("invalid subject label selector '%s', expected format: key1=value,key2=value", labelSelector)
		}
		selector[keyValue[0]] = keyValue[1]
	}
	return selector, nil
}

// mapFromArray takes an array of strings where each item is a (key, value) pair
// separated by a delimiter and returns a map where keys are mapped to their respective values.
// If allowSingles is true, values without a delimiter will be added as keys pointing to empty strings
func mapFromArray(arr []string, delimiter string, allowSingles bool) (map[string]string, error) {
	if len(arr) == 0 {
		return nil, nil
	}

	returnMap := map[string]string{}
	for _, pairStr := range arr {
		pairSlice := strings.SplitN(pairStr, delimiter, 2)
		if len(pairSlice) <= 1 {
			if len(pairSlice) == 0 || !allowSingles {
				return nil, fmt.Errorf("Argument requires a value that contains the %q character; got %q", delimiter, pairStr)
			}
			returnMap[pairSlice[0]] = ""
		} else {
			if pairSlice[0] == "" {
				return nil, fmt.Errorf("The key is empty")
			}
			if _, ok := returnMap[pairSlice[0]]; ok {
				return nil, fmt.Errorf("The key %q has been duplicate in %v", pairSlice[0], arr)
			}
			returnMap[pairSlice[0]] = pairSlice[1]
		}
	}
	return returnMap, nil
}
