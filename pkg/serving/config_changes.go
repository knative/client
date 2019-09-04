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
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/ptr"
	"knative.dev/serving/pkg/apis/autoscaling"
	"knative.dev/serving/pkg/apis/serving"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// VolumeSourceType is a type standing for enumeration of ConfigMap and Secret
type VolumeSourceType int

// Enumeration of volume source types: ConfigMap or Secret
const (
	ConfigMapVolumeSourceType VolumeSourceType = iota
	SecretVolumeSourceType
)

var UserImageAnnotationKey = "client.knative.dev/user-image"

// UpdateEnvVars gives the configuration all the env var values listed in the given map of
// vars.  Does not touch any environment variables not mentioned, but it can add
// new env vars and change the values of existing ones, then sort by env key name.
func UpdateEnvVars(template *servingv1alpha1.RevisionTemplateSpec, toUpdate map[string]string, toRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	updated := updateEnvVarsFromMap(container.Env, toUpdate)
	updated = removeEnvVars(updated, toRemove)
	// Sort by env key name
	sort.SliceStable(updated, func(i, j int) bool {
		return updated[i].Name < updated[j].Name
	})
	container.Env = updated

	return nil
}

// UpdateEnvFrom updates envFrom
func UpdateEnvFrom(template *servingv1alpha1.RevisionTemplateSpec, toUpdate []string, toRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	envFrom, err := updateEnvFrom(container.EnvFrom, toUpdate)
	if err != nil {
		return err
	}
	container.EnvFrom, err = removeEnvFrom(envFrom, toRemove)
	return err
}

// UpdateVolumeMountsAndVolumes updates the configuration for volume mounts and volumes.
func UpdateVolumeMountsAndVolumes(template *servingv1alpha1.RevisionTemplateSpec,
	mountsToUpdate map[string]string, mountsToRemove []string, volumesToUpdate map[string]string, volumesToRemove []string) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}

	volumeSourceInfoByName := make(map[string]*volumeSourceInfo)
	for name, value := range volumesToUpdate {
		info, err := parseVolumeSourceString(value)
		if err != nil {
			return err
		}
		volumeSourceInfoByName[name] = info
	}

	volumes, err := updateVolumesFromMap(template.Spec.Volumes, volumeSourceInfoByName)
	if err != nil {
		return err
	}

	volumeMounts, err := updateVolumeMountsFromMap(container.VolumeMounts, mountsToUpdate, volumes)
	if err != nil {
		return err
	}

	container.VolumeMounts = removeVolumeMounts(volumeMounts, mountsToRemove)
	template.Spec.Volumes, err = removeVolumes(volumes, volumesToRemove, container.VolumeMounts)

	return err
}

// UpdateMinScale updates min scale annotation
func UpdateMinScale(template *servingv1alpha1.RevisionTemplateSpec, min int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.MinScaleAnnotationKey, strconv.Itoa(min))
}

// UpdatMaxScale updates max scale annotation
func UpdateMaxScale(template *servingv1alpha1.RevisionTemplateSpec, max int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.MaxScaleAnnotationKey, strconv.Itoa(max))
}

// UpdateConcurrencyTarget updates container concurrency annotation
func UpdateConcurrencyTarget(template *servingv1alpha1.RevisionTemplateSpec, target int) error {
	return UpdateRevisionTemplateAnnotation(template, autoscaling.TargetAnnotationKey, strconv.Itoa(target))
}

// UpdateConcurrencyLimit updates container concurrency limit
func UpdateConcurrencyLimit(template *servingv1alpha1.RevisionTemplateSpec, limit int64) error {
	err := serving.ValidateContainerConcurrency(ptr.Int64(limit)).ViaField("spec.containerConcurrency")
	if err != nil {
		return fmt.Errorf("invalid 'concurrency-limit' value: %s", err)
	}

	template.Spec.ContainerConcurrency = ptr.Int64(limit)
	return nil
}

// UpdateRevisionTemplateAnnotation updates an annotation for the given Revision Template.
// Also validates the autoscaling annotation values
func UpdateRevisionTemplateAnnotation(template *servingv1alpha1.RevisionTemplateSpec, annotation string, value string) error {
	annoMap := template.Annotations
	if annoMap == nil {
		annoMap = make(map[string]string)
		template.Annotations = annoMap
	}

	// Validate autoscaling annotations and returns error if invalid input provided
	// without changing the existing spec
	in := make(map[string]string)
	in[annotation] = value
	if err := autoscaling.ValidateAnnotations(in); err != nil {
		return err
	}

	annoMap[annotation] = value
	return nil
}

