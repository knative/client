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

package dynamic

import (
	"strings"
	"testing"

	"gotest.tools/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	eventingv1beta1 "knative.dev/eventing/pkg/apis/eventing/v1beta1"
	messagingv1beta1 "knative.dev/eventing/pkg/apis/messaging/v1beta1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/util"
)

const testNamespace = "current"

func TestNamespace(t *testing.T) {
	client := createFakeKnDynamicClient(testNamespace, newSourceCRDObj("foo"))
	assert.Equal(t, client.Namespace(), testNamespace)
}

func TestListCRDs(t *testing.T) {
	client := createFakeKnDynamicClient(
		testNamespace,
		newSourceCRDObj("foo"),
		newSourceCRDObj("bar"),
	)
	assert.Check(t, client.RawClient() != nil)

	t.Run("List CRDs with match", func(t *testing.T) {
		options := metav1.ListOptions{}
		uList, err := client.ListCRDs(options)
		assert.NilError(t, err)
		assert.Equal(t, len(uList.Items), 2)
	})

	t.Run("List CRDs without match", func(t *testing.T) {
		options := metav1.ListOptions{}
		sourcesLabels := labels.Set{"duck.knative.dev/source": "true1"}
		options.LabelSelector = sourcesLabels.String()
		uList, err := client.ListCRDs(options)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(uList.Items), 0)
	})
}

func TestListSourceTypes(t *testing.T) {
	client := createFakeKnDynamicClient(
		testNamespace,
		newSourceCRDObj("foo"),
		newSourceCRDObj("bar"),
	)

	t.Run("List source types", func(t *testing.T) {
		uList, err := client.ListSourcesTypes()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(uList.Items), 2)
		assert.Equal(t, uList.Items[0].GetName(), "foo")
		assert.Equal(t, uList.Items[1].GetName(), "bar")
	})
}

func TestListSources(t *testing.T) {
	t.Run("No GVRs set", func(t *testing.T) {
		obj := newSourceCRDObj("foo")
		client := createFakeKnDynamicClient(testNamespace, obj)
		assert.Check(t, client.RawClient() != nil)
		_, err := client.ListSources()
		assert.Check(t, err != nil)
		assert.Check(t, util.ContainsAll(err.Error(), "can't", "find", "source", "kind", "CRD"))
	})

	t.Run("sources not installed", func(t *testing.T) {
		client := createFakeKnDynamicClient(testNamespace)
		_, err := client.ListSources()
		assert.Check(t, err != nil)
		assert.Check(t, util.ContainsAll(err.Error(), "no sources", "found", "backend", "verify", "installation"))
	})

	t.Run("source list empty", func(t *testing.T) {
		client := createFakeKnDynamicClient(testNamespace,
			newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
		)
		sources, err := client.ListSources()
		assert.NilError(t, err)
		assert.Equal(t, len(sources.Items), 0)
	})

	t.Run("source list non empty", func(t *testing.T) {
		client := createFakeKnDynamicClient(testNamespace,
			newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
			newSourceCRDObjWithSpec("apiserversources", "sources.knative.dev", "v1alpha1", "ApiServerSource"),
			newSourceUnstructuredObj("p1", "sources.knative.dev/v1alpha1", "PingSource"),
			newSourceUnstructuredObj("a1", "sources.knative.dev/v1alpha1", "ApiServerSource"),
			newSourceUnstructuredObj("c1", "sources.knative.dev/v1alpha1", "CronJobSource"),
		)
		sources, err := client.ListSources(WithTypeFilter("pingsource"), WithTypeFilter("ApiServerSource"))
		assert.NilError(t, err)
		assert.Equal(t, len(sources.Items), 2)
		assert.DeepEqual(t, sources.GroupVersionKind(), schema.GroupVersionKind{sourceListGroup, sourceListVersion, sourceListKind})
	})
}

