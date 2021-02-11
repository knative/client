// Copyright © 2020 The Knative Authors
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

package dynamic

import (
	"testing"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"knative.dev/client/pkg/util"
)

func TestKindFromUnstructured(t *testing.T) {
	kind, err := kindFromUnstructured(
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
	)
	assert.NilError(t, err)
	assert.Equal(t, kind, "PingSource")
	_, err = kindFromUnstructured(newSourceCRDObj("foo"))
	assert.Check(t, err != nil)
}

func TestGVRFromUnstructured(t *testing.T) {
	obj := newSourceCRDObj("foo")

	obj.Object["spec"] = map[string]interface{}{}
	_, err := gvrFromUnstructured(obj)
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "can't", "find", "group"))

	obj.Object["spec"] = map[string]interface{}{
		"group": "sources.knative.dev",
	}
	_, err = gvrFromUnstructured(obj)
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "can't", "find", "version"))

	// with deprecated CRD field spec version
	obj.Object["spec"] = map[string]interface{}{
		"group":   "sources.knative.dev",
		"version": "v1alpha1",
	}
	_, err = gvrFromUnstructured(obj)
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "can't", "find", "resource"))

	// with CRD field spec versions
	obj.Object["spec"] = map[string]interface{}{
		"group": "sources.knative.dev",
		"versions": []interface{}{
			map[string]interface{}{"name": "v1alpha1", "served": true},
		},
	}
	_, err = gvrFromUnstructured(obj)
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "can't", "find", "resource"))

	obj.Object["spec"] = map[string]interface{}{
		"group": "sources.knative.dev",
		"versions": []interface{}{
			map[string]interface{}{},
		},
	}
	_, err = gvrFromUnstructured(obj)
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "can't", "find", "version"))
}

func TestUnstructuredCRDFromGVK(t *testing.T) {
	u := UnstructuredCRDFromGVK(schema.GroupVersionKind{Group: "sources.knative.dev", Version: "v1alpha2", Kind: "ApiServerSource"})
	g, err := groupFromUnstructured(u)
	assert.NilError(t, err)
	assert.Equal(t, g, "sources.knative.dev")

	v, err := versionFromUnstructured(u)
	assert.NilError(t, err)
	assert.Equal(t, v, "v1alpha2")

	k, err := kindFromUnstructured(u)
	assert.NilError(t, err)
	assert.Equal(t, k, "ApiServerSource")

	r, err := resourceFromUnstructured(u)
	assert.NilError(t, err)
	assert.Equal(t, r, "apiserversources")
}
