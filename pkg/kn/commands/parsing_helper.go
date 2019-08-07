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
)

// MapFromArray takes an array of strings where each item is a (key, value) pair
// separated by a delimiter and returns a map where keys are mapped to their respsective values.
func MapFromArray(arr []string, delimiter string) (map[string]string, error) {
	returnMap := map[string]string{}
	for _, pairStr := range arr {
		pairSlice := strings.SplitN(pairStr, delimiter, 2)
		if len(pairSlice) <= 1 {
			return nil, fmt.Errorf("Argument requires a value that contains the %q character; got %q", delimiter, pairStr)
		}
		returnMap[pairSlice[0]] = pairSlice[1]
	}
	return returnMap, nil
}
