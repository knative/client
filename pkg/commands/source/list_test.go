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
	"context"
	"strings"
	"testing"

	"knative.dev/client-pkg/pkg/dynamic/fake"

	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"knative.dev/client-pkg/pkg/util"
	"knative.dev/client/pkg/commands"
)

const (
	crdGroup          = "apiextensions.k8s.io"
	crdVersion        = "v1"
	crdKind           = "CustomResourceDefinition"
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

func TestSourceListTypesNoSourcesInstalled(t *testing.T) {
	_, err := sourceFakeCmd([]string{"source", "list-types"})
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "no", "Knative Sources", "found", "backend", "verify", "installation"))
}

func TestSourceListTypesNoSourcesWithJsonOutput(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list-types", "-o", "json"},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apiextensions.k8s.io/v1",
				"kind":       "CustomResourceDefinitionList",
			},
		})
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(strings.Join(output[:], "\n"), "\"apiVersion\": \"apiextensions.k8s.io/v1\"", "\"items\": []", "\"kind\": \"CustomResourceDefinitionList\""))
}

func TestSourceListTypes(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list-types"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1beta2", "PingSource"),
		newSourceCRDObjWithSpec("apiserversources", "sources.knative.dev", "v1", "ApiServerSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(output[0], "TYPE", "S", "NAME", "DESCRIPTION"))
	assert.Check(t, util.ContainsAll(output[1], "ApiServerSource", "X", "apiserversources"))
	assert.Check(t, util.ContainsAll(output[2], "PingSource", "X", "pingsources"))
}

func TestSourceListTypesNoHeaders(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list-types", "--no-headers"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1", "PingSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsNone(output[0], "TYPE", "NAME", "DESCRIPTION"))
	assert.Check(t, util.ContainsAll(output[0], "PingSource"))
}

func TestListBuiltInSourceTypes(t *testing.T) {
	sources, err := listBuiltInSourceTypes(context.Background(), fake.CreateFakeKnDynamicClient("current"))
	assert.NilError(t, err)
	if sources == nil {
		t.Fatal("sources = nil, want not nil")
	}
	assert.Equal(t, len(sources.Items), 4)
}

func TestSourceListNoSourcesInstalled(t *testing.T) {
	_, err := sourceFakeCmd([]string{"source", "list"})
	assert.Check(t, err != nil)
	assert.Check(t, util.ContainsAll(err.Error(), "no sources", "found", "backend", "verify", "installation"))
}

func TestSourceListEmpty(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list", "-o", "json"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1beta2", "PingSource"),
	)
	assert.NilError(t, err)
	outputJson := strings.Join(output[:], "\n")
	assert.Assert(t, util.ContainsAll(outputJson, "\"apiVersion\": \"client.knative.dev/v1alpha1\"", "\"items\": [],", "\"kind\": \"SourceList\""))
}

func TestSourceList(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1beta2", "PingSource"),
		newSourceCRDObjWithSpec("sinkbindings", "sources.knative.dev", "v1", "SinkBinding"),
		newSourceCRDObjWithSpec("apiserversources", "sources.knative.dev", "v1", "ApiServerSource"),
		newSourceUnstructuredObj("p1", "sources.knative.dev/v1beta2", "PingSource"),
		newSourceUnstructuredObj("s1", "sources.knative.dev/v1", "SinkBinding"),
		newSourceUnstructuredObj("a1", "sources.knative.dev/v1", "ApiServerSource"),
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
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1beta2", "PingSource"),
		newSourceUnstructuredObj("p1", "sources.knative.dev/v1beta2", "PingSource"),
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
						"apiVersion": "serving.knative.dev/v1",
						"kind":       "Service",
						"name":       "foo",
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

func TestSourceListAllNamespace(t *testing.T) {
	output, err := sourceFakeCmd([]string{"source", "list", "--all-namespaces"},
		newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1beta2", "PingSource"),
		newSourceCRDObjWithSpec("sinkbindings", "sources.knative.dev", "v1", "SinkBinding"),
		newSourceCRDObjWithSpec("apiserversources", "sources.knative.dev", "v1", "ApiServerSource"),
		newSourceUnstructuredObj("p1", "sources.knative.dev/v1beta2", "PingSource"),
		newSourceUnstructuredObj("s1", "sources.knative.dev/v1", "SinkBinding"),
		newSourceUnstructuredObj("a1", "sources.knative.dev/v1", "ApiServerSource"),
	)
	assert.NilError(t, err)
	assert.Check(t, util.ContainsAll(output[0], "NAMESPACE", "NAME", "TYPE", "RESOURCE", "SINK", "READY"))
	assert.Check(t, util.ContainsAll(output[1], "current", "a1", "ApiServerSource", "apiserversources.sources.knative.dev", "ksvc:foo", "True"))
	assert.Check(t, util.ContainsAll(output[2], "current", "p1", "PingSource", "pingsources.sources.knative.dev", "ksvc:foo", "True"))
	assert.Check(t, util.ContainsAll(output[3], "current", "s1", "SinkBinding", "sinkbindings.sources.knative.dev", "ksvc:foo", "True"))
}
