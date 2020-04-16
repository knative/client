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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
)

func TestGetAPIServerResourceArray(t *testing.T) {
	t.Run("get single apiserver resource", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1:key1=val1"},
		}
		created, _ := createFlag.getAPIServerVersionKindSelector()

		wanted := []v1alpha2.APIVersionKindSelector{{
			Kind:          "Service",
			APIVersion:    "serving.knative.dev/v1",
			LabelSelector: createLabelSelector("key1", "val1"),
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get single apiserver resource when isController is default", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Service:serving.knative.dev/v1"},
		}
		created, _ := createFlag.getAPIServerVersionKindSelector()
		wanted := []v1alpha2.APIVersionKindSelector{{
			Kind:       "Service",
			APIVersion: "serving.knative.dev/v1",
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get multiple apiserver resources", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1", "Pod:v2:key1=val1,key2=val2"},
		}
		created, _ := createFlag.getAPIServerVersionKindSelector()
		wanted := []v1alpha2.APIVersionKindSelector{{
			Kind:       "Event",
			APIVersion: "v1",
		}, {
			Kind:          "Pod",
			APIVersion:    "v2",
			LabelSelector: createLabelSelector("key1", "val1", "key2", "val2"),
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get multiple apiserver resources without label selector", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1", "Pod:v1"},
		}
		created, _ := createFlag.getAPIServerVersionKindSelector()

		wanted := []v1alpha2.APIVersionKindSelector{{
			Kind:       "Event",
			APIVersion: "v1",
		}, {
			Kind:       "Pod",
			APIVersion: "v1",
		}}
		assert.DeepEqual(t, wanted, created)
	})

	t.Run("get apiserver resource when label controller has error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"Event:v1:xxx,bla"},
		}
		_, err := createFlag.getAPIServerVersionKindSelector()
		errorMsg := "invalid label selector in resource specification Event:v1:xxx,bla (expected: <kind:apiVersion[:label1=val1,label2=val2,..]>"
		assert.Error(t, err, errorMsg)
	})

	t.Run("get apiserver resources when kind has error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{":v2"},
		}
		_, err := createFlag.getAPIServerVersionKindSelector()
		errorMsg := "cannot find 'kind' part in resource specification :v2 (expected: <kind:apiVersion[:label1=val1,label2=val2,..]>"
		assert.Error(t, err, errorMsg)
	})

	t.Run("get apiserver resources when APIVersion has error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Ref",
			Resources:          []string{"kind"},
		}
		_, err := createFlag.getAPIServerVersionKindSelector()
		errorMsg := "cannot find 'APIVersion' part in resource specification kind (expected: <kind:apiVersion[:label1=val1,label2=val2,..]>"
		assert.Error(t, err, errorMsg)
	})
}

func TestGetUpdateAPIServerResourceArray(t *testing.T) {
	t.Run("get removed apiserver resources", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1", "Pod:v2-"},
		}
		added, removed, _ := createFlag.getUpdateAPIVersionKindSelectorArray()
		addwanted := []v1alpha2.APIVersionKindSelector{{
			Kind:       "Event",
			APIVersion: "v1",
		}}
		removewanted := []v1alpha2.APIVersionKindSelector{{
			Kind:       "Pod",
			APIVersion: "v2",
		}}
		assert.DeepEqual(t, added, addwanted)
		assert.DeepEqual(t, removed, removewanted)

		// default api version and isController
		createFlag = APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Event:v1:key1=val1,key2=val2-", "Pod:v1"},
		}
		added, removed, _ = createFlag.getUpdateAPIVersionKindSelectorArray()

		removewanted = []v1alpha2.APIVersionKindSelector{{
			Kind:          "Event",
			APIVersion:    "v1",
			LabelSelector: createLabelSelector("key1", "val1", "key2", "val2"),
		}}
		addwanted = []v1alpha2.APIVersionKindSelector{{
			Kind:       "Pod",
			APIVersion: "v1",
		}}
		assert.DeepEqual(t, added, addwanted)
		assert.DeepEqual(t, removed, removewanted)
	})
}

func TestUpdateExistingAPIServerResourceArray(t *testing.T) {
	existing := []v1alpha2.APIVersionKindSelector{{
		Kind:       "Event",
		APIVersion: "v1",
	}, {
		Kind:       "Pod",
		APIVersion: "v1",
	}}

	t.Run("update existing apiserver resources", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Deployment:v1:key1=val1,key2=val2", "Pod:v1-"},
		}
		updated, _ := createFlag.updateExistingAPIVersionKindSelectorArray(existing)
		updatedWanted := []v1alpha2.APIVersionKindSelector{{
			Kind:       "Event",
			APIVersion: "v1",
		}, {
			Kind:          "Deployment",
			APIVersion:    "v1",
			LabelSelector: createLabelSelector("key1", "val1", "key2", "val2"),
		}}
		assert.DeepEqual(t, updated, updatedWanted)
	})

	t.Run("update existing apiserver resources with error", func(t *testing.T) {
		createFlag := APIServerSourceUpdateFlags{
			ServiceAccountName: "test-sa",
			Mode:               "Resource",
			Resources:          []string{"Deployment:v1", "Pod:v2-"},
		}
		_, err := createFlag.updateExistingAPIVersionKindSelectorArray(existing)
		errorMsg := "cannot find resources to remove: Pod:v2"
		assert.Error(t, err, errorMsg)
	})
}

func createLabelSelector(keyAndVal ...string) *metav1.LabelSelector {
	labels := make(map[string]string)
	for i := 0; i < len(keyAndVal); i += 2 {
		labels[keyAndVal[i]] = keyAndVal[i+1]
	}
	return &metav1.LabelSelector{MatchLabels: labels}
}
