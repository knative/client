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
	"fmt"
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/tracker"
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

func TestAddedAndRemovalListFromArray(t *testing.T) {
	addList, removeList := AddedAndRemovalListsFromArray([]string{"addvalue1", "remove1-", "addvalue2", "remove2-"})
	assert.DeepEqual(t, []string{"addvalue1", "addvalue2"}, addList)
	assert.DeepEqual(t, []string{"remove1", "remove2"}, removeList)

	addList, removeList = AddedAndRemovalListsFromArray([]string{"remove1-"})
	assert.DeepEqual(t, []string{}, addList)
	assert.DeepEqual(t, []string{"remove1"}, removeList)

	addList, removeList = AddedAndRemovalListsFromArray([]string{"addvalue1"})
	assert.DeepEqual(t, []string{"addvalue1"}, addList)
	assert.DeepEqual(t, []string{}, removeList)
}

func TestToTrackerReference(t *testing.T) {
	testToTrackerReference(t,
		"Broker:eventing.knative.dev/v1beta1:default", "demo",
		&tracker.Reference{
			APIVersion: "eventing.knative.dev/v1beta1",
			Kind:       "Broker",
			Namespace:  "demo",
			Name:       "default",
			Selector:   nil,
		}, nil)
	testToTrackerReference(t,
		"Broker:eventing.knative.dev/v1beta1:default", "",
		&tracker.Reference{
			APIVersion: "eventing.knative.dev/v1beta1",
			Kind:       "Broker",
			Namespace:  "",
			Name:       "default",
			Selector:   nil,
		}, nil)
	testToTrackerReference(t,
		"Job:batch/v1:app=heartbeat-cron,priority=high", "demo",
		&tracker.Reference{
			APIVersion: "batch/v1",
			Kind:       "Job",
			Namespace:  "demo",
			Name:       "",
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":      "heartbeat-cron",
					"priority": "high",
				},
			},
		}, nil)
	testToTrackerReference(t, "", "demo", nil,
		funcRef(func(t *testing.T, err error) {
			assert.ErrorContains(t, err, "not in format kind:api/version:nameOrSelector")
		}))
	testToTrackerReference(t, "Job:batch/v1:app=acme,cmea", "demo", nil,
		funcRef(func(t *testing.T, err error) {
			assert.ErrorContains(t, err, "expected format: key1=value,key2=value")
		}))
	testToTrackerReference(t, "Job:batch/v1/next:acme", "demo", nil,
		funcRef(func(t *testing.T, err error) {
			assert.ErrorContains(t, err, "unexpected GroupVersion string")
		}))
}

func testToTrackerReference(t *testing.T, input, namespace string, expected *tracker.Reference, errMatch *func(*testing.T, error)) {
	t.Helper()
	t.Run(fmt.Sprintf("%s:%s", input, namespace), func(t *testing.T) {
		ref, err := ToTrackerReference(input, namespace)
		if err != nil {
			if errMatch != nil {
				m := *errMatch
				m(t, err)
			} else {
				assert.NilError(t, err, "unexpected error")
			}
		}
		assert.DeepEqual(t, expected, ref)
	})
}

func funcRef(ref func(t *testing.T, err error)) *func(*testing.T, error) {
	return &ref
}
