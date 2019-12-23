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

package apiserver

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetAPIServerResourceArray(t *testing.T) {
	t.Run("get single apiserver resource", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1alpha1:true"},
		}
		created, _ := createFlag.getAPIServerResourceArray()
		wanted := []resourceSpec{{
			Kind:         "Service",
			ApiVersion:   "serving.knative.dev/v1alpha1",
			IsController: true,
		}}
		assert.DeepEqual(t, wanted, created)

		// default isController
		createFlag = APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1alpha1"},
		}
		created, _ = createFlag.getAPIServerResourceArray()
		wanted = []resourceSpec{{
			Kind:         "Service",
			ApiVersion:   "serving.knative.dev/v1alpha1",
			IsController: false,
		}}
		assert.DeepEqual(t, wanted, created)

		// default api version and isController
		createFlag = APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:v1"},
		}
		created, _ = createFlag.getAPIServerResourceArray()
		wanted = []resourceSpec{{
			Kind:         "Service",
			ApiVersion:   "v1",
			IsController: false,
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get multiple apiserver resources", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1:true", "Pod:v2:false"},
		}
		created, _ := createFlag.getAPIServerResourceArray()
		wanted := []resourceSpec{{
			Kind:         "Event",
			ApiVersion:   "v1",
			IsController: true,
		}, {
			Kind:         "Pod",
			ApiVersion:   "v2",
			IsController: false,
		}}
		assert.DeepEqual(t, wanted, created)

		// default api version and isController
		createFlag = APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1", "Pod:v1"},
		}
		created, _ = createFlag.getAPIServerResourceArray()

		wanted = []resourceSpec{{
			Kind:         "Event",
			ApiVersion:   "v1",
			IsController: false,
		}, {
			Kind:         "Pod",
			ApiVersion:   "v1",
			IsController: false,
		}}
		assert.DeepEqual(t, wanted, created)
	})
}

func TestGetUpdateAPIServerResourceArray(t *testing.T) {
	t.Run("get removed apiserver resources", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1:true", "Pod:v2:false-"},
		}
		added, removed, _ := createFlag.getUpdateAPIServerResourceArray()
		addwanted := []resourceSpec{{
			Kind:         "Event",
			ApiVersion:   "v1",
			IsController: true,
		}}
		removewanted := []resourceSpec{{
			Kind:         "Pod",
			ApiVersion:   "v2",
			IsController: false,
		}}
		assert.DeepEqual(t, added, addwanted)
		assert.DeepEqual(t, removed, removewanted)

		// default api version and isController
		createFlag = APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1-", "Pod:v1"},
		}
		added, removed, _ = createFlag.getUpdateAPIServerResourceArray()

		removewanted = []resourceSpec{{
			Kind:         "Event",
			ApiVersion:   "v1",
			IsController: false,
		}}
		addwanted = []resourceSpec{{
			Kind:         "Pod",
			ApiVersion:   "v1",
			IsController: false,
		}}
		assert.DeepEqual(t, added, addwanted)
		assert.DeepEqual(t, removed, removewanted)
	})
}
