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

package commands

import (
	"fmt"
	"strings"

	"gotest.tools/assert/cmp"
)

// ContainsMultipleSubstring is a comparison utility, compares given substrings against
// target string and returns the gotest.tools/assert/cmp.Comaprison function.
// Provide message to form an error message of format 'Missing $message: $missing_elements'
func ContainsMultipleSubstrings(target string, substrings []string, message string) cmp.Comparison {
	return func() cmp.Result {
		var missing []string
		for _, sub := range substrings {
			if !strings.Contains(target, sub) {
				missing = append(missing, sub)
			}
		}
		if len(missing) > 0 {
			return cmp.ResultFailure(fmt.Sprintf("Missing %s: %s", message, strings.Join(missing[:], ", ")))
		}
		return cmp.ResultSuccess
	}
}
