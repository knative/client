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

package duck

import (
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	sourcesv1alpha2 "knative.dev/eventing/pkg/apis/sources/v1alpha2"
	duck "knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"knative.dev/client/pkg/kn/commands"
	knflags "knative.dev/client/pkg/kn/commands/flags"
)

// Source struct holds common properties between different eventing sources
// which we want to print for commands like 'kn source list'.
// The properties held in this struct is meant for simple access by human readable
// printer function
type Source struct {
	metav1.TypeMeta
	// Name of the created source object
	Name string
	// Namespace of object, used for printing with namespace for example: 'kn source list -A'
	Namespace string
	// Kind of the source object created
	SourceKind string
	// Resource this source object represent
	Resource string
	// Sink configured for this source object
	Sink string
	// String representation if source is ready
	Ready string
}

// SourceList for holding list of Source type objects
type SourceList struct {
	metav1.TypeMeta
	Items []Source
}

const DSListKind = "List"

// GetNamespace returns the namespace of the Source, used for printing
// sources with namespace for commands like 'kn source list -A'
func (s *Source) GetNamespace() string { return s.Namespace }

// DeepCopyObject noop method to satisfy Object interface
func (s *Source) DeepCopyObject() runtime.Object { return s }

// DeepCopyObject noop method to satisfy Object interface
func (s *SourceList) DeepCopyObject() runtime.Object { return s }

// toSource transforms eventing source object received as Unstructured object
// into Source object
func toSource(u *unstructured.Unstructured) Source {
	ds := Source{}
	ds.Name = u.GetName()
	ds.Namespace = u.GetNamespace()
	ds.SourceKind = u.GetKind()
	ds.Resource = getSourceTypeName(u)
	ds.Sink = findSink(u)
	ds.Ready = isReady(u)
	// set empty GVK
	ds.APIVersion, ds.Kind = schema.GroupVersionKind{}.ToAPIVersionAndKind()
	return ds
}

// ToSourceList transforms list of eventing sources objects received as
// UnstructuredList object into SourceList object
func ToSourceList(uList *unstructured.UnstructuredList) *SourceList {
	dsl := SourceList{Items: []Source{}}
	//dsl.Items = make(Source, 0, len(uList.Items))
	for _, u := range uList.Items {
		dsl.Items = append(dsl.Items, toSource(&u))
	}
	// set empty group, version and non empty kind
	dsl.APIVersion, dsl.Kind = schema.GroupVersion{}.WithKind(DSListKind).ToAPIVersionAndKind()
	return &dsl
}

func getSourceTypeName(source *unstructured.Unstructured) string {
	return fmt.Sprintf("%s%s.%s",
		strings.ToLower(source.GetKind()),
		"s",
		strings.Split(source.GetAPIVersion(), "/")[0],
	)
}

func sinkFromUnstructured(u *unstructured.Unstructured) (*duckv1.Destination, error) {
	content := u.UnstructuredContent()
	sink, found, err := unstructured.NestedFieldCopy(content, "spec", "sink")
	if err != nil {
		return nil, fmt.Errorf("cant find sink in given unstructured object at spec.sink field: %v", err)
	}

	if !found {
		return nil, nil
	}

	sinkM, err := json.Marshal(sink)
	if err != nil {
		return nil, fmt.Errorf("error marshaling sink %v: %v", sink, err)
	}

	var sinkD duckv1.Destination
	if err := json.Unmarshal(sinkM, &sinkD); err != nil {
		return nil, fmt.Errorf("failed to unmarshal source sink: %v", err)
	}

	return &sinkD, nil
}

func conditionsFromUnstructured(u *unstructured.Unstructured) (*duckv1.Conditions, error) {
	content := u.UnstructuredContent()
	conds, found, err := unstructured.NestedFieldCopy(content, "status", "conditions")
	if !found || err != nil {
		return nil, fmt.Errorf("cant find conditions in given unstructured object at status.conditions field: %v", err)
	}

	condsM, err := json.Marshal(conds)
	if err != nil {
		return nil, fmt.Errorf("error marshaling conditions %v: %v", conds, err)
	}

	var condsD duckv1.Conditions
	if err := json.Unmarshal(condsM, &condsD); err != nil {
		return nil, fmt.Errorf("failed to unmarshal source status conditions: %v", err)
	}

	return &condsD, nil
}

func findSink(source *unstructured.Unstructured) string {
	switch source.GetKind() {
	case "ApiServerSource":
		var apiSource sourcesv1alpha2.ApiServerSource
		if err := duck.FromUnstructured(source, &apiSource); err == nil {
			return knflags.SinkToString(apiSource.Spec.Sink)
		}
	case "SinkBinding":
		var binding sourcesv1alpha2.SinkBinding
		if err := duck.FromUnstructured(source, &binding); err == nil {
			return knflags.SinkToString(binding.Spec.Sink)
		}
	case "PingSource":
		var pingSource sourcesv1alpha2.PingSource
		if err := duck.FromUnstructured(source, &pingSource); err == nil {
			return knflags.SinkToString(pingSource.Spec.Sink)
		}
	default:
		sink, err := sinkFromUnstructured(source)
		if err != nil {
			return "<unknown>"
		}
		if sink == nil {
			return ""
		}
		return knflags.SinkToString(*sink)
	}
	return "<unknown>"
}

func isReady(source *unstructured.Unstructured) string {
	switch source.GetKind() {
	case "ApiServerSource":
		var tSource sourcesv1alpha2.ApiServerSource
		if err := duck.FromUnstructured(source, &tSource); err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	case "SinkBinding":
		var tSource sourcesv1alpha2.SinkBinding
		if err := duck.FromUnstructured(source, &tSource); err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	case "PingSource":
		var tSource sourcesv1alpha2.PingSource
		if err := duck.FromUnstructured(source, &tSource); err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	default:
		conds, err := conditionsFromUnstructured(source)
		if err != nil {
			// dont throw error in listing: if it cant find the status, return unknown
			return "<unknown>"
		}
		return commands.ReadyCondition(*conds)
	}
	return "<unknown>"
}
