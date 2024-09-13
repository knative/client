/*
 Copyright 2024 The Knative Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package sink

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientdynamic "knative.dev/client/pkg/dynamic"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// ErrSinkIsRequired is returned when no sink is given.
var ErrSinkIsRequired = errors.New("sink is required")

// ErrSinkIsInvalid is returned when the sink has invalid format.
var ErrSinkIsInvalid = errors.New("sink has invalid format")

// Type is a type of Reference.
type Type int

const (
	// TypeURL is a URL version of the sink.
	TypeURL Type = iota
	// TypeReference is a Kuberentes version of the sink.
	TypeReference
)

// Reference represents either a URL or Kubernetes resource.
type Reference struct {
	Type Type
	*KubeReference
	*apis.URL
}

// KubeReference represents a Kubernetes resource as given by command-line args.
type KubeReference struct {
	GVR       schema.GroupVersionResource
	Name      string
	Namespace string
}

// DefaultMappings are used to easily map prefixes for sinks to their
// GroupVersionResources.
var DefaultMappings = withAliasses(map[string]schema.GroupVersionResource{
	"kservice": {
		Resource: "services",
		Group:    "serving.knative.dev",
		Version:  "v1",
	},
	"broker": {
		Resource: "brokers",
		Group:    "eventing.knative.dev",
		Version:  "v1",
	},
	"channel": {
		Resource: "channels",
		Group:    "messaging.knative.dev",
		Version:  "v1",
	},
	"service": { // K8s' service
		Resource: "services",
		Group:    "",
		Version:  "v1",
	},
}, defaultMappingAliasses)

var defaultMappingAliasses = map[string]string{
	knativeServiceShorthand: "kservice",
	"svc":                   "service",
}

const knativeServiceShorthand = "ksvc"

// Resolve returns the Destination referred to by the sink. It validates that
// any object the user is referring to exists.
func (r *Reference) Resolve(ctx context.Context, knclient clientdynamic.KnDynamicClient) (*duckv1.Destination, error) {
	if r.Type == TypeURL {
		return &duckv1.Destination{URI: r.URL}, nil
	}
	if r.Type != TypeReference {
		return nil, fmt.Errorf("%w: unexpected type %q",
			ErrSinkIsInvalid, r.Type)
	}
	client := knclient.RawClient()
	obj, err := client.Resource(r.GVR).
		Namespace(r.Namespace).
		Get(ctx, r.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSinkIsInvalid, err)
	}

	destination := &duckv1.Destination{
		Ref: &duckv1.KReference{
			Kind:       obj.GetKind(),
			APIVersion: obj.GetAPIVersion(),
			Name:       obj.GetName(),
			Namespace:  r.Namespace,
		},
	}
	return destination, nil
}

// String creates a text representation of the reference
// Deprecated: use AsText instead
func (r *Reference) String() string {
	if r == nil {
		return ""
	}
	// unexpected random-like value
	ns := "vaizaeso3sheem5ebie5eeh9Aew5eekei3thie4ezooy9geef6iesh9auPhai7na"
	if r.KubeReference != nil {
		ns = r.Namespace
	}
	return r.AsText(ns)
}

// AsText creates a text representation of the resource, and should
// be used by giving a current namespace.
func (r *Reference) AsText(currentNamespace string) string {
	if r.Type == TypeURL {
		return r.URL.String()
	}
	if r.Type == TypeReference {
		repr := r.GvrAsText() + ":" + r.Name
		if currentNamespace != r.Namespace {
			repr = fmt.Sprintf("%s:%s", repr, r.Namespace)
		}
		return repr
	}
	return fmt.Errorf("%w: unexpected type %q",
		ErrSinkIsInvalid, r.Type).Error()
}

// GvrAsText returns the
func (r *Reference) GvrAsText() string {
	if r == nil || r.KubeReference == nil {
		return fmt.Errorf("%w: unexpected type %#v",
			ErrSinkIsInvalid, r).Error()
	}
	for alias, as := range defaultMappingAliasses {
		if gvr, ok := DefaultMappings[as]; ok && gvr == r.GVR {
			return alias
		}
	}
	for alias, gvr := range DefaultMappings {
		if r.GVR == gvr {
			return alias
		}
	}
	return fmt.Sprintf("%s.%s/%s",
		r.GVR.Resource, r.GVR.Group, r.GVR.Version)

}

// Parse returns the sink reference of given sink representation, which may
// refer to URL or to the Kubernetes resource. The namespace given should be
// the current namespace within the context.
func Parse(sinkRepr, namespace string, mappings map[string]schema.GroupVersionResource) (*Reference, error) {
	if sinkRepr == "" {
		return nil, ErrSinkIsRequired
	}
	prefix, name, ns := parseSink(sinkRepr)
	if prefix == "" {
		// URI target
		uri, err := apis.ParseURL(name)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrSinkIsInvalid, err)
		}
		return &Reference{
			Type: TypeURL,
			URL:  uri,
		}, nil
	}
	gvr, ok := mappings[prefix]
	if !ok {
		idx := strings.LastIndex(prefix, "/")
		var groupVersion string
		var kind string
		if idx != -1 && idx < len(prefix)-1 {
			groupVersion, kind = prefix[:idx], prefix[idx+1:]
		} else {
			kind = prefix
		}
		parsedVersion, err := schema.ParseGroupVersion(groupVersion)
		if err != nil {
			return nil, err
		}

		// For the RAWclient the resource name must be in lower case plural form.
		// This is the best effort to sanitize the inputs, but the safest way is to provide
		// the appropriate form in user's input.
		if !strings.HasSuffix(kind, "s") {
			kind = kind + "s"
		}
		kind = strings.ToLower(kind)
		gvr = parsedVersion.WithResource(kind)
	}
	if ns != "" {
		namespace = ns
	}
	return &Reference{
		Type: TypeReference,
		KubeReference: &KubeReference{
			GVR:       gvr,
			Name:      name,
			Namespace: namespace,
		},
	}, nil
}

// GuessFromDestination converts the duckv1.Destination to the Reference by guessing
// the type by convention.
// Will return nil, if given empty destination.
func GuessFromDestination(dest duckv1.Destination) *Reference {
	if dest.URI != nil {
		return &Reference{
			Type: TypeURL,
			URL:  dest.URI,
		}
	}
	if dest.Ref == nil {
		return nil
	}
	ref := &corev1.ObjectReference{
		Kind:       dest.Ref.Kind,
		Namespace:  dest.Ref.Namespace,
		Name:       dest.Ref.Name,
		APIVersion: dest.Ref.APIVersion,
	}
	gvk := ref.GroupVersionKind()
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	return &Reference{
		Type: TypeReference,
		KubeReference: &KubeReference{
			GVR:       gvr,
			Name:      ref.Name,
			Namespace: ref.Namespace,
		},
	}
}

func withAliasses(
	mappings map[string]schema.GroupVersionResource,
	aliases map[string]string,
) map[string]schema.GroupVersionResource {
	result := make(map[string]schema.GroupVersionResource, len(aliases)+len(mappings))
	for k, v := range mappings {
		result[k] = v
	}
	for as, alias := range aliases {
		if val, ok := result[alias]; ok {
			result[as] = val.GroupResource().WithVersion(val.Version)
		}
	}
	return result
}
