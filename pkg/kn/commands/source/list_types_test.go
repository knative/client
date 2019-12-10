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

package source

import (
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"

	"strings"
	"testing"
)

func newUnstructured(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1beta1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"namespace": "current",
				"name":      name,
				"labels": map[string]interface{}{
					"duck.knative.dev/source": "true",
				},
			},
		},
	}
}

func newUnstructuredWithSpecNames(name string, value map[string]interface{}) *unstructured.Unstructured {
	u := newUnstructured(name)
	u.Object["spec"] = map[string]interface{}{"names": value}
	return u
}

func fakeListTypes(args []string, objects ...runtime.Object) (output []string, err error) {
	knParams := &commands.KnParams{}
	// not using the fake dynamic client returned here
	cmd, _, buf := commands.CreateDynamicTestKnCommand(NewSourceCommand(knParams), knParams, objects...)

	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		return
	}

	output = strings.Split(buf.String(), "\n")
	return
}

func TestSourceListTypes(t *testing.T) {
	output, err := fakeListTypes([]string{"source", "list-types"},
		newUnstructuredWithSpecNames("foo.in", map[string]interface{}{"kind": "foo"}),
		newUnstructuredWithSpecNames("bar.in", map[string]interface{}{"kind": "bar"}),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(output[0], "TYPE", "NAME", "DESCRIPTION"))
	assert.Check(t, util.ContainsAll(output[1], "bar", "bar.in"))
	assert.Check(t, util.ContainsAll(output[2], "foo", "foo.in"))
}

func TestSourceListTypesNoHeaders(t *testing.T) {
	output, err := fakeListTypes([]string{"source", "list-types", "--no-headers"},
		newUnstructuredWithSpecNames("foo.in", map[string]interface{}{"kind": "foo"}),
		newUnstructuredWithSpecNames("bar.in", map[string]interface{}{"kind": "bar"}),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsNone(output[0], "TYPE", "NAME", "DESCRIPTION"))
}
