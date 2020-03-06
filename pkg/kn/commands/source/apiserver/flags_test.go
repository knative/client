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
	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
)

func TestGetAPIServerResourceArray(t *testing.T) {
	t.Run("get single apiserver resource", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1:true"},
		}
		created, _ := createFlag.getAPIServerResourceArray()
		wanted := []v1alpha1.ApiServerResource{{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
			Controller: true,
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get single apiserver resource when isController is default", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1"},
		}
		created, _ := createFlag.getAPIServerResourceArray()
		wanted := []v1alpha1.ApiServerResource{{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
			Controller: false,
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
		wanted := []v1alpha1.ApiServerResource{{
			Kind:       "Event",
			APIVersion: "v1",
			Controller: true,
		}, {
			Kind:       "Pod",
			APIVersion: "v2",
			Controller: false,
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get multiple apiserver resources when isController is default", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1", "Pod:v1"},
		}
		created, _ := createFlag.getAPIServerResourceArray()

		wanted := []v1alpha1.ApiServerResource{{
			Kind:       "Event",
			APIVersion: "v1",
			Controller: false,
		}, {
			Kind:       "Pod",
			APIVersion: "v1",
			Controller: false,
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get apiserver resource when isController has error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Event:v1:xxx"},
		}
		_, err := createFlag.getAPIServerResourceArray()
		errorMsg := "controller flag is not a boolean in resource specification Event:v1:xxx (expected: <Kind:ApiVersion[:controllerFlag]>)"
		assert.Error(t, err, errorMsg)
	})

	t.Run("get apiserver resources when kind has error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{":v2:true"},
		}
		_, err := createFlag.getAPIServerResourceArray()
		errorMsg := "cannot find 'Kind' part in resource specification :v2:true (expected: <Kind:ApiVersion[:controllerFlag]>"
		assert.Error(t, err, errorMsg)
	})

	t.Run("get apiserver resources when APIVersion has error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"kind"},
		}
		_, err := createFlag.getAPIServerResourceArray()
		errorMsg := "cannot find 'APIVersion' part in resource specification kind (expected: <Kind:ApiVersion[:controllerFlag]>"
		assert.Error(t, err, errorMsg)
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
		addwanted := []v1alpha1.ApiServerResource{{
			Kind:       "Event",
			APIVersion: "v1",
			Controller: true,
		}}
		removewanted := []v1alpha1.ApiServerResource{{
			Kind:       "Pod",
			APIVersion: "v2",
			Controller: false,
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

		removewanted = []v1alpha1.ApiServerResource{{
			Kind:       "Event",
			APIVersion: "v1",
			Controller: false,
		}}
		addwanted = []v1alpha1.ApiServerResource{{
			Kind:       "Pod",
			APIVersion: "v1",
			Controller: false,
		}}
		assert.DeepEqual(t, added, addwanted)
		assert.DeepEqual(t, removed, removewanted)
	})
}

func TestUpdateExistingAPIServerResourceArray(t *testing.T) {
	existing := []v1alpha1.ApiServerResource{{
		Kind:       "Event",
		APIVersion: "v1",
		Controller: false,
	}, {
		Kind:       "Pod",
		APIVersion: "v1",
		Controller: false,
	}}

	t.Run("update existing apiserver resources", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Deployment:v1:true", "Pod:v1:false-"},
		}
		updated, _ := createFlag.updateExistingAPIServerResourceArray(existing)
		updatedWanted := []v1alpha1.ApiServerResource{{
			Kind:       "Event",
			APIVersion: "v1",
			Controller: false,
		}, {
			Kind:       "Deployment",
			APIVersion: "v1",
			Controller: true,
		}}
		assert.DeepEqual(t, updated, updatedWanted)
	})

	t.Run("update existing apiserver resources with error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Deployment:v1:true", "Pod:v2:false-"},
		}
		_, err := createFlag.updateExistingAPIServerResourceArray(existing)
		errorMsg := "cannot find resource Pod:v2:false to remove"
		assert.Error(t, err, errorMsg)
	})
}