var ApiTooOldError = errors.New("the service is using too old of an API format for the operation")

// UpdateName updates the revision name.
func UpdateName(template *servingv1alpha1.RevisionTemplateSpec, name string) error {
	if template.Spec.DeprecatedContainer != nil {
		return ApiTooOldError
	}
	template.Name = name
	return nil
}

// EnvToMap is an utility function to translate between the API list form of env vars, and the
// more convenient map form.
func EnvToMap(vars []corev1.EnvVar) (map[string]string, error) {
	result := map[string]string{}
	for _, envVar := range vars {
		_, present := result[envVar.Name]
		if present {
			return nil, fmt.Errorf("env var name present more than once: %v", envVar.Name)
		}
		result[envVar.Name] = envVar.Value
	}
	return result, nil
}

// UpdateImage a given image
func UpdateImage(template *servingv1alpha1.RevisionTemplateSpec, image string) error {
	// When not setting the image to a digest, add the user image annotation.
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Image = image
	return nil
}

// UnsetUserImageAnnot removes the user image annotation
func UnsetUserImageAnnot(template *servingv1alpha1.RevisionTemplateSpec) {
	delete(template.Annotations, UserImageAnnotationKey)
}

// SetUserImageAnnot sets the user image annotation if the image isn't by-digest already.
func SetUserImageAnnot(template *servingv1alpha1.RevisionTemplateSpec) {
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
func FreezeImageToDigest(template *servingv1alpha1.RevisionTemplateSpec, baseRevision *servingv1alpha1.Revision) error {
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
		return fmt.Errorf("could not freeze image to digest since current revision contains unexpected image.")
	}

	if baseRevision.Status.ImageDigest != "" {
		return UpdateImage(template, baseRevision.Status.ImageDigest)
	}
	return nil
}

// UpdateContainerPort updates container with a give port
func UpdateContainerPort(template *servingv1alpha1.RevisionTemplateSpec, port int32) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	container.Ports = []corev1.ContainerPort{{
		ContainerPort: port,
	}}
	return nil
}

// UpdateResources updates resources as requested
func UpdateResources(template *servingv1alpha1.RevisionTemplateSpec, requestsResourceList corev1.ResourceList, limitsResourceList corev1.ResourceList) error {
	container, err := ContainerOfRevisionTemplate(template)
	if err != nil {
		return err
	}
	if container.Resources.Requests == nil {
		container.Resources.Requests = corev1.ResourceList{}
	}

	for k, v := range requestsResourceList {
		container.Resources.Requests[k] = v
	}

	if container.Resources.Limits == nil {
		container.Resources.Limits = corev1.ResourceList{}
	}

	for k, v := range limitsResourceList {
		container.Resources.Limits[k] = v
	}

	return nil
}

// UpdateLabels updates the labels identically on a service and template.
// Does not overwrite the entire Labels field, only makes the requested updates
func UpdateLabels(service *servingv1alpha1.Service, template *servingv1alpha1.RevisionTemplateSpec, toUpdate map[string]string, toRemove []string) error {
	if service.ObjectMeta.Labels == nil {
		service.ObjectMeta.Labels = make(map[string]string)
	}
	if template.ObjectMeta.Labels == nil {
		template.ObjectMeta.Labels = make(map[string]string)
	}
	for key, value := range toUpdate {
		service.ObjectMeta.Labels[key] = value
		template.ObjectMeta.Labels[key] = value
	}
	for _, key := range toRemove {
		delete(service.ObjectMeta.Labels, key)
		delete(template.ObjectMeta.Labels, key)
	}
	return nil
}

// UpdateAnnotations updates the annotations identically on a service and template.
// Does not overwrite the entire Annotations field, only makes the requested updates.
func UpdateAnnotations(
	service *servingv1alpha1.Service,
	template *servingv1alpha1.RevisionTemplateSpec,
	toUpdate map[string]string,
	toRemove []string) error {

	if service.ObjectMeta.Annotations == nil {
		service.ObjectMeta.Annotations = make(map[string]string)
	}

	if template.ObjectMeta.Annotations == nil {
		template.ObjectMeta.Annotations = make(map[string]string)
	}

	for key, value := range toUpdate {
		service.ObjectMeta.Annotations[key] = value
		template.ObjectMeta.Annotations[key] = value
	}

	for _, key := range toRemove {
		delete(service.ObjectMeta.Annotations, key)
		delete(template.ObjectMeta.Annotations, key)
	}

	return nil
}

