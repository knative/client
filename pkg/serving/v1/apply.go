package v1

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/util"
)

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

// Helper methods supporting Apply()

// patchStrategy is the patch strategy to us when sending an apply patch
// This is fixed now to a 'normal' JSON patch, as a strategic merge patch is not supported
// (yet) for a custom resource definition.
var patchStrategy = types.MergePatchType

func (cl *knServingClient) patch(modifiedService *servingv1.Service, currentService *servingv1.Service, uOriginalService []byte) (bool, error) {
	uModifiedService, err := getModifiedConfiguration(modifiedService, true)
	if err != nil {
		return false, err
	}
	hasChanged, err := cl.patchSimple(currentService, uModifiedService, uOriginalService)
	for i := 1; i <= 5 && apierrors.IsConflict(err); i++ {
		if i > 1 {
			time.Sleep(1 * time.Second)
		}
		currentService, err = cl.GetService(currentService.Name)
		if err != nil {
			return false, err
		}
		hasChanged, err = cl.patchSimple(currentService, uModifiedService, uOriginalService)
	}

	// TODO: Introduce a "force" parameter that would do the same as `kn service create --force`
	// However, if using a ThreeWayJSONMergePatch, then no conflict can happen, so no need yet for `--force`
	/*
		if err != nil && (apierrors.IsConflict(err) || apierrors.IsInvalid(err)) && p.Force {
			patchBytes, patchService, err = p.deleteAndCreate(current, modified, namespace, name)
		}
	*/
	return hasChanged, err
}

func (cl *knServingClient) patchSimple(currentService *servingv1.Service, uModifiedService []byte, uOriginalService []byte) (bool, error) {
	// Serialize the current configuration of the object from the server.
	uCurrentService, err := encodeService(currentService)
	if err != nil {
		return false, err
	}

	var patch []byte
	if patchStrategy == types.MergePatchType {
		patch, err = jsonmergepatch.CreateThreeWayJSONMergePatch(uOriginalService, uModifiedService, uCurrentService)
	} else if patchStrategy == types.StrategicMergePatchType {
		// Compute a three way strategic merge patch to send to server.
		lookupPatchMeta, err := strategicpatch.NewPatchMetaFromStruct(servingv1.Service{})
		if err != nil {
			return false, err
		}
		// TODO: Allow "overwrite" to be configured
		patch, _ = strategicpatch.CreateThreeWayMergePatch(uOriginalService, uModifiedService, uCurrentService, lookupPatchMeta, false)
	} else {
		return false, fmt.Errorf("unsupported patch type %s", patchStrategy)
	}

	if err != nil {
		return false, err
	}

	if string(patch) == "{}" {
		return false, nil
	}

	// Check if the generation has been counted up, only then the backend detected a change
	savedService, err := cl.PatchService(currentService.Name, patchStrategy, patch)
	if err != nil {
		return false, err
	}
	return savedService.Generation != savedService.Status.ObservedGeneration, nil
}

func getOriginalConfiguration(service *servingv1.Service) []byte {
	annots := service.Annotations
	if annots == nil {
		return nil
	}
	original, ok := annots[v1.LastAppliedConfigAnnotation]
	if !ok {
		return nil
	}
	return []byte(original)
}

func getModifiedConfiguration(service *servingv1.Service, annotate bool) ([]byte, error) {

	// First serialize the object without the annotation to prevent recursion,
	// then add that serialization to it as the annotation and serialize it again.
	var uModifiedService []byte

	// Otherwise, use the server side version of the object.
	// Get the current annotations from the object.
	annots := service.Annotations
	if annots == nil {
		annots = map[string]string{}
	}

	original := annots[v1.LastAppliedConfigAnnotation]
	delete(annots, v1.LastAppliedConfigAnnotation)
	service.Annotations = annots

	uModifiedService, err := encodeService(service)
	if err != nil {
		return nil, err
	}

	if annotate {
		annots[v1.LastAppliedConfigAnnotation] = strings.TrimRight(string(uModifiedService), "\n")

		service.Annotations = annots
		uModifiedService, err = encodeService(service)
		if err != nil {
			return nil, err
		}
	}

	// Restore the object to its original condition.
	annots[v1.LastAppliedConfigAnnotation] = original
	service.Annotations = annots
	return uModifiedService, nil
}

func updateLastAppliedAnnotation(service *servingv1.Service) error {
	annots := service.Annotations
	if annots == nil {
		annots = map[string]string{}
	}
	lastApplied, err := encodeService(service)
	if err != nil {
		return err
	}

	// Cleanup any trailing newlines
	annots[v1.LastAppliedConfigAnnotation] = strings.TrimRight(string(lastApplied), "\n")

	service.Annotations = annots
	return nil
}

func encodeService(service *servingv1.Service) ([]byte, error) {
	scheme := runtime.NewScheme()
	err := servingv1.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}
	factory := serializer.NewCodecFactory(scheme)
	encoder := factory.EncoderForVersion(unstructured.UnstructuredJSONScheme, servingv1.SchemeGroupVersion)
	err = util.UpdateGroupVersionKindWithScheme(service, servingv1.SchemeGroupVersion, scheme)
	if err != nil {
		return nil, err
	}

	serviceUnstructured, err := util.ToUnstructured(service)
	if err != nil {
		return nil, err
	}

	// Remove/adapt service so that it can be used in the apply-annotation
	cleanupServiceUnstructured(serviceUnstructured)

	return runtime.Encode(encoder, serviceUnstructured)
}

func cleanupServiceUnstructured(uService *unstructured.Unstructured) {
	// Set container name if not set
	//setUserContainerIfNotSet(uService.Object)
	clearCreationTimestamps(uService.Object)
	removeStatus(uService.Object)
	removeContainerNameAndResourcesIfNotSet(uService.Object)

}

func removeContainerNameAndResourcesIfNotSet(uService map[string]interface{}) {
	uContainer := extractUserContainer(uService)
	if uContainer == nil {
		return
	}
	name, ok := uContainer["name"]
	if ok && name == "" {
		delete(uContainer, "name")
	}

	resources := uContainer["resources"]
	if resources == nil {
		return
	}
	resourcesMap := resources.(map[string]interface{})
	if len(resourcesMap) == 0 {
		delete(uContainer, "resources")
	}
}

func extractUserContainer(uService map[string]interface{}) map[string]interface{} {
	tSpec := extractTemplateSpec(uService)
	if tSpec == nil {
		return nil
	}
	containers := tSpec["containers"]
	if len(containers.([]interface{})) == 0 {
		return nil
	}
	return containers.([]interface{})[0].(map[string]interface{})
}

func removeStatus(uService map[string]interface{}) {
	delete(uService, "status")
}

func clearCreationTimestamps(uService map[string]interface{}) {
	meta := uService["metadata"]
	if meta != nil {
		delete(meta.(map[string]interface{}), "creationTimestamp")
	}
	template := extractTemplate(uService)
	if template != nil {
		meta = template["metadata"]
		if meta != nil {
			delete(meta.(map[string]interface{}), "creationTimestamp")
		}
	}
}

func extractTemplateSpec(uService map[string]interface{}) map[string]interface{} {
	templ := extractTemplate(uService)
	if templ == nil {
		return nil
	}
	templSpec := templ["spec"]
	if templSpec == nil {
		return nil
	}

	return templSpec.(map[string]interface{})
}

func extractTemplate(uService map[string]interface{}) map[string]interface{} {
	spec := uService["spec"]
	if spec == nil {
		return nil
	}
	templ := spec.(map[string]interface{})["template"]
	if templ == nil {
		return nil
	}
	return templ.(map[string]interface{})
}