func TestListSourcesUsingGVKs(t *testing.T) {
	t.Run("No GVKs given", func(t *testing.T) {
		client := createFakeKnDynamicClient(testNamespace)
		assert.Check(t, client.RawClient() != nil)
		s, err := client.ListSourcesUsingGVKs(nil)
		assert.NilError(t, err)
		assert.Check(t, s == nil)
	})

	t.Run("source list with given GVKs", func(t *testing.T) {
		client := createFakeKnDynamicClient(testNamespace,
			newSourceCRDObjWithSpec("pingsources", "sources.knative.dev", "v1alpha1", "PingSource"),
			newSourceCRDObjWithSpec("apiserversources", "sources.knative.dev", "v1alpha1", "ApiServerSource"),
			newSourceUnstructuredObj("p1", "sources.knative.dev/v1alpha1", "PingSource"),
			newSourceUnstructuredObj("a1", "sources.knative.dev/v1alpha1", "ApiServerSource"),
		)
		assert.Check(t, client.RawClient() != nil)
		gv := schema.GroupVersion{"sources.knative.dev", "v1alpha1"}
		gvks := []schema.GroupVersionKind{gv.WithKind("ApiServerSource"), gv.WithKind("PingSource")}

		s, err := client.ListSourcesUsingGVKs(&gvks)
		assert.NilError(t, err)
		assert.Check(t, s != nil)
		assert.Equal(t, len(s.Items), 2)
		assert.DeepEqual(t, s.GroupVersionKind(), schema.GroupVersionKind{sourceListGroup, sourceListVersion, sourceListKind})

		// withType
		s, err = client.ListSourcesUsingGVKs(&gvks, WithTypeFilter("PingSource"))
		assert.NilError(t, err)
		assert.Check(t, s != nil)
		assert.Equal(t, len(s.Items), 1)
		assert.DeepEqual(t, s.GroupVersionKind(), schema.GroupVersionKind{sourceListGroup, sourceListVersion, sourceListKind})
	})

}

// createFakeKnDynamicClient gives you a dynamic client for testing containing the given objects.
// See also the one in the fake package. Duplicated here to avoid a dependency loop.
func createFakeKnDynamicClient(testNamespace string, objects ...runtime.Object) KnDynamicClient {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1alpha1", Kind: "Service"}, &servingv1.Service{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "eventing.knative.dev", Version: "v1alpha1", Kind: "Broker"}, &eventingv1beta1.Broker{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "messaging.knative.dev", Version: "v1beta1", Kind: "Channel"}, &messagingv1beta1.Channel{})
	client := dynamicfake.NewSimpleDynamicClient(scheme, objects...)
	return NewKnDynamicClient(client, testNamespace)
}

func newSourceCRDObj(name string) *unstructured.Unstructured {
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
	obj.SetLabels(labels.Set{sourcesLabelKey: sourcesLabelValue})
	return obj
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
		},
	}
}

func newChannelCRDObj(name string) *unstructured.Unstructured {
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
	obj.SetLabels(labels.Set{channelLabelKey: channelLabelValue})
	return obj
}

func newChannelCRDObjWithSpec(name, group, version, kind string) *unstructured.Unstructured {
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
	obj.SetLabels(labels.Set{channelLabelKey: channelLabelValue})
	return obj
}

func newChannelUnstructuredObj(name, apiVersion, kind string) *unstructured.Unstructured {
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
						"name": "foo",
					},
				},
			},
		},
	}
}
func TestListChannelsTypes(t *testing.T) {
	client := createFakeKnDynamicClient(
		testNamespace,
		newChannelCRDObj("foo"),
		newChannelCRDObj("bar"),
	)

	t.Run("List channel types", func(t *testing.T) {
		uList, err := client.ListChannelsTypes()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(uList.Items), 2)
		assert.Equal(t, uList.Items[0].GetName(), "foo")
		assert.Equal(t, uList.Items[1].GetName(), "bar")
	})
}

func TestListChannelsUsingGVKs(t *testing.T) {
	t.Run("No GVKs given", func(t *testing.T) {
		client := createFakeKnDynamicClient(testNamespace)
		assert.Check(t, client.RawClient() != nil)
		s, err := client.ListChannelsUsingGVKs(nil)
		assert.NilError(t, err)
		assert.Check(t, s == nil)
	})

	t.Run("channel list with given GVKs", func(t *testing.T) {
		client := createFakeKnDynamicClient(testNamespace,
			newChannelCRDObjWithSpec("InMemoryChannel", "messaging.knative.dev", "v1beta1", "InMemoryChannel"),
			newChannelUnstructuredObj("i1", "messaging.knative.dev/v1beta1", "InMemoryChannel"),
		)
		assert.Check(t, client.RawClient() != nil)
		gv := schema.GroupVersion{"messaging.knative.dev", "v1beta1"}
		gvks := []schema.GroupVersionKind{gv.WithKind("InMemoryChannel")}

		s, err := client.ListChannelsUsingGVKs(&gvks)
		assert.NilError(t, err)
		assert.Check(t, s != nil)
		assert.Equal(t, len(s.Items), 1)
		assert.DeepEqual(t, s.GroupVersionKind(), schema.GroupVersionKind{channelListGroup, channelListVersion, channelListKind})

		// withType
		s, err = client.ListChannelsUsingGVKs(&gvks, WithTypeFilter("InMemoryChannel"))
		assert.NilError(t, err)
		assert.Check(t, s != nil)
		assert.Equal(t, len(s.Items), 1)
		assert.DeepEqual(t, s.GroupVersionKind(), schema.GroupVersionKind{channelListGroup, channelListVersion, channelListKind})
	})

}
