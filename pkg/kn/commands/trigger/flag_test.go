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

package trigger

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetFilter(t *testing.T) {
	t.Run("get multiple filters", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: filterArray{"type=abc.edf.ghi", "attr=value"},
		}
		created := createFlag.GetFilters()
		wanted := map[string]string{
			"type": "abc.edf.ghi",
			"attr": "value",
		}
		assert.DeepEqual(t, wanted, created)
	})
}
