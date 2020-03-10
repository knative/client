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
	"reflect"
	"strings"
	"testing"

	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
)

type containsAllTestCase struct {
	target     string
	substrings []string
	success    bool
	missing    []string
}

type containsNoneTestCase struct {
	target     string
	substrings []string
	success    bool
	contains   []string
}

func TestContainsAll(t *testing.T) {
	for i, tc := range []containsAllTestCase{
		{
			target:     "NAME SERVICE AGE CONDITIONS READY REASON",
			substrings: []string{"REASON", "AGE"},
			success:    true,
		},
		{
			"No resources found.",
			[]string{"NAME", "AGE"},
			false,
			[]string{"NAME", "AGE"},
		},
		{
			"NAME SERVICE AGE CONDITIONS READY REASON",
			[]string{"NAME", "URL", "DOMAIN", "READY"},
			false,
			[]string{"URL", "DOMAIN"},
		},
		{
			target:     "Sword!",
			substrings: []string{},
			success:    true,
		},
	} {
		result := ContainsAll(tc.target, tc.substrings...)()
		if result.Success() != tc.success {
			t.Errorf("%d: Expecting %s to contain %s", i, tc.target, tc.substrings)
		}
		if !tc.success {
			message := fmt.Sprintf("\nActual output: %s\nMissing strings: %s", tc.target, strings.Join(tc.missing[:], ", "))
			if !reflect.DeepEqual(result, cmp.ResultFailure(message)) {
				t.Errorf("%d: Incorrect error message returned\nGot: %v\nExpecting: %s", i, result, message)
			}
		}
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	for i, tc := range []containsAllTestCase{
		{
			target:     "NAME SERVICE AGE CONDITIONS READY REASON",
			substrings: []string{"reason", "age"},
			success:    true,
		},
		{
			"No resources found.",
			[]string{"NAME", "AGE"},
			false,
			[]string{"name", "age"},
		},
		{
			"NAME SERVICE AGE CONDITIONS READY REASON",
			[]string{"name", "url", "domain", "ready"},
			false,
			[]string{"url", "domain"},
		},
	} {
		result := ContainsAllIgnoreCase(tc.target, tc.substrings...)()
		if result.Success() != tc.success {
			t.Errorf("%d: Expecting %s to contain %s", i, tc.target, tc.substrings)
		}
		if !tc.success {
			message := fmt.Sprintf("\nActual output (lower-cased): %s\nMissing strings (lower-cased): %s", strings.ToLower(tc.target), strings.ToLower(strings.Join(tc.missing[:], ", ")))
			if !reflect.DeepEqual(result, cmp.ResultFailure(message)) {
				t.Errorf("%d: Incorrect error message returned\n. Got: %v\nExpecting: %s", i, result, message)
			}
		}
	}
}

func TestContainsNone(t *testing.T) {
	for i, tc := range []containsNoneTestCase{
		{
			target:     "NAME SERVICE AGE CONDITIONS",
			substrings: []string{"REASON", "READY", "foo", "bar"},
			success:    true,
		},
		{
			"NAME SERVICE AGE CONDITIONS READY REASON",
			[]string{"foo", "bar", "NAME", "AGE"},
			false,
			[]string{"NAME", "AGE"},
		},
		{
			"No resources found",
			[]string{"NAME", "URL", "DOMAIN", "READY", "resources"},
			false,
			[]string{"resources"},
		},
		{
			target:     "Sword!",
			substrings: []string{},
			success:    true,
		},
	} {
		comparison := ContainsNone(tc.target, tc.substrings...)
		result := comparison()
		if result.Success() != tc.success {
			t.Errorf("%d: Expecting %s to contain %s", i, tc.target, tc.substrings)
		}
		if !tc.success {
			message := fmt.Sprintf("\nActual output: %s\nContains strings: %s", tc.target, strings.Join(tc.contains[:], ", "))
			if !reflect.DeepEqual(result, cmp.ResultFailure(message)) {
				t.Errorf("%d: Incorrect error message returned\nExpecting: %s", i, message)
			}
		}
	}
}

func TestSliceContainsIgnoreCase(t *testing.T) {
	assert.Equal(t,
		SliceContainsIgnoreCase([]string{"FOO", "bar"}, "foo"),
		true)
	assert.Equal(t,
		SliceContainsIgnoreCase([]string{"BAR", "bar"}, "foo"),
		false)
}
