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

package serving

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"knative.dev/pkg/ptr"
	"knative.dev/serving/pkg/apis/autoscaling"
	servingconfig "knative.dev/serving/pkg/apis/config"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"knative.dev/client/pkg/kn/flags"
)

// VolumeSourceType is a type standing for enumeration of ConfigMap and Secret
type VolumeSourceType int

// Enumeration of volume source types: ConfigMap or Secret
const (
	ConfigMapVolumeSourceType VolumeSourceType = iota
	SecretVolumeSourceType
	PortFormatErr = "the port specification '%s' is not valid. Please provide in the format 'NAME:PORT', where 'NAME' is optional. Examples: '--port h2c:8080' , '--port 8080'."
)

var (
	UserImageAnnotationKey = "client.knative.dev/user-image"
	ApiTooOldError         = errors.New("the service is using too old of an API format for the operation")
)

func (vt VolumeSourceType) String() string {
	names := [...]string{"config-map", "secret"}
	if vt < ConfigMapVolumeSourceType || vt > SecretVolumeSourceType {
		return "unknown"
	}
	return names[vt]
}

// UpdateMinScale updates min scale annotation
func UpdateMinScale(template *servingv1.RevisionTemplateSpec, min int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.MinScaleAnnotationKey, strconv.Itoa(min))
}

// UpdateMaxScale updates max scale annotation
func UpdateMaxScale(template *servingv1.RevisionTemplateSpec, max int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.MaxScaleAnnotationKey, strconv.Itoa(max))
}

// UpdateAutoscaleWindow updates the autoscale window annotation
func UpdateAutoscaleWindow(template *servingv1.RevisionTemplateSpec, window string) error {
	_, err := time.ParseDuration(window)
	if err != nil {
		return fmt.Errorf("invalid duration for 'autoscale-window': %v", err)
	}
	return UpdateRevisionTemplateAnnotation(template, autoscaling.WindowAnnotationKey, window)
}

// UpdateConcurrencyTarget updates container concurrency annotation
func UpdateConcurrencyTarget(template *servingv1.RevisionTemplateSpec, target int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.TargetAnnotationKey, strconv.Itoa(target))
}

// UpdateConcurrencyUtilization updates container target utilization percentage annotation
func UpdateConcurrencyUtilization(template *servingv1.RevisionTemplateSpec, target int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.TargetUtilizationPercentageKey, strconv.Itoa(target))
}

// UpdateConcurrencyLimit updates container concurrency limit
func UpdateConcurrencyLimit(template *servingv1.RevisionTemplateSpec, limit int64) error {
	if limit < 0 {
		return fmt.Errorf("invalid concurrency-limit %d (must not be less than 0)", limit)
	}
	template.Spec.ContainerConcurrency = ptr.Int64(limit)
	return nil
}

// UnsetUserImageAnnot removes the user image annotation
func UnsetUserImageAnnot(template *servingv1.RevisionTemplateSpec) {
	delete(template.Annotations, UserImageAnnotationKey)
}

// SetUserImageAnnot sets the user image annotation if the image isn't by-digest already.
func SetUserImageAnnot(template *servingv1.RevisionTemplateSpec) {
	// If the current image isn't by-digest, set the user-image annotation to it
	// so we remember what it was.
	currentContainer, _ := ContainerOfRevisionTemplate(template)
	ui := currentContainer.Image
	if strings.Contains(ui, "@") {
		prev, ok := template.Annotations[UserImageAnnotationKey]
		if ok {
			ui = prev
		}
	}
	if template.Annotations == nil {
		template.Annotations = make(map[string]string)
	}
	template.Annotations[UserImageAnnotationKey] = ui
}

// FreezeImageToDigest sets the image on the template to the image digest of the base revision.
func FreezeImageToDigest(template *servingv1.RevisionTemplateSpec, baseRevision *servingv1.Revision) error {
	if baseRevision == nil {
		return nil
	}

	currentContainer, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}

	baseContainer, err := ContainerOfRevisionSpec(&baseRevision.Spec)
	if err != nil {
		return err
	}
	if currentContainer.Image != baseContainer.Image {
		return fmt.Errorf("could not freeze image to digest since current revision contains unexpected image")
	}

	if baseRevision.Status.DeprecatedImageDigest != "" {
		return flags.UpdateImage(&template.Spec.PodSpec, baseRevision.Status.DeprecatedImageDigest)
	}
	return nil
}

// UpdateLabels updates the labels by adding items from `add` then removing any items from `remove`
func UpdateLabels(labelsMap map[string]string, add map[string]string, remove []string) map[string]string {
	if labelsMap == nil {
		labelsMap = map[string]string{}
	}

	for key, value := range add {
		labelsMap[key] = value
	}
	for _, key := range remove {
		delete(labelsMap, key)
	}

	return labelsMap
}

// UpdateServiceAnnotations updates annotations for the given Service Metadata.
func UpdateServiceAnnotations(service *servingv1.Service, toUpdate map[string]string, toRemove []string) error {
	if service.Annotations == nil && len(toUpdate) > 0 {
		service.Annotations = make(map[string]string)
	}
	return updateAnnotations(service.Annotations, toUpdate, toRemove)
}

// UpdateRevisionTemplateAnnotations updates annotations for the given Revision Template.
// Also validates the autoscaling annotation values
func UpdateRevisionTemplateAnnotations(template *servingv1.RevisionTemplateSpec, toUpdate map[string]string, toRemove []string) error {
	ctx := context.TODO()
	autoscalerConfig := servingconfig.FromContextOrDefaults(ctx).Autoscaler
	autoscalerConfig.AllowZeroInitialScale = true
	if err := autoscaling.ValidateAnnotations(ctx, autoscalerConfig, toUpdate); err != nil {
		return err
	}
	if template.Annotations == nil {
		template.Annotations = make(map[string]string)
	}
	return updateAnnotations(template.Annotations, toUpdate, toRemove)
}

// UpdateRevisionTemplateAnnotation updates an annotation for the given Revision Template.
// Also validates the autoscaling annotation values
func UpdateRevisionTemplateAnnotation(template *servingv1.RevisionTemplateSpec, annotation string, value string) error {
	return UpdateRevisionTemplateAnnotations(template, map[string]string{annotation: value}, []string{})
}

// =======================================================================================

func updateAnnotations(annotations map[string]string, toUpdate map[string]string, toRemove []string) error {
	for key, value := range toUpdate {
		annotations[key] = value
	}
	for _, key := range toRemove {
		delete(annotations, key)
	}
	return nil
}