// UpdateServiceAccountName updates the service account name used for the corresponding knative service
func UpdateServiceAccountName(template *servingv1alpha1.RevisionTemplateSpec, serviceAccountName string) error {
	serviceAccountName = strings.TrimSpace(serviceAccountName)
	template.Spec.ServiceAccountName = serviceAccountName
	return nil
}

// =======================================================================================

func updateEnvVarsFromMap(env []corev1.EnvVar, toUpdate map[string]string) []corev1.EnvVar {
	set := sets.NewString()
	for i := range env {
		envVar := &env[i]
		if val, ok := toUpdate[envVar.Name]; ok {
			envVar.Value = val
			set.Insert(envVar.Name)
		}
	}
	for name, val := range toUpdate {
		if !set.Has(name) {
			env = append(env, corev1.EnvVar{Name: name, Value: val})
		}
	}
	return env
}

func removeEnvVars(env []corev1.EnvVar, toRemove []string) []corev1.EnvVar {
	for _, name := range toRemove {
		for i, envVar := range env {
			if envVar.Name == name {
				env = append(env[:i], env[i+1:]...)
				break
			}
		}
	}
	return env
}

type volumeSourceInfo struct {
	volumeSourceType VolumeSourceType
	volumeSourceName string
}

func parseVolumeSourceString(s string) (*volumeSourceInfo, error) {
	slices := strings.SplitN(s, ":", 2)
	if len(slices) != 2 {
		return nil, fmt.Errorf("Argument requires a value that contains the : character; got %q", s)
	}

	var volumeSourceType VolumeSourceType
	volumeSourceName := strings.TrimSpace(slices[1])
	if strings.TrimSpace(slices[0]) == "config-map" {
		volumeSourceType = ConfigMapVolumeSourceType
	} else if strings.TrimSpace(slices[0]) == "secret" {
		volumeSourceType = SecretVolumeSourceType
	} else {
		return nil, fmt.Errorf("Unsupported volume source type \"%q\"; supported volume source types are \"config-map\" and \"secret\"", slices[0])
	}

	return &volumeSourceInfo{
		volumeSourceType: volumeSourceType,
		volumeSourceName: volumeSourceName,
	}, nil
}

func updateEnvFrom(envFromSources []corev1.EnvFromSource, toUpdate []string) ([]corev1.EnvFromSource, error) {
	insertConfigMapSet := make(map[string]bool)
	insertSecretSet := make(map[string]bool)

	for _, s := range toUpdate {
		info, err := parseVolumeSourceString(s)
		if err != nil {
			return nil, err
		}
		switch info.volumeSourceType {
		case ConfigMapVolumeSourceType:
			insertConfigMapSet[info.volumeSourceName] = false
		case SecretVolumeSourceType:
			insertSecretSet[info.volumeSourceName] = false
		}
	}

	for i := range envFromSources {
		envSrc := &envFromSources[i]
		if envSrc.ConfigMapRef != nil {
			if _, ok := insertConfigMapSet[envSrc.ConfigMapRef.Name]; ok {
				insertConfigMapSet[envSrc.ConfigMapRef.Name] = true
			}
		}

		if envSrc.SecretRef != nil {
			if _, ok := insertSecretSet[envSrc.SecretRef.Name]; ok {
				insertSecretSet[envSrc.SecretRef.Name] = true
			}
		}
	}

	for k, v := range insertConfigMapSet {
		if !v {
			envFromSources = append(envFromSources, corev1.EnvFromSource{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: k,
					}}})
		}
	}

	for k, v := range insertSecretSet {
		if !v {
			envFromSources = append(envFromSources, corev1.EnvFromSource{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: k,
					}}})
		}
	}

	return envFromSources, nil
}

