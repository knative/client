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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	eventinglegacy "knative.dev/eventing/pkg/apis/legacysources/v1alpha1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"
	pkgduck "knative.dev/pkg/apis/duck"
	pkgduckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
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

func findSink(source *unstructured.Unstructured) string {
	switch source.GetKind() {
	case "ApiServerSource":
		var apiSource eventinglegacy.ApiServerSource
		if err := pkgduck.FromUnstructured(source, &apiSource); err == nil {
			return sinkToString(apiSource.Spec.Sink)
		}
	case "CronJobSource":
		var cronSource eventinglegacy.CronJobSource
		if err := pkgduck.FromUnstructured(source, &cronSource); err == nil {
			return sinkToString(cronSource.Spec.Sink)
		}
	case "SinkBinding":
		var binding sourcesv1alpha1.SinkBinding
		if err := pkgduck.FromUnstructured(source, &binding); err == nil {
			return flags.SinkToString(binding.Spec.Sink)
		}
	case "PingSource":
		var pingSource sourcesv1alpha1.PingSource
		if err := pkgduck.FromUnstructured(source, &pingSource); err == nil {
			return flags.SinkToString(*pingSource.Spec.Sink)
		}
	}
	// TODO: Find out how to find sink in untyped sources
	return "<unknown>"
}

func isReady(source *unstructured.Unstructured) string {
	switch source.GetKind() {
	case "ApiServerSource":
		var tSource eventinglegacy.ApiServerSource
		if err := pkgduck.FromUnstructured(source, &tSource); err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	case "CronJobSource":
		var tSource eventinglegacy.CronJobSource
		if err := pkgduck.FromUnstructured(source, &tSource); err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	case "SinkBinding":
		var tSource eventinglegacy.SinkBinding
		if err := pkgduck.FromUnstructured(source, &tSource); err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	case "PingSource":
		var tSource sourcesv1alpha1.PingSource
		if err := pkgduck.FromUnstructured(source, &tSource); err == nil {
			return commands.ReadyCondition(tSource.Status.Conditions)
		}
	}
	// TODO: Find out how to find ready conditions for untyped sources
	return "<unknown>"
}

// temporary sinkToString for deprecated sources
func sinkToString(sink *pkgduckv1beta1.Destination) string {
	if sink != nil {
		if sink.Ref != nil {
			if sink.Ref.Kind == "Service" {
				return fmt.Sprintf("svc:%s", sink.Ref.Name)
			}
			return fmt.Sprintf("%s:%s", strings.ToLower(sink.Ref.Kind), sink.Ref.Name)
		}

		if sink.URI != nil {
			return sink.URI.String()
		}
	}
	return ""
}
