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
	"sort"
	"testing"

	"gotest.tools/assert"
)

func TestGetFilters(t *testing.T) {
	t.Run("get multiple filters", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: []string{"type=abc.edf.ghi", "attr=value"},
		}
		created, err := createFlag.GetFilters()
		wanted := map[string]string{
			"type": "abc.edf.ghi",
			"attr": "value",
		}
		assert.NilError(t, err, "Filter should be created")
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get filters with errors", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: []string{"type"},
		}
		_, err := createFlag.GetFilters()
		assert.ErrorContains(t, err, "Invalid --filter")

		createFlag = TriggerUpdateFlags{
			Filters: []string{"type="},
		}
		filters, _ := createFlag.GetFilters()
		wanted := map[string]string{"type": ""}
		assert.DeepEqual(t, wanted, filters)

		createFlag = TriggerUpdateFlags{
			Filters: []string{"=value"},
		}
		_, err = createFlag.GetFilters()
		assert.ErrorContains(t, err, "Invalid --filter")

		createFlag = TriggerUpdateFlags{
			Filters: []string{"="},
		}
		_, err = createFlag.GetFilters()
		assert.ErrorContains(t, err, "Invalid --filter")
	})

	t.Run("get duplicate filters", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: []string{"type=foo", "type=bar"},
		}
		_, err := createFlag.GetFilters()
		assert.ErrorContains(t, err, "duplicate")
	})
}

func TestGetUpdateFilters(t *testing.T) {
	t.Run("get updated filters", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: []string{"type=abc.edf.ghi", "attr=value"},
		}
		updated, removed, err := createFlag.GetUpdateFilters()
		wanted := map[string]string{
			"type": "abc.edf.ghi",
			"attr": "value",
		}
		assert.NilError(t, err, "UpdateFilter should be created")
		assert.DeepEqual(t, wanted, updated)
		assert.Assert(t, len(removed) == 0)
	})

	t.Run("get deleted filters", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: []string{"type-", "attr-"},
		}
		updated, removed, err := createFlag.GetUpdateFilters()
		wanted := []string{"type", "attr"}
		sort.Strings(wanted)
		sort.Strings(removed)
		assert.NilError(t, err, "UpdateFilter should be created")
		assert.DeepEqual(t, wanted, removed)
		assert.Assert(t, len(updated) == 0)
	})

	t.Run("get updated & deleted filters", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: []string{"type=foo", "attr-", "source=bar", "env-"},
		}
		updated, removed, err := createFlag.GetUpdateFilters()
		wantedRemoved := []string{"attr", "env"}
		wantedUpdated := map[string]string{
			"type":   "foo",
			"source": "bar",
		}
		sort.Strings(wantedRemoved)
		sort.Strings(removed)
		assert.NilError(t, err, "UpdateFilter should be created")
		assert.DeepEqual(t, wantedRemoved, removed)
		assert.DeepEqual(t, wantedUpdated, updated)
	})

	t.Run("update duplicate filters", func(t *testing.T) {
		createFlag := TriggerUpdateFlags{
			Filters: []string{"type=foo", "type=bar"},
		}
		_, _, err := createFlag.GetUpdateFilters()
		assert.ErrorContains(t, err, "duplicate")
	})
}
