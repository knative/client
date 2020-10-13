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
	"strings"

	"gotest.tools/assert/cmp"
)

// ContainsAll is a comparison utility, compares given substrings against
// target string and returns the gotest.tools/assert/cmp.Comparison function.
// Provide target string as first arg, followed by any number of substring as args
func ContainsAll(target string, substrings ...string) cmp.Comparison {
	return func() cmp.Result {
		var missing []string
		for _, sub := range substrings {
			if !strings.Contains(target, sub) {
				missing = append(missing, sub)
			}
		}
		if len(missing) > 0 {
			return cmp.ResultFailure(fmt.Sprintf("\nActual output: %s\nMissing strings: %s", target, strings.Join(missing[:], ", ")))
		}
		return cmp.ResultSuccess
	}
}

// Like ContainsAll but ignores the case when checking
func ContainsAllIgnoreCase(target string, substrings ...string) cmp.Comparison {
	return func() cmp.Result {
		var missing []string
		lTarget := strings.ToLower(target)
		for _, sub := range substrings {
			if !strings.Contains(lTarget, strings.ToLower(sub)) {
				missing = append(missing, sub)
			}
		}
		if len(missing) > 0 {
			return cmp.ResultFailure(fmt.Sprintf("\nActual output (lower-cased): %s\nMissing strings (lower-cased): %s", lTarget, strings.ToLower(strings.Join(missing[:], ", "))))
		}
		return cmp.ResultSuccess
	}
}

// ContainsNone is a comparison utility, compares given substrings against
// target string and returns the gotest.tools/assert/cmp.Comparison function.
// Provide target string as first arg, followed by any number of substring as args
func ContainsNone(target string, substrings ...string) cmp.Comparison {
	return func() cmp.Result {
		var contains []string
		for _, sub := range substrings {
			if strings.Contains(target, sub) {
				contains = append(contains, sub)
			}
		}
		if len(contains) > 0 {
			return cmp.ResultFailure(fmt.Sprintf("\nActual output: %s\nContains strings: %s", target, strings.Join(contains[:], ", ")))
		}
		return cmp.ResultSuccess
	}
}

// SliceContainsIgnoreCase checks (case insensitive) if given target string is present in slice
func SliceContainsIgnoreCase(slice []string, target string) bool {
	for _, each := range slice {
		if strings.EqualFold(target, each) {
			return true
		}
	}
	return false
}
