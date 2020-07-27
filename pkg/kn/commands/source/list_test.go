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
	"strings"
	"testing"

	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	dynamicfake "k8s.io/client-go/dynamic/fake"
	clientdynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
)

const (
	crdGroup          = "apiextensions.k8s.io"
	crdVersion        = "v1beta1"
	crdKind           = "CustomResourceDefinition"
	crdKinds          = "customresourcedefinitions"
	sourcesLabelKey   = "duck.knative.dev/source"
	sourcesLabelValue = "true"
	testNamespace     = "current"
)

// sourceFakeCmd takes cmd to be executed using dynamic client
// pass the objects to be registered to dynamic client
func sourceFakeCmd(args []string, objects ...runtime.Object) (output []string, err error) {
	knParams := &commands.KnParams{}
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
	output, err := sourceFakeCmd([]string{"source", "list-types"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
		newSourceCRDObjWithSpec("apiserversources", "sources.knative.dev", "v1alpha1", "ApiServerSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(output[0], "TYPE", "NAME", "DESCRIPTION"))
	assert.Check(t, util.ContainsAll(output[1], "ApiServerSource", "apiserversources"))
	assert.Check(t, util.ContainsAll(output[2], "PingSource", "pingsources"))
}

func TestSourceListTypesNoHeaders(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list-types", "--no-headers"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsNone(output[0], "TYPE", "NAME", "DESCRIPTION"))
	assert.Check(t, util.ContainsAll(output[0], "PingSource"))
}

func TestListBuiltInSources(t *testing.T) {
	fakeDynamic := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	sources, err := listBuiltInSourceTypes(clientdynamic.NewKnDynamicClient(fakeDynamic, "current"))
	assert.NilError(t, err)
	assert.Check(t, sources != nil)
	assert.Equal(t, len(sources.Items), 4)
}

func TestSourceList(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
		newSourceCRDObjWithSpec("sinkbindings", "sources.knative.dev", "v1alpha1", "SinkBinding"),
		newSourceCRDObjWithSpec("apiserversources", "sources.knative.dev", "v1alpha1", "ApiServerSource"),
		newSourceUnstructuredObj("p1", "sources.knative.dev/v1alpha1", "PingSource"),
		newSourceUnstructuredObj("s1", "sources.knative.dev/v1alpha1", "SinkBinding"),
		newSourceUnstructuredObj("a1", "sources.knative.dev/v1alpha1", "ApiServerSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(output[0], "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output[1], "a1", "ApiServerSource", "apiserversources.sources.knative.dev", "ksvc:foo", "True"))
	assert.Check(t, util.ContainsAll(output[2], "p1", "PingSource", "pingsources.sources.knative.dev", "ksvc:foo", "True"))
	assert.Check(t, util.ContainsAll(output[3], "s1", "SinkBinding", "sinkbindings.sources.knative.dev", "ksvc:foo", "True"))
}

func TestSourceListUntyped(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list"},
		newSourceCRDObjWithSpec("kafkasources", "sources.knative.dev", "v1alpha1", "KafkaSource"),
		newSourceUnstructuredObj("k1", "sources.knative.dev/v1alpha1", "KafkaSource"),
		newSourceUnstructuredObj("k2", "sources.knative.dev/v1alpha1", "KafkaSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(output[0], "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output[1], "k1", "KafkaSource", "kafkasources.sources.knative.dev", "ksvc:foo", "True"))
	assert.Check(t, util.ContainsAll(output[2], "k2", "KafkaSource", "kafkasources.sources.knative.dev", "ksvc:foo", "True"))
}

func TestSourceListNoHeaders(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list", "--no-headers"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
		newSourceUnstructuredObj("p1", "sources.knative.dev/v1alpha1", "PingSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsNone(output[0], "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output[0], "p1"))
}

func newSourceCRDObjWithSpec(name, group, version, kind string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crdGroup + "/" + crdVersion,
			"kind":       crdKind,
			"metadata": map[string]interface{}{
				"namespace": testNamespace,
				"name":      name,
			},
		},
	}
	obj.Object["spec"] = map[string]interface{}{
		"group":   group,
		"version": version,
		"names": map[string]interface{}{
			"kind":   kind,
			"plural": strings.ToLower(kind) + "s",
		},
	}
	obj.SetLabels(labels.Set{sourcesLabelKey: sourcesLabelValue})
	return obj
}

func newSourceUnstructuredObj(name, apiVersion, kind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": "current",
				"name":      name,
			},
			"spec": map[string]interface{}{
				"sink": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind": "Service",
						"name": "foo",
					},
				},
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"Type":   "Ready",
						"Status": "True",
					},
				},
			},
		},
	}
}
