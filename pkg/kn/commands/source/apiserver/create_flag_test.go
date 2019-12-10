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
	sources_v1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
)

func TestGetApiServerResourceArray(t *testing.T) {
	t.Run("get single apiserver resource", func(t *testing.T) {
		createFlag := ApiServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1alpha1:true"},
		}
		created := createFlag.GetApiServerResourceArray()
		wanted := []sources_v1alpha1.ApiServerResource{{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
			Controller: true,
		}}
		assert.DeepEqual(t, wanted, created)

		// default isController
		createFlag = ApiServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1alpha1"},
		}
		created = createFlag.GetApiServerResourceArray()
		wanted = []sources_v1alpha1.ApiServerResource{{
			APIVersion: "serving.knative.dev/v1alpha1",
			Kind:       "Service",
			Controller: false,
		}}
		assert.DeepEqual(t, wanted, created)

		// default api version and isController
		createFlag = ApiServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service"},
		}
		created = createFlag.GetApiServerResourceArray()
		wanted = []sources_v1alpha1.ApiServerResource{{
			APIVersion: "v1",
			Kind:       "Service",
			Controller: false,
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get multiple apiserver resources", func(t *testing.T) {
		createFlag := ApiServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1:true", "Pod:v2:false"},
		}
		created := createFlag.GetApiServerResourceArray()
		wanted := []sources_v1alpha1.ApiServerResource{{
			APIVersion: "v1",
			Kind:       "Event",
			Controller: true,
		}, {
			APIVersion: "v2",
			Kind:       "Pod",
			Controller: false,
		}}
		assert.DeepEqual(t, wanted, created)

		// default api version and isController
		createFlag = ApiServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event", "Pod"},
		}
		created = createFlag.GetApiServerResourceArray()

		wanted = []sources_v1alpha1.ApiServerResource{{
			APIVersion: "v1",
			Kind:       "Event",
			Controller: false,
		}, {
			APIVersion: "v1",
			Kind:       "Pod",
			Controller: false,
		}}
		assert.DeepEqual(t, wanted, created)
	})
}