func removeEnvFrom(envFromSources []corev1.EnvFromSource, toRemove []string) ([]corev1.EnvFromSource, error) {
	for _, name := range toRemove {
		info, err := parseVolumeSourceString(name)
		if err != nil {
			return nil, err
		}
		for i, envSrc := range envFromSources {
			if (info.volumeSourceType == ConfigMapVolumeSourceType && envSrc.ConfigMapRef != nil && info.volumeSourceName == envSrc.ConfigMapRef.Name) ||
				(info.volumeSourceType == SecretVolumeSourceType && envSrc.SecretRef != nil && info.volumeSourceName == envSrc.SecretRef.Name) {
				envFromSources = append(envFromSources[:i], envFromSources[i+1:]...)
				break
			}
		}
	}
	return envFromSources, nil
}

func updateVolume(volume *corev1.Volume, info *volumeSourceInfo) error {
	switch info.volumeSourceType {
	case ConfigMapVolumeSourceType:
		volume.Secret = nil
		volume.ConfigMap = &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: info.volumeSourceName}}
	case SecretVolumeSourceType:
		volume.ConfigMap = nil
		volume.Secret = &corev1.SecretVolumeSource{SecretName: info.volumeSourceName}
	default:
		return fmt.Errorf("Invalid VolumeSourceType")
	}
	return nil
}

func existsVolumeNameInVolumes(volumeName string, volumes []corev1.Volume) bool {
	for _, volume := range volumes {
		if volume.Name == volumeName {
			return true
		}
	}
	return false
}

func existsVolumeNameInVolumeMounts(volumeName string, volumeMounts []corev1.VolumeMount) bool {
	for _, volumeMount := range volumeMounts {
		if volumeMount.Name == volumeName {
			return true
		}
	}
	return false
}

// updateVolumeMountsFromMap updates or adds volume mounts. If a given name of a volume is not existing, it returns an error
func updateVolumeMountsFromMap(volumeMounts []corev1.VolumeMount, toUpdate map[string]string, volumes []corev1.Volume) ([]corev1.VolumeMount, error) {
	set := make(map[string]bool)

	for i := range volumeMounts {
		volumeMount := &volumeMounts[i]
		name, present := toUpdate[volumeMount.MountPath]

		if present {
			if !existsVolumeNameInVolumes(name, volumes) {
				return nil, fmt.Errorf("There is no volume matched with %q", name)
			}

			volumeMount.ReadOnly = true
			volumeMount.Name = name
			set[volumeMount.MountPath] = true
		}
	}

	for mountPath, name := range toUpdate {
		if !set[mountPath] {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      name,
				ReadOnly:  true,
				MountPath: mountPath,
			})
		}
	}

	return volumeMounts, nil
}

func removeVolumeMounts(volumeMounts []corev1.VolumeMount, toRemove []string) []corev1.VolumeMount {
	for _, mountPath := range toRemove {
		for i, volumeMount := range volumeMounts {
			if volumeMount.MountPath == mountPath {
				volumeMounts = append(volumeMounts[:i], volumeMounts[i+1:]...)
				break
			}
		}
	}
	return volumeMounts
}

// updateVolumesFromMap updates or adds volumes regardless whether the volume is used or not
func updateVolumesFromMap(volumes []corev1.Volume, toUpdate map[string]*volumeSourceInfo) ([]corev1.Volume, error) {
	set := make(map[string]bool)

	for i := range volumes {
		volume := &volumes[i]
		info, present := toUpdate[volume.Name]
		if present {
			err := updateVolume(volume, info)
			if err != nil {
				return nil, err
			}
			set[volume.Name] = true
		}
	}

	for name, info := range toUpdate {
		if !set[name] {
			volumes = append(volumes, corev1.Volume{Name: name})
			updateVolume(&volumes[len(volumes)-1], info)
		}
	}

	return volumes, nil
}

// removeVolumes removes volumes. If there is a volume mount referencing the volume, it causes an error
func removeVolumes(volumes []corev1.Volume, toRemove []string, volumeMounts []corev1.VolumeMount) ([]corev1.Volume, error) {
	for _, name := range toRemove {
		for i, volume := range volumes {
			if volume.Name == name {
				if existsVolumeNameInVolumeMounts(name, volumeMounts) {
					return nil, fmt.Errorf("The volume %q cannot be removed because it is mounted", name)
				}
				volumes = append(volumes[:i], volumes[i+1:]...)
				break
			}
		}
	}
	return volumes, nil
}
